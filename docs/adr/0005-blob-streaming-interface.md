# 0005 — Blob store interface is streaming (io.ReadCloser)

**Date**: 2026-08-09
**Status**: Proposed
**Deciders**: Vivek Agarwal
**Issue**: N/A

---

## Context

The blob store interface currently reads a whole artifact into memory:

```go
type Blob interface { Get(slug string) ([]byte, error) }
```

`GET /a/{slug}/raw` therefore buffers the entire artifact per request. With a future S3
adapter and larger artifacts, N concurrent viewers of an M-byte artifact cost N×M of
heap. This is cheapest to fix now, while the filesystem store is the only adapter and
the only call site.

---

## Decision

We will change the blob contract to stream:

```go
type Blob interface {
    Get(slug string) (io.ReadCloser, error) // caller closes
    Put(slug string, r io.Reader) error
}
```

The HTTP layer `io.Copy`s the reader to the response after the `CanView` allow; the
filesystem adapter returns the opened `*os.File`; the future S3 adapter returns the
`GetObject` body. Buffering, if ever needed, becomes the caller's explicit choice.

---

## Consequences

### Positive

- Constant memory per request regardless of artifact size; no OOM under concurrency.
- The S3 adapter (ADR 0006) drops in without buffering whole objects.

### Negative

- Callers must close the reader (defer `Close()`); a leak is now possible if forgotten.

### Neutral

- Content-Type still comes from `Artifact` metadata, not the stream.
- Ordering unchanged: the reader is obtained only **after** `CanView` allows.

---

## Alternatives Considered

### Keep `[]byte`

Rejected: forces full buffering everywhere; the S3 adapter would have to read entire
objects into memory, exactly the failure mode this avoids.

### Return `io.ReadSeekCloser` + use `http.ServeContent`

Deferred: gives Range/If-Modified handling but the fs `*os.File` already satisfies it,
and S3 bodies do not seek without extra requests. Revisit if Range support is needed.
