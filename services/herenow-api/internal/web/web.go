// Package web embeds the viewer into the binary so a single artifact serves
// everything. v0 ships a minimal sandboxed-iframe loader; the forked
// artifact-runtime (apps/viewer) builds into embed/ in a later phase.
package web

import _ "embed"

//go:embed embed/viewer.html
var ViewerHTML string
