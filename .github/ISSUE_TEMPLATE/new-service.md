---
name: New Service
about: Request or plan the creation of a new microservice
title: "feat: new service — [service-name]"
labels: new-service, needs-spec, needs-adr
assignees: ""
---

## Service Name

`[service-name]` (use kebab-case; this becomes the directory name)

## Purpose

<!-- What does this service do? What problem does it solve? Why does it need to be a separate service? -->

## Language & Type

- **Language**: TypeScript / Python / Go / Rust
- **Service type**: REST API / GraphQL API / WebSocket / gRPC / Worker / CLI / Agentic
- **Runtime**: Node.js / Bun / Python / Go binary / Rust binary

## Interfaces

### Exposes

<!-- APIs, events published, or queue topics produced -->

-

### Consumes

<!-- APIs it calls, events consumed, queue topics subscribed -->

-

## Data Requirements

- [ ] PostgreSQL
- [ ] MongoDB
- [ ] Redis (cache/session)
- [ ] Redis (queue)
- [ ] Kafka
- [ ] RabbitMQ / Celery
- [ ] None

## ADR

<!-- A new service almost always requires an ADR (why a new service, not extending an existing one) -->

**ADR**: `docs/adr/[NNNN]-[title].md` (REQUIRED — create before scaffolding)

## Spec

**Spec**: `docs/specs/[ISSUE-NUMBER]-[service-name].md` (REQUIRED — create before scaffolding)

## Checklist

Before running `scripts/new-service.sh`:

- [ ] ADR created and accepted
- [ ] Spec created and approved
- [ ] Service name agreed upon
- [ ] Language and framework decided
- [ ] Interfaces with other services documented above
- [ ] Team member assigned as service owner

## Additional Context

<!-- Architecture diagrams, external API docs, related issues -->
