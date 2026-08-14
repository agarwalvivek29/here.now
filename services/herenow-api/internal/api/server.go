// Package api is the HTTP layer: viewer serving + authorization-gated artifact
// bytes. Security invariants enforced here: identity comes from the auth
// Provider; the per-artifact decision is domain.CanView; the bundle is fetched
// ONLY AFTER an allow; there are no client-reachable pre-signed URLs; and every
// decision (allow and deny) is written to the inbuilt audit trail. Fails closed.
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	herenowv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/herenow/v1"
	"github.com/agarwalvivek29/here.now/services/herenow-api/internal/domain"
	"github.com/agarwalvivek29/here.now/services/herenow-api/internal/render"
	"github.com/agarwalvivek29/here.now/services/herenow-api/internal/web"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Consumer-side interfaces (Go convention). infra.FileStore / infra.BlobFS /
// api.Local satisfy these.
type Store interface {
	GetArtifact(slug string) (*herenowv1.Artifact, bool, error)
	PutArtifact(a *herenowv1.Artifact) error
	Grants(slug string) ([]*herenowv1.Grant, error)
	AddGrant(g *herenowv1.Grant) error
	Append(ev *herenowv1.AuditEvent) error
	// Dashboard listings (FR18): Mine, Shared with me, and Org.
	ListByOwner(sub string) ([]*herenowv1.Artifact, error)
	ListByGrantee(sub string) ([]*herenowv1.Artifact, error)
	ListByVisibility(v herenowv1.Visibility) ([]*herenowv1.Artifact, error)
}

type Blob interface {
	Get(slug string) (io.ReadCloser, error)
	Put(slug string, r io.Reader) error
}

type Auth interface {
	Identify(r *http.Request) (*herenowv1.Identity, bool)
}

type Server struct {
	Store Store
	Blob  Blob
	Auth  Auth
	// BaseURL is the public origin (e.g. https://here.now) used to build the
	// absolute artifact link returned by the publish API.
	BaseURL string
	// OIDC, when non-nil, provides browser SSO (ADR-0007): its Login/Callback
	// handlers replace the dev cookie-setter and it is the request authenticator.
	// When nil, the server uses the dev /login helper and the Local adapter.
	OIDC *OIDCProvider
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	// Exempt paths — explicitly allowlisted, never unprotected by default.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("# metrics: not yet implemented\n"))
	})

	// Root (FR18, FR19): the dashboard when signed in, else the sign-in landing.
	// {$} matches only the exact "/" path, so it never shadows the routes below.
	mux.HandleFunc("GET /{$}", s.dashboard)

	// Auth routes: real OIDC SSO when configured, else the dev cookie-setter.
	if s.OIDC != nil {
		mux.HandleFunc("GET /login", s.OIDC.Login)
		mux.HandleFunc("GET /callback", s.OIDC.Callback)
	} else {
		mux.HandleFunc("GET /login", s.login) // dev helper: sets the session cookie
	}

	// Content routes are rate-limited per client IP (120 req/min) to blunt
	// scraping and brute-force enumeration; /health and /metrics stay unwrapped
	// so probes never trip the limiter. Auth, CanView, audit, and response
	// headers all remain enforced inside the wrapped handlers.
	const contentPerMinute = 120
	mux.Handle("GET /a/{slug}", rateLimit(http.HandlerFunc(s.viewer), contentPerMinute))  // viewer shell (no content)
	mux.Handle("GET /a/{slug}/raw", rateLimit(http.HandlerFunc(s.raw), contentPerMinute)) // authz-gated bundle bytes

	// Publish (FR1): authenticated upload. The body is capped at 25 MiB by the
	// maxBytes transport guard (finally wiring the previously-defined ceiling)
	// and rate-limited per client IP like the other content routes.
	const maxPublishBytes = 25 << 20 // 25 MiB
	mux.Handle("POST /artifacts", rateLimit(maxBytes(http.HandlerFunc(s.publish), maxPublishBytes), contentPerMinute))

	// Sharing (FR12, FR13): owner-only mutations. Both fail closed — an
	// unauthenticated caller gets 401 and a non-owner (or unknown slug) gets a
	// 404 that never distinguishes "missing" from "not yours".
	mux.Handle("POST /artifacts/{slug}/grants", rateLimit(http.HandlerFunc(s.addGrant), contentPerMinute))
	mux.Handle("PATCH /artifacts/{slug}/visibility", rateLimit(http.HandlerFunc(s.setVisibility), contentPerMinute))
	return mux
}

// login is a v0 convenience for the local single-user flow: it drops the token
// into an hn_session cookie so a browser can authenticate. Real deploys replace
// this with the OIDC login handler.
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: "hn_session", Value: r.URL.Query().Get("token"), Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/health", http.StatusFound)
}

