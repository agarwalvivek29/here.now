# 0009 — Postgres store adapter for grants & "shared with me"

**Date**: 2026-08-12
**Status**: Proposed
**Deciders**: Vivek Agarwal
**Issue**: N/A

---

## Context

The v0 metadata store is an in-memory + whole-file-rewrite `FileStore` (single process). The
dashboard's **"Shared with me"** tab (FR18) requires a reverse grant lookup
(`grantee_sub → artifacts`) — a many-to-many join. At team scale this is an indexed query,
not an in-memory scan, and concurrent multi-process access needs real transactions.

---

## Decision

Add a **PostgreSQL store adapter** behind the existing `Store` interface (artifacts, grants,
audit), selectable by config. Fast-follow (not P0).

- Tables: `artifacts`, `grants` (indexed on `grantee_sub` and `slug`), `audit_events`
  (append-only, indexed on `seq`).
- The append-only, hash-chained audit contract (ADR baseline) is preserved: rows are
  insert-only; the chain is computed as today.
- The `FileStore` remains the zero-dependency default for local/self-host-small.

---

## Consequences

### Positive

- Indexed "shared with me" / "org" queries; safe concurrent multi-process access.
- Enables horizontal scaling of the API tier (shared DB state).

### Negative

- A dependency to run and back up. Kept optional — `FileStore` stays the small-deploy default.

### Neutral

- Requires the `Store` interface to stay adapter-agnostic (no file-specific assumptions leak
  into domain/API).

---

## Alternatives Considered

### SQLite

Rejected for multi-process/HA: single-writer; fine for single-node but doesn't unlock the
horizontal scaling Postgres does.

### Keep `FileStore` only

Rejected at scale: whole-file rewrites and in-memory scans don't hold up for large grant sets
or concurrent writers.
