// Package api is the HTTP layer: viewer serving + authorization-gated artifact
// bytes. Security invariants enforced here: identity comes from the auth
// Provider; the per-artifact decision is domain.CanView; the bundle is fetched
// ONLY AFTER an allow; there are no client-reachable pre-signed URLs; and every
// decision (allow and deny) is written to the inbuilt audit trail. Fails closed.
package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	herenowv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/herenow/v1"
	"github.com/agarwalvivek29/here.now/services/herenow-api/internal/domain"
	"github.com/agarwalvivek29/here.now/services/herenow-api/internal/web"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Consumer-side interfaces (Go convention). infra.FileStore / infra.BlobFS /
// api.Local satisfy these.
type Store interface {
	GetArtifact(slug string) (*herenowv1.Artifact, bool, error)
	PutArtifact(a *herenowv1.Artifact) error
	Grants(slug string) ([]*herenowv1.Grant, error)
	Append(ev *herenowv1.AuditEvent) error
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

func (s *Server) viewer(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy",
		"default-src 'none'; frame-src 'self'; script-src 'unsafe-inline'; style-src 'unsafe-inline'")
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

	// Stream the request body straight to the blob store (no full-payload
	// buffering). An over-limit body trips the maxBytes guard here and surfaces
	// as a MaxBytesError, which we translate to 413.
	if err := s.Blob.Put(slug, r.Body); err != nil {
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

func (s *Server) audit(who *herenowv1.Identity, slug string, action herenowv1.AuditAction, allowed bool) {
	sub := ""
	if who != nil {
		sub = who.GetSub()
	}
	_ = s.Store.Append(&herenowv1.AuditEvent{
		Ts: timestamppb.Now(), PrincipalSub: sub, Slug: slug, Action: action, Allowed: allowed,
	})
}
