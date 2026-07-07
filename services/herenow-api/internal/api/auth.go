package api

import (
	"net/http"
	"strings"

	herenowv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/herenow/v1"
)

// Local is a single-user auth Provider for local / dev deploys. A bearer token
// or the hn_session cookie must equal the configured token to resolve the one
// configured identity. Real deploys use the OIDC / forward-auth adapters (see
// docs/adr/0002 for why here.now diverges from the template's JWT+API-key default).
type Local struct {
	Token string
	ID    *herenowv1.Identity
}

func (l *Local) Identify(r *http.Request) (*herenowv1.Identity, bool) {
	tok := bearer(r)
	if tok == "" {
		if c, err := r.Cookie("hn_session"); err == nil {
			tok = c.Value
		}
	}
	if l.Token != "" && tok == l.Token {
		return l.ID, true
	}
	return nil, false
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	return ""
}
