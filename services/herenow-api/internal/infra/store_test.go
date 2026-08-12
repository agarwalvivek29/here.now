package infra

import (
	"testing"

	herenowv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/herenow/v1"
)

// TestFileStoreRoundTrip verifies that metadata persisted by one FileStore is
// reloaded correctly by a fresh FileStore over the same directory — i.e. the
// atomic writeJSON path produces a readable file.
func TestFileStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()

	s, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	art := &herenowv1.Artifact{
		Slug:        "hello-world",
		OwnerSub:    "user-123",
		Title:       "Hello World",
		Visibility:  herenowv1.Visibility_VISIBILITY_PRIVATE,
		ContentType: "text/html",
	}
	if err := s.PutArtifact(art); err != nil {
		t.Fatalf("PutArtifact: %v", err)
	}

	grant := &herenowv1.Grant{
		Slug:       "hello-world",
		GranteeSub: "user-456",
		GrantedBy:  "user-123",
	}
	if err := s.AddGrant(grant); err != nil {
		t.Fatalf("AddGrant: %v", err)
	}

	// Fresh store over the same dir must reload both records from disk.
	reloaded, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("reload NewFileStore: %v", err)
	}

	got, ok, err := reloaded.GetArtifact("hello-world")
	if err != nil {
		t.Fatalf("GetArtifact: %v", err)
	}
	if !ok {
		t.Fatal("artifact not found after reload")
	}
	if got.GetOwnerSub() != "user-123" || got.GetTitle() != "Hello World" ||
		got.GetVisibility() != herenowv1.Visibility_VISIBILITY_PRIVATE ||
		got.GetContentType() != "text/html" {
		t.Fatalf("artifact round-trip mismatch: %+v", got)
	}

	grants, err := reloaded.Grants("hello-world")
	if err != nil {
		t.Fatalf("Grants: %v", err)
	}
	if len(grants) != 1 {
		t.Fatalf("expected 1 grant after reload, got %d", len(grants))
	}
	if grants[0].GetGranteeSub() != "user-456" || grants[0].GetGrantedBy() != "user-123" {
		t.Fatalf("grant round-trip mismatch: %+v", grants[0])
	}
}

// TestListByGrantee verifies the grants→artifacts join: a grantee sees a
// granted artifact, a non-grantee sees nothing, and multiple grants for the
// same slug dedupe to a single artifact.
func TestListByGrantee(t *testing.T) {
	s, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	art := &herenowv1.Artifact{
		Slug:       "shared-doc",
		OwnerSub:   "owner-1",
		Visibility: herenowv1.Visibility_VISIBILITY_INVITED,
	}
	if err := s.PutArtifact(art); err != nil {
		t.Fatalf("PutArtifact: %v", err)
	}
	// Two grants to the same grantee on the same slug — must dedupe.
	for i := 0; i < 2; i++ {
		if err := s.AddGrant(&herenowv1.Grant{Slug: "shared-doc", GranteeSub: "grantee-1", GrantedBy: "owner-1"}); err != nil {
			t.Fatalf("AddGrant: %v", err)
		}
	}
	// A dangling grant to a nonexistent artifact must be skipped.
	if err := s.AddGrant(&herenowv1.Grant{Slug: "ghost", GranteeSub: "grantee-1", GrantedBy: "owner-1"}); err != nil {
		t.Fatalf("AddGrant: %v", err)
	}

	got, err := s.ListByGrantee("grantee-1")
	if err != nil {
		t.Fatalf("ListByGrantee: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("grantee expected 1 deduped artifact, got %d", len(got))
	}
	if got[0].GetSlug() != "shared-doc" {
		t.Fatalf("expected shared-doc, got %q", got[0].GetSlug())
	}

	// A user with no grants sees nothing.
	none, err := s.ListByGrantee("stranger")
	if err != nil {
		t.Fatalf("ListByGrantee stranger: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("non-grantee expected 0 artifacts, got %d", len(none))
	}
}

// TestListByVisibility verifies the Org-tab query returns only artifacts whose
// visibility matches, regardless of owner.
func TestListByVisibility(t *testing.T) {
	s, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	seed := []*herenowv1.Artifact{
		{Slug: "org-a", OwnerSub: "owner-1", Visibility: herenowv1.Visibility_VISIBILITY_ORG},
		{Slug: "org-b", OwnerSub: "owner-2", Visibility: herenowv1.Visibility_VISIBILITY_ORG},
		{Slug: "priv", OwnerSub: "owner-1", Visibility: herenowv1.Visibility_VISIBILITY_PRIVATE},
		{Slug: "inv", OwnerSub: "owner-1", Visibility: herenowv1.Visibility_VISIBILITY_INVITED},
	}
	for _, a := range seed {
		if err := s.PutArtifact(a); err != nil {
			t.Fatalf("PutArtifact(%q): %v", a.GetSlug(), err)
		}
	}

	org, err := s.ListByVisibility(herenowv1.Visibility_VISIBILITY_ORG)
	if err != nil {
		t.Fatalf("ListByVisibility: %v", err)
	}
	if len(org) != 2 {
		t.Fatalf("expected 2 org artifacts, got %d", len(org))
	}
	for _, a := range org {
		if a.GetVisibility() != herenowv1.Visibility_VISIBILITY_ORG {
			t.Fatalf("non-org artifact leaked: %q (%v)", a.GetSlug(), a.GetVisibility())
		}
	}

	priv, err := s.ListByVisibility(herenowv1.Visibility_VISIBILITY_PRIVATE)
	if err != nil {
		t.Fatalf("ListByVisibility private: %v", err)
	}
	if len(priv) != 1 || priv[0].GetSlug() != "priv" {
		t.Fatalf("expected only 'priv' as private, got %v", priv)
	}
}