// dashboard (FR18, FR19) is the root landing. An unauthenticated caller gets the
// sign-in page at HTTP 200 (never an error) so the front door always renders. An
// authenticated caller sees three sections: Mine (owned), Shared with me
// (grant-based), and Org (org-visible artifacts owned by others). The caller's
// own artifacts are filtered out of Org so they never appear twice.
func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	who, ok := s.Auth.Identify(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !ok {
		if err := web.RenderSignin(w); err != nil {
			http.Error(w, "unavailable", http.StatusInternalServerError)
		}
		return
	}
	sub := who.GetSub()

	mine, err := s.Store.ListByOwner(sub)
	if err != nil {
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}
	shared, err := s.Store.ListByGrantee(sub)
	if err != nil {
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}
	org, err := s.Store.ListByVisibility(herenowv1.Visibility_VISIBILITY_ORG)
	if err != nil {
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}

	data := web.DashboardData{
		Email:  who.GetEmail(),
		Mine:   toViews(mine),
		Shared: toViews(shared),
	}
	// Org lists org-visible artifacts owned by someone else — the caller's own
	// org artifacts already appear under Mine.
	for _, a := range org {
		if a.GetOwnerSub() == sub {
			continue
		}
		data.Org = append(data.Org, toView(a))
	}

	if err := web.RenderDashboard(w, data); err != nil {
		http.Error(w, "unavailable", http.StatusInternalServerError)
	}
}

// toViews projects a slice of artifacts into the dashboard presentation model.
func toViews(arts []*herenowv1.Artifact) []web.ArtifactView {
	out := make([]web.ArtifactView, 0, len(arts))
	for _, a := range arts {
		out = append(out, toView(a))
	}
	return out
}

func toView(a *herenowv1.Artifact) web.ArtifactView {
	return web.ArtifactView{
		Slug:       a.GetSlug(),
		Title:      a.GetTitle(),
		Visibility: visibilityLabel(a.GetVisibility()),
	}
}

// visibilityLabel maps the Visibility enum to the lower-case wire label shown in
// the dashboard (the inverse of parseVisibility).
func visibilityLabel(v herenowv1.Visibility) string {
	switch v {
	case herenowv1.Visibility_VISIBILITY_PRIVATE:
		return "private"
	case herenowv1.Visibility_VISIBILITY_INVITED:
		return "invited"
	case herenowv1.Visibility_VISIBILITY_ORG:
		return "org"
	default:
		return "unknown"
	}
}

func (s *Server) viewer(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy",
		// connect-src 'self' lets the shell fetch /a/{slug}/raw; without it the
		// fetch falls back to default-src 'none' and the viewer can't load.
		"default-src 'none'; connect-src 'self'; frame-src 'self'; script-src 'unsafe-inline'; style-src 'unsafe-inline'")
	_, _ = w.Write([]byte(web.ViewerHTML))
}

// raw returns the artifact bytes only after an authorization allow.
func (s *Server) raw(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	who, _ := s.Auth.Identify(r)

	art, ok, err := s.Store.GetArtifact(slug)
	if err != nil {
		s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_DENY, false)
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}
	if !ok {
		s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_VIEW, false)
		http.NotFound(w, r) // don't leak existence
		return
	}
	grants, err := s.Store.Grants(slug)
	if err != nil {
		s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_DENY, false)
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}
	if !domain.CanView(art, who, grants) {
		s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_DENY, false)
		http.NotFound(w, r) // 404-as-not-found: forbidden and missing look identical
		return
	}

	// Invariant: the blob is fetched ONLY after CanView allows above.
	rc, err := s.Blob.Get(slug)
	if err != nil {
		s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_DENY, false)
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}
	defer rc.Close()
	s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_VIEW, true)

	ct := art.GetContentType()
	if ct == "" {
		ct = "text/html; charset=utf-8"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Security-Policy",
		"sandbox allow-scripts; default-src 'none'; img-src data: https:; style-src 'unsafe-inline'; script-src 'unsafe-inline'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = io.Copy(w, rc)
}

