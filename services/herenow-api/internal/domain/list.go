package domain

import (
	herenowv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/herenow/v1"
)

// ListVisibleTo returns the subset of arts that who may view, reusing CanView
// for the per-artifact decision. For each artifact it passes only the grants
// whose slug matches that artifact, so cross-slug grants never leak. Input
// order is preserved; the result is nil when nothing is visible.
func ListVisibleTo(who *herenowv1.Identity, arts []*herenowv1.Artifact, grants []*herenowv1.Grant) []*herenowv1.Artifact {
	var out []*herenowv1.Artifact
	for _, a := range arts {
		var scoped []*herenowv1.Grant
		for _, g := range grants {
			if g.GetSlug() == a.GetSlug() {
				scoped = append(scoped, g)
			}
		}
		if CanView(a, who, scoped) {
			out = append(out, a)
		}
	}
	return out
}
