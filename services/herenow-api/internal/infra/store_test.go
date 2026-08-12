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