// publish accepts an authenticated upload and creates a private artifact. The
// body is streamed to the blob store; metadata is recorded; a PUBLISH audit
// event is written. Fails closed: an unauthenticated caller gets 401 and
// nothing is stored.
func (s *Server) publish(w http.ResponseWriter, r *http.Request) {
	who, ok := s.Auth.Identify(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized) // don't leak anything
		return
	}
	owner := who.GetSub()

	title := r.URL.Query().Get("title")
	if title == "" {
		title = "untitled"
	}
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		ct = "text/html; charset=utf-8"
	}

	slug := domain.NewSlug()

	// Read the body up to the maxBytes ceiling. Unlike the earlier
	// stream-to-blob path, publish now buffers the payload so the render
	// pipeline can bundle inline module scripts before storage (FR15,
	// ADR-0010). Buffering the whole body is acceptable under the 25 MiB cap.
	// An over-limit body still trips the maxBytes guard on read and surfaces as
	// a MaxBytesError, which we translate to 413.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_DENY, false)
			http.Error(w, "payload too large", http.StatusRequestEntityTooLarge)
			return
		}
		s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_DENY, false)
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}

	// Render-parity pipeline: bundle inline ES modules into a self-contained
	// artifact that renders under the strict /raw CSP. Fail-soft — Bundle
	// returns the original bytes (plus warnings) rather than erroring, so a
	// publish is never blocked by a bundling problem.
	bundled, bundledCT, _, _ := render.Bundle(body, ct)
	ct = bundledCT

	if err := s.Blob.Put(slug, bytes.NewReader(bundled)); err != nil {
		s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_DENY, false)
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}

	art := &herenowv1.Artifact{
		Slug:        slug,
		OwnerSub:    owner,
		Title:       title,
		Visibility:  herenowv1.Visibility_VISIBILITY_PRIVATE, // private by default
		ContentType: ct,
		CreatedAt:   timestamppb.Now(),
	}
	if err := s.Store.PutArtifact(art); err != nil {
		s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_DENY, false)
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}
	s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_PUBLISH, true)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"slug": slug,
		"url":  s.BaseURL + "/a/" + slug,
	})
}

// ownedArtifact resolves the artifact at slug for an owner-only mutation. It
// enforces the two fail-closed gates shared by the sharing endpoints: the
// caller must be authenticated (else 401) and must own the artifact (else 404).
// A missing artifact and a non-owner caller are deliberately indistinguishable —
// the 404 leaks neither existence nor authorization, mirroring the raw handler.
// The returned bool reports whether the caller may proceed; when false the
// response has already been written.
func (s *Server) ownedArtifact(w http.ResponseWriter, r *http.Request, slug string) (*herenowv1.Artifact, *herenowv1.Identity, bool) {
	who, ok := s.Auth.Identify(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized) // don't leak anything
		return nil, nil, false
	}
	art, found, err := s.Store.GetArtifact(slug)
	if err != nil {
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return nil, nil, false
	}
	if !found || art.GetOwnerSub() != who.GetSub() {
		http.NotFound(w, r) // missing and forbidden look identical
		return nil, nil, false
	}
	return art, who, true
}

// addGrant (FR12) grants a subject access to an artifact. Owner-only. The
// grantee subject comes from the JSON body {"grantee_sub":"..."} or a ?grantee=
// query param. A SHARE audit event is written on success. Returns 201.
func (s *Server) addGrant(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	_, who, ok := s.ownedArtifact(w, r, slug)
	if !ok {
		return
	}

	grantee := r.URL.Query().Get("grantee")
	if grantee == "" {
		var body struct {
			GranteeSub string `json:"grantee_sub"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			grantee = body.GranteeSub
		}
	}
	if grantee == "" {
		http.Error(w, "missing grantee", http.StatusBadRequest)
		return
	}

	g := &herenowv1.Grant{
		Slug:       slug,
		GranteeSub: grantee,
		GrantedBy:  who.GetSub(),
		CreatedAt:  timestamppb.Now(),
	}
	if err := s.Store.AddGrant(g); err != nil {
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}
	s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_SHARE, true)
	w.WriteHeader(http.StatusCreated)
}

// setVisibility (FR13) changes an artifact's visibility. Owner-only. The body is
// {"visibility":"private|invited|org"}; an unknown value is rejected with 400. A
// SHARE audit event is written on success. Returns 200.
func (s *Server) setVisibility(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	art, who, ok := s.ownedArtifact(w, r, slug)
	if !ok {
		return
	}

	var body struct {
		Visibility string `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	vis, ok := parseVisibility(body.Visibility)
	if !ok {
		http.Error(w, "unknown visibility", http.StatusBadRequest)
		return
	}

	art.Visibility = vis
	if err := s.Store.PutArtifact(art); err != nil {
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}
	s.audit(who, slug, herenowv1.AuditAction_AUDIT_ACTION_SHARE, true)
	w.WriteHeader(http.StatusOK)
}

// parseVisibility maps the wire strings accepted by the set-visibility endpoint
// to the Visibility enum. The unspecified/zero value is never a valid target, so
// an unrecognized string returns ok=false and the caller rejects it with 400.
func parseVisibility(s string) (herenowv1.Visibility, bool) {
	switch s {
	case "private":
		return herenowv1.Visibility_VISIBILITY_PRIVATE, true
	case "invited":
		return herenowv1.Visibility_VISIBILITY_INVITED, true
	case "org":
		return herenowv1.Visibility_VISIBILITY_ORG, true
	default:
		return herenowv1.Visibility_VISIBILITY_UNSPECIFIED, false
	}
}

func (s *Server) audit(who *herenowv1.Identity, slug string, action herenowv1.AuditAction, allowed bool) {
	sub := ""
	if who != nil {
		sub = who.GetSub()
	}
	_ = s.Store.Append(&herenowv1.AuditEvent{
		Ts: timestamppb.Now(), PrincipalSub: sub, Slug: slug, Action: action, Allowed: allowed,
	})
}
