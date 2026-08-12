package domain

import (
	"testing"

	herenowv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/herenow/v1"
)

// TestListVisibleTo checks that ListVisibleTo returns exactly the artifacts the
// identity may view: owners see their own PRIVATE docs, invited grantees see
// granted INVITED docs, ORG docs are visible to any authed user, and a nil
// identity sees nothing. Input order is preserved.
func TestListVisibleTo(t *testing.T) {
	const owner = "sub-owner"
	const viewer = "sub-viewer"

	own := &herenowv1.Artifact{Slug: "own", OwnerSub: owner, Visibility: herenowv1.Visibility_VISIBILITY_PRIVATE}
	invited := &herenowv1.Artifact{Slug: "invited", OwnerSub: owner, Visibility: herenowv1.Visibility_VISIBILITY_INVITED}
	org := &herenowv1.Artifact{Slug: "org", OwnerSub: owner, Visibility: herenowv1.Visibility_VISIBILITY_ORG}
	arts := []*herenowv1.Artifact{own, invited, org}

	grants := []*herenowv1.Grant{
		{Slug: "invited", GranteeSub: viewer},
		// A grant on a different slug must not leak visibility to "org"/"own".
		{Slug: "own", GranteeSub: viewer},
	}

	// Owner sees all three (owner always wins).
	if got := ListVisibleTo(&herenowv1.Identity{Sub: owner}, arts, grants); len(got) != 3 {
		t.Fatalf("owner expected 3 visible, got %d", len(got))
	}

	// Viewer: sees invited (via grant) and org (authed), but NOT own PRIVATE —
	// the stray "own" grant must not grant view on a PRIVATE artifact.
	got := ListVisibleTo(&herenowv1.Identity{Sub: viewer}, arts, grants)
	if len(got) != 2 {
		t.Fatalf("viewer expected 2 visible, got %d", len(got))
	}
	if got[0].GetSlug() != "invited" || got[1].GetSlug() != "org" {
		t.Fatalf("viewer visibility/order mismatch: got %q, %q", got[0].GetSlug(), got[1].GetSlug())
	}

	// Nil identity sees nothing (fail-closed), and the result is empty.
	if got := ListVisibleTo(nil, arts, grants); len(got) != 0 {
		t.Fatalf("nil identity expected 0 visible, got %d", len(got))
	}
}
