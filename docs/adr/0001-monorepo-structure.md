# 0001 — Monorepo with Per-Service Isolation

**Date**: 2026-03-06
**Status**: Accepted
**Deciders**: Project founders
**Issue**: N/A (template bootstrap)

---

## Context

We are building a platform with multiple frontend apps and backend microservices spanning several languages (TypeScript, Python, Go, Rust). Services need to share types, utilities, and proto definitions, while remaining independently deployable.

Key constraints:

- Team works across all services; we want a unified developer experience
- Services are independently deployable to separate cloud environments
- We use multiple languages — the structure must accommodate all of them
- AI agents collaborate on this codebase and need clear, scoped context for each service

The main alternatives were: separate repositories per service (polyrepo) vs a single repository containing all services (monorepo).

---

## Decision

We will use a **monorepo** with strict **per-service isolation** boundaries.

Specifically:

- All services live under `services/[name]/`
- All frontend apps live under `apps/[name]/`
- Shared code lives only in `packages/[name]/` — never copied between services
- Each service is self-contained: its own `Dockerfile`, `.env.example`, and `AGENTS.md`
- Services communicate only via explicit interfaces (HTTP, gRPC, queues) — never via shared in-process state

---

## Consequences

### Positive

- Single place for shared types (`packages/`) eliminates drift
- Unified CI/CD pipeline with per-service build targets
- Atomic commits that span multiple services (e.g., API contract change + consumer update)
- AI agents have clear scope: read the service's `AGENTS.md` and work within the service boundary
- One set of conventions, enforced by shared tooling

### Negative

- Repository grows large over time; requires thoughtful CI caching
- `packages/` changes require coordination across all consumers
- Developers need discipline to not reach across service boundaries

### Neutral

- CI must be smart about which services need rebuilding on a given change (use path filters)
- Each language may need its own root tooling config (e.g., `Cargo.toml` workspace, `pnpm-workspace.yaml`, `go.work`)

---

## Alternatives Considered

### Polyrepo (separate repo per service)

Services live in separate repositories. Rejected because:

- Shared type changes require PRs across multiple repos
- Harder to keep conventions consistent
- More friction for AI agents operating across service boundaries
- CI/CD setup duplicated per repo

### Monorepo with shared in-process code

Services share a common codebase module and are deployed as one binary. Rejected because:

- Defeats independent deployability
- A bug in one service can affect all others
- Does not scale to multiple languages
