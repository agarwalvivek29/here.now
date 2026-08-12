# 0003 — Branch strategy: build ArtifactA v2 on `main` (trunk-based)

**Date**: 2026-08-12
**Status**: Accepted
**Deciders**: Vivek Agarwal
**Issue**: N/A

---

## Context

ArtifactA v2 is a large, multi-track build. It was briefly isolated on a long-lived
`astra-v2` branch to protect a "current working" `main`. But `main` today is only the
**skeleton wedge** — not a deployed product — so there is nothing live to protect, and
`CLAUDE.md` already mandates a **trunk-based-on-`main`** workflow. A long-lived divergent
branch adds merge risk and ceremony for no pre-launch benefit.

---

## Decision

Build v2 **directly on `main`**, trunk-based.

- Each task is a **short-lived worktree branch off `main`** → PR into `main` → review →
  auto-merge after review. Dependent tasks re-branch off the updated `main`.
- Keep commits small and green; **`main` must always build**.
- The `astra-v2` branch was created and pushed, then **abandoned**; its two commits
  (roadmap, gitignore) were fast-forwarded into `main`.
- A capability that would break the shippable wedge while half-built goes behind a flag.

This supersedes the initial `astra-v2`-isolation decision.

---

## Consequences

### Positive

- No branch drift; continuous integration; aligns with `CLAUDE.md` trunk-based policy.
- Every task is reviewed and lands green, so `main` stays coherent.

### Negative

- `main` tolerates in-progress (not-yet-launched) v2 capabilities — mitigated by small green
  PRs, flags, and the test/security gate (FR29–FR32) before any promotion/deploy.

### Neutral

- Requires discipline: PRs are small, file-scoped, and independently reviewable.

---

## Alternatives Considered

### Long-lived `astra-v2` isolation branch

Rejected: unnecessary pre-launch (nothing deployed to protect) and accrues merge risk;
contradicts the repo's trunk-based policy.

### Separate repository for v2

Rejected: loses shared history, schema, and tooling; promotion becomes a manual re-import.
