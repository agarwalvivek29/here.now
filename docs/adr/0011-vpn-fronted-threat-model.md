# 0011 — VPN-fronted deployment threat model

**Date**: 2026-08-12
**Status**: Accepted
**Deciders**: Vivek Agarwal
**Issue**: N/A
**Amends**: [0002](./0002-auth-model.md), `ARCHITECTURE.md`

---

## Context

`ARCHITECTURE.md` and `PRODUCT.md` originally assumed **"no VPN — app auth + RBAC must be
safe standalone on the public internet."** For ArtifactA the deployment assumption changed:
the instance runs **inside the corporate network, reachable only over the VPN**, not exposed
to the public internet.

---

## Decision

Adopt a **VPN-fronted** threat model: the VPN is the **network perimeter**.

- The instance is **not** exposed to the public internet; recipients must be on the corporate
  VPN and authenticate via company SSO (OIDC).
- **Per-artifact RBAC (`CanView`) and the tamper-evident audit trail are retained** as
  defense-in-depth and for compliance (least privilege + who-viewed-what) — the VPN does
  **not** replace application-level authorization.
- TLS is still terminated (in-transit PII protection).
- Public-internet hardening (aggressive rate-limiting, existence-not-leaked) drops from "sole
  boundary" to "good practice" — but fail-closed and 404-not-403 behavior are kept as cheap
  defense-in-depth.

**ACTION:** update `ARCHITECTURE.md` "Architectural Constraints" and ADR-0002 context to
reference this ADR.

---

## Consequences

### Positive

- Realistic for the target buyer (enterprise/fintech behind a corporate network).
- Reduces the blast radius of any app-level bug (attacker must already be on the VPN).

### Negative

- External/non-account sharing is out (already a non-goal); recipients must be on the VPN.

### Neutral

- App RBAC + audit remain the product's core value regardless of the network perimeter.

---

## Alternatives Considered

### Public-internet-safe (original assumption)

Not wrong, but no longer the deployment reality; retained hardening as defense-in-depth only.

### Rely on VPN + SSO alone (drop per-artifact RBAC)

Rejected: destroys the compliance/least-privilege value that is the product's reason to exist.
