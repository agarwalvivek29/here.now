package domain

import (
	"strings"
	"testing"

	herenowv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/herenow/v1"
)

// artifact is a terse constructor for the table cases below.
func artifact(owner string, vis herenowv1.Visibility) *herenowv1.Artifact {
	return &herenowv1.Artifact{Slug: "doc", OwnerSub: owner, Visibility: vis}
}

// grant to a grantee on a slug. CanView keys only on grantee (see slug note).
func grant(slug, grantee string) *herenowv1.Grant {
	return &herenowv1.Grant{Slug: slug, GranteeSub: grantee}
}

func TestCanView(t *testing.T) {
	const owner = "sub-owner"
	const viewer = "sub-viewer"

	tests := []struct {
		name   string
		art    *herenowv1.Artifact
		who    *herenowv1.Identity
		grants []*herenowv1.Grant
		want   bool
	}{
		// Fail-closed: no authenticated subject can ever view.
		{"nil identity denied", artifact(owner, herenowv1.Visibility_VISIBILITY_ORG), nil, nil, false},
		{"empty sub denied", artifact(owner, herenowv1.Visibility_VISIBILITY_ORG), &herenowv1.Identity{Sub: ""}, nil, false},
		{"org nil identity denied", artifact(owner, herenowv1.Visibility_VISIBILITY_ORG), nil, nil, false},
		{"org empty sub denied", artifact(owner, herenowv1.Visibility_VISIBILITY_ORG), &herenowv1.Identity{Sub: ""}, nil, false},

		// Owner always wins, regardless of visibility.
		{"owner sees private", artifact(owner, herenowv1.Visibility_VISIBILITY_PRIVATE), &herenowv1.Identity{Sub: owner}, nil, true},
		{"owner sees invited", artifact(owner, herenowv1.Visibility_VISIBILITY_INVITED), &herenowv1.Identity{Sub: owner}, nil, true},
		{"owner sees org", artifact(owner, herenowv1.Visibility_VISIBILITY_ORG), &herenowv1.Identity{Sub: owner}, nil, true},
		{"owner sees unspecified", artifact(owner, herenowv1.Visibility_VISIBILITY_UNSPECIFIED), &herenowv1.Identity{Sub: owner}, nil, true},

		// PRIVATE: only the owner, never a non-owner.
		{"private non-owner denied", artifact(owner, herenowv1.Visibility_VISIBILITY_PRIVATE), &herenowv1.Identity{Sub: viewer}, nil, false},
		{"private non-owner with stray grant denied", artifact(owner, herenowv1.Visibility_VISIBILITY_PRIVATE), &herenowv1.Identity{Sub: viewer}, []*herenowv1.Grant{grant("doc", viewer)}, false},

		// INVITED: non-owner needs a grant whose grantee matches.
		{"invited matching grant allowed", artifact(owner, herenowv1.Visibility_VISIBILITY_INVITED), &herenowv1.Identity{Sub: viewer}, []*herenowv1.Grant{grant("doc", viewer)}, true},
		{"invited no grants denied", artifact(owner, herenowv1.Visibility_VISIBILITY_INVITED), &herenowv1.Identity{Sub: viewer}, nil, false},
		{"invited grant for other grantee denied", artifact(owner, herenowv1.Visibility_VISIBILITY_INVITED), &herenowv1.Identity{Sub: viewer}, []*herenowv1.Grant{grant("doc", "sub-someone-else")}, false},
		{"invited matching grantee among many allowed", artifact(owner, herenowv1.Visibility_VISIBILITY_INVITED), &herenowv1.Identity{Sub: viewer}, []*herenowv1.Grant{grant("doc", "sub-a"), grant("doc", viewer), grant("doc", "sub-b")}, true},

		// ORG: any authenticated non-owner is allowed.
		{"org non-owner allowed", artifact(owner, herenowv1.Visibility_VISIBILITY_ORG), &herenowv1.Identity{Sub: viewer}, nil, true},

		// UNSPECIFIED: closed to non-owners (no case in the switch).
		{"unspecified non-owner denied", artifact(owner, herenowv1.Visibility_VISIBILITY_UNSPECIFIED), &herenowv1.Identity{Sub: viewer}, nil, false},
		{"unspecified non-owner with grant denied", artifact(owner, herenowv1.Visibility_VISIBILITY_UNSPECIFIED), &herenowv1.Identity{Sub: viewer}, []*herenowv1.Grant{grant("doc", viewer)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanView(tt.art, tt.who, tt.grants); got != tt.want {
				t.Errorf("CanView() = %v, want %v", got, tt.want)
			}
		})
	}
}

// CanView does NOT scope grants by slug — it matches on grantee only, trusting
// the caller to pass grants already filtered to the artifact's slug. This test
// pins that current behavior: a grant for a DIFFERENT slug but the right grantee
// still allows. Flagged for FR29 — if per-call slug scoping is intended, CanView
// must also compare g.GetSlug() to a.GetSlug() and this expectation flips.
func TestCanView_notSlugScoped(t *testing.T) {
	art := artifact("sub-owner", herenowv1.Visibility_VISIBILITY_INVITED)
	who := &herenowv1.Identity{Sub: "sub-viewer"}
	grants := []*herenowv1.Grant{grant("a-different-slug", "sub-viewer")}
	if !CanView(art, who, grants) {
		t.Errorf("CanView() = false; expected true — CanView matches grantee only, not slug")
	}
}

func TestNewSlug(t *testing.T) {
	s := NewSlug()

	if s == "" {
		t.Fatal("NewSlug() returned empty string")
	}
	if s != strings.ToLower(s) {
		t.Errorf("NewSlug() = %q, expected all lower-case", s)
	}

	// URL-safe base32 (no padding), lower-cased: a-z and 2-7 only.
	const allowed = "abcdefghijklmnopqrstuvwxyz234567"
	for _, r := range s {
		if !strings.ContainsRune(allowed, r) {
			t.Errorf("NewSlug() = %q contains disallowed rune %q", s, r)
		}
	}

	// 16 random bytes → 26 base32 chars with no padding.
	if len(s) != 26 {
		t.Errorf("NewSlug() length = %d, want 26", len(s))
	}
}

func TestNewSlug_unique(t *testing.T) {
	if NewSlug() == NewSlug() {
		t.Error("two successive NewSlug() calls returned the same value")
	}
}
