# 0010 — Viewer render-parity harness

**Date**: 2026-08-12
**Status**: Proposed
**Deciders**: Vivek Agarwal
**Issue**: N/A

---

## Context

Render parity is the adoption make-or-break (PRODUCT.md, FR15/NFR2). The v0 viewer injects
bytes into an iframe `srcdoc` — fine for self-contained HTML, but the artifacts assistants
actually emit are **React + an ESM import-map + Tailwind**, pulling dependencies from a CDN.
Under the sandbox CSP those external fetches are blocked, so such artifacts render blank.

---

## Decision

Build a **dependency-resolution harness** so the viewer renders modern component artifacts
with vendor-parity fidelity. Two candidate mechanisms, to be chosen during implementation
against real sample artifacts:

- **(A) Bundle-at-publish** — resolve/bundle dependencies into a self-contained bundle when
  the artifact is published; the viewer serves static, dependency-free output. Best fidelity
  - offline/VPN-friendly (no CDN egress), at the cost of a publish-time build step.
- **(B) Pinned import-map + SRI** — serve a locked import-map with subresource-integrity
  hashes from an operator-controlled, VPN-reachable asset origin. Lighter publish path;
  requires hosting the pinned dependency set.

Given the **VPN-behind** deployment (no public CDN egress assumed), **(A) bundle-at-publish
is the leading option**; (B) is the fallback if bundling proves too heavy. Rendering stays
inside the sandboxed, separate content origin (ADR 0008).

---

## Consequences

### Positive

- Modern React/Tailwind artifacts render like the vendor viewer — the core adoption bet.
- Bundling keeps rendering fully inside operator infra (no third-party CDN at view time).

### Negative

- A build/resolution step (publish-time for A, curation for B) — the most complex P0 item.

### Neutral

- The chosen mechanism is a serve/publish detail behind the artifact contract (ADR 0004);
  the store/authz/audit model is unaffected.

---

## Alternatives Considered

### Keep raw `srcdoc` (v0)

Rejected: renders only self-contained HTML; fails the make-or-break for component artifacts.

### Allow live CDN fetches from the sandbox

Rejected: breaks the VPN/no-third-party-egress posture and weakens the sandbox CSP.
