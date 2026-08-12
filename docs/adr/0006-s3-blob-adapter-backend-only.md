# 0006 — S3-compatible blob adapter, backend-only (no presigned URLs)

**Date**: 2026-08-09
**Status**: Proposed
**Deciders**: Vivek Agarwal
**Issue**: N/A

---

## Context

`ARCHITECTURE.md` commits to "Filesystem (v0) → S3-compatible" for blob storage. The
`Blob` interface (ADR 0005) makes the adapter mechanically simple: one object per slug.

But here.now's security model forbids the reflexive S3 pattern. `ARCHITECTURE.md` states
"no client-reachable pre-signed URLs" and "bytes served ONLY after an authorization
allow." A presigned URL handed to the browser would let it fetch bytes directly from S3,
**bypassing `domain.CanView` and the audit trail** — breaking the two invariants the
product's compliance claim rests on.

---

## Decision

We will add an **S3-compatible blob adapter used strictly on the backend**:

- The server performs `GetObject`/`PutObject` with its own credentials (key = `slug`),
  runs `CanView`, appends the audit event, then **streams bytes through itself**.
- **No presigned URLs** are ever issued to clients. The browser never talks to S3.
- The bucket is private (Block Public Access on, no public policy); access is the
  server's IAM identity only.
- The adapter targets the S3 API generically (AWS SDK v2, configurable endpoint) so
  AWS S3, MinIO, Cloudflare R2, Backblaze B2, and Wasabi all work. Self-host default:
  **MinIO** in docker-compose.
- Encryption: SSE at rest minimum; **app-side envelope encryption before `Put`** is the
  strong option so an untrusted S3 endpoint never sees plaintext.

---

## Consequences

### Positive

- Preserves per-view authorization + audit unconditionally (every byte is gated/logged).
- Operators pick any S3-compatible backend; local self-host stays single-container.
- App-side encryption reinforces "never retrievable from third-party infra".

### Negative

- Streaming through the app costs an extra hop + egress vs. direct-from-edge — accepted,
  because the audit/authz invariant makes the hop mandatory and artifacts are small.

### Neutral

- Requires the streaming `Blob` interface (ADR 0005) to avoid buffering whole objects.
- Keyed one object per slug, consistent with the single-bundle format (ADR 0004).

---

## Alternatives Considered

### Presigned URLs (direct client → S3)

The standard S3 offload pattern. Rejected outright: bypasses `CanView` and audit,
violating `ARCHITECTURE.md` constraints and the product's core claim.

### CloudFront/CDN in front of the bucket

Rejected for the same reason at v2: a cacheable edge URL is a client-reachable path to
bytes that skips per-view authorization. Revisit only with signed, per-view, audited
edge auth — out of scope.
