# 0012 — Assistant-agnostic publish surfaces (API + CLI + MCP + Skill)

**Date**: 2026-08-12
**Status**: Accepted
**Deciders**: Vivek Agarwal
**Issue**: N/A

---

## Context

ArtifactA's P0 north star: **any** AI assistant — not just Claude — must be able to publish
an artifact. Assistants differ widely in what they can call (shell, HTTP, MCP, skills), so a
single integration surface would exclude some. The publish path must also preserve the auth
model (ADR 0007: the human authenticates the CLI once; agents ride that session).

---

## Decision

Expose publish through **layered surfaces that all wrap one publish API**:

- **REST API** (`POST /artifacts`) — the canonical contract; the lowest common denominator any
  agent with HTTP can call.
- **CLI** (`artifacta publish`) — wraps the API; holds the OAuth session (ADR 0007); the
  universal surface for shell-capable agents.
- **MCP server** — an ergonomic tool surface for MCP-speaking assistants (fast-follow).
- **Skill** — a Claude Code / agent skill wrapping the CLI/API for skill-capable assistants.

All surfaces converge on the same API + auth + authz + audit path — no surface bypasses
`CanView` or the audit trail.

---

## Consequences

### Positive

- Meets "any assistant can publish" without bespoke per-assistant work: HTTP/CLI is universal,
  MCP/Skill are ergonomic layers.
- One code path to secure and audit (surfaces are thin wrappers).

### Negative

- Multiple surfaces to maintain and document — mitigated by keeping them thin over one API.

### Neutral

- P0 ships API + CLI + Skill; MCP is fast-follow. All bind identity via ADR 0007.

---

## Alternatives Considered

### MCP-only

Rejected: excludes assistants that don't speak MCP; not "any assistant".

### CLI-only

Rejected: awkward for HTTP-native or MCP-native agents; the API must be first-class.
