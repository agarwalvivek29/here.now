# 0008 — Serve artifact bytes from a separate, cookieless content origin

**Date**: 2026-08-12
**Status**: Accepted
**Deciders**: Vivek Agarwal
**Issue**: N/A

---

## Context

The viewer renders **untrusted** artifact HTML/JS. The app's session cookie (`hn_session`)
is scoped to the app origin. If artifact content is served from the **same origin** as the
app, a sandbox escape could read that cookie or reach same-origin app routes. The current v0
viewer relies only on an `allow-scripts` (null-origin) iframe — decent, but the bytes still
transit the app origin.

---

## Decision

Serve artifact bytes (`/raw`, and future multi-file sub-paths) from a **distinct, cookieless
content origin** (a separate host/subdomain).

- The **app origin** serves the viewer shell, dashboard, and auth — and carries the session
  cookie.
- The **content origin** serves only artifact bytes, carries **no** auth cookies, and is
  framed inside a sandboxed iframe (`allow-scripts`, no `allow-same-origin`), with a strict
  CSP and `X-Content-Type-Options: nosniff`.
- Authorization is still enforced **server-side** (`CanView`) before any byte is served
  (ADR invariant); the separate origin is defense-in-depth, not the access control.

---

## Consequences

### Positive

- Session-cookie isolation: artifact JS can never read `hn_session` even on a sandbox escape.
- Matches the strongest pattern in the space (artifact.cafe's separate content origin).

### Negative

- Extra host / DNS / TLS wiring in the deploy story. Documented in the ops setup.

### Neutral

- The viewer fetches bytes cross-origin (uncredentialed); no cookies are needed there because
  authorization already happened on the app origin's gated route.

---

## Alternatives Considered

### Same-origin `srcdoc` sandbox only (v0 behavior)

Rejected as the endpoint: weaker isolation; bytes and cookies share an origin.

### `data:` URL rendering

Rejected: breaks relative asset references needed by future multi-file bundles.
