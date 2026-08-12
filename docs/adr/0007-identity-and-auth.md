# 0007 — Identity & auth: OIDC browser SSO + CLI loopback-PKCE + assistant-rides-session

**Date**: 2026-08-12
**Status**: Accepted
**Deciders**: Vivek Agarwal
**Issue**: N/A
**Refines**: [0002](./0002-auth-model.md)

---

## Context

ArtifactA's P0 is "any AI assistant can publish", recipients authenticate via company SSO,
and the instance runs behind a VPN. We need both **end-user** auth (browser) and
**programmatic** auth (CLI/assistant) against the org's OIDC IdP. The CLI is a **public
client** (cannot hold a secret — OAuth 2.1). Headless assistants have no browser of their own.

---

## Decision

- **Browser SSO** — OIDC **Authorization-Code + PKCE**. `/login` → IdP → `/callback` →
  session cookie. `id_token` verified against the issuer JWKS; identity = `sub`.
- **CLI login** — OIDC **Authorization-Code + PKCE via loopback** (opens the IdP in a
  browser, callback on `127.0.0.1`). `client_id` + issuer provided at install (optionally
  baked into a per-org CLI build). Tokens stored in the **OS keychain**.
- **Assistant acts as the user** by **riding the CLI's stored session** — no separate
  headless login. The human authenticates interactively once at install; agents publish
  thereafter through the CLI/MCP/Skill.
- Grants bind to the immutable `sub`. The `local` single-token adapter is retained for
  zero-dependency/dev deploys.

Loopback-PKCE is chosen over the device flow for the CLI: the human-at-install has a
browser, and loopback is the smoother UX. Device flow can be added later for pure-headless
installs.

---

## Consequences

### Positive

- One identity across CLI and browser (same `sub`) — a CLI-published artifact is owned by
  the same person who logs into the dashboard.
- Headless agents need no login of their own; the "any assistant" P0 is met without secrets
  in distributed binaries.
- Standards-based; works against any OIDC issuer.

### Negative

- The first CLI login needs a same-machine browser (loopback). Documented; device-flow is
  the escape hatch for pure-headless installs.

### Neutral

- The CLI becomes an HTTP API client — it stops touching the store directly.

---

## Alternatives Considered

### Device Authorization Grant (RFC 8628) for the CLI

Deferred: better for browserless devices, but the install-time human has a browser and
loopback is simpler. Add later if needed.

### Static API keys

Rejected: extractable from distributed binaries and carry no per-user identity.

### Shared JWT / API-key middleware

Rejected per ADR 0002 — cannot express per-artifact RBAC or end-user identity.
