# 0016 — Comment threads (replies)

**Date**: 2026-08-25
**Status**: Accepted
**Deciders**: Vivek Agarwal
**Issue**: N/A
**Relates to**: [0014](./0014-artifact-comments.md) (which deferred threading),
[0015](./0015-anchored-comments.md) (a thread's root may be anchored)

---

## Context

Comments (ADR-0014) shipped flat and explicitly deferred replies. But a lone note isn't a
conversation — a reviewer raises a point, the author answers, the reviewer confirms. Each comment
needs to become a thread so that back-and-forth stays together instead of scattering as sibling
notes.

---

## Decision

A reply is just a comment with a `parent_id` — reuse the whole model rather than add a new type.

- **Schema**: `Comment.parent_id` (string). Empty = a **root** (a thread); set = a **reply** in
  that thread.
- **One level only.** A reply's parent must be a root. Replying to a reply is rejected (400).
  Flat-within-thread keeps rendering and reasoning simple; deep nesting adds no real value here.
- **Replies inherit the root.** A reply takes the root's `version` (a thread lives on one
  version) and is **never anchored** (the root carries the anchor for the whole thread).
- **Who can reply**: anyone who can _view_ the artifact — same gate as commenting (ADR-0014). No
  new authorization.
- **Resolve is thread-level.** Only a root can be resolved (owner-only, unchanged). Resolving a
  reply id — or any unknown id — is a not-leaking 404.
- **Endpoint**: the existing `POST /artifacts/{slug}/comments` takes an optional `parent_id`; no
  new route. The list endpoint returns roots and replies flat (each with its `parent_id`) and the
  viewer groups them into threads. Count in the bar is **threads** (roots), not total comments.

---

## Consequences

### Positive

- Conversations hold together; the review loop (raise → answer → confirm) reads naturally.
- Zero new type, store, endpoint, or auth path — replies ride the comment machinery.

### Negative

- One-level threading can't express sub-threads; acceptable for review conversations.

### Neutral

- Replies are user content → rendered via textContent like all comment bodies (never innerHTML).

## Alternatives Considered

### A nested `replies` array on `Comment`

Rejected: a flat list with `parent_id` reuses the existing create/list/store paths unchanged and
groups trivially in the viewer; a nested array would need bespoke storage and serialization.

### Arbitrary-depth threads

Rejected: review conversations are shallow; unbounded nesting complicates layout and adds no
value. One level is enough now and can be revisited.
