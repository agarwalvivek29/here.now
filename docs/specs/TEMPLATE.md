# Spec: [Feature Name]

> Copy this file to `docs/specs/[ISSUE-NUMBER]-[feature-name].md` and fill it in.
> The spec must be completed and linked in the GitHub Issue before implementation begins.

**Issue**: #[ISSUE-NUMBER]
**Status**: Draft | Review | Approved | Implemented
**Author**: [name]
**Date**: YYYY-MM-DD
**Services Affected**: [list services]

---

## Summary

[1-3 sentences describing what this feature does and why it exists.]

---

## Background and Motivation

[Why are we building this? What problem does it solve? What happens if we don't build it?
Include any relevant context, user research, or business requirements.]

---

## Scope

### In Scope

- [Specific thing 1 that will be built]
- [Specific thing 2]

### Out of Scope

- [Thing that might seem related but is NOT part of this feature]
- [Explicitly list deferred work to prevent scope creep]

---

## Acceptance Criteria

> Each criterion must be verifiable. Write them as testable statements.

- [ ] Given [context], when [action], then [expected outcome]
- [ ] Given [context], when [action], then [expected outcome]
- [ ] [Edge case handled]
- [ ] [Error case handled]

---

## Technical Design

### Architecture Overview

[High-level description of the approach. Include a diagram if helpful (ASCII or Mermaid).]

```
[Optional: ASCII diagram of data flow or component interaction]
```

### API Changes

#### New Endpoints

```
POST /v1/[resource]
Request: { ... }
Response: { ... }
```

#### Modified Endpoints

[List any changes to existing endpoints]

#### Removed Endpoints

[List any endpoints being removed and the migration path]

### Data Model Changes

#### New Tables / Collections

```sql
-- or show JSON schema for MongoDB
CREATE TABLE [name] (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ...
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

#### Modified Tables

[Show ALTER TABLE statements or schema diff]

### Queue / Event Changes

[New events published or consumed, with payload schema]

### Dependencies

[New packages, services, or infrastructure needed]

---

## Security Considerations

- [Authentication/authorization implications]
- [Data exposure risks]
- [Rate limiting needs]
- [Input validation requirements]

---

## Observability

- **Logs**: What should be logged? At what level?
- **Metrics**: Any new metrics to track? (e.g., queue depth, latency percentile)
- **Alerts**: Should any alert be created for failure conditions?

---

## Testing Plan

> Run `/plan-eng-review` before implementation begins. It generates a QA test plan consumed by `/qa`.
> After running it, paste the "Affected Pages/Routes" summary below and link the artifact path.

**gstack test plan**: `~/.gstack/projects/{slug}/{user}-{branch}-test-plan-{date}.md`

### Affected Pages / Routes

- [paste from /plan-eng-review output]

### Unit Tests

- [What functions/modules need unit tests]

### Integration Tests

- [What interactions need integration tests]

---

## Migration / Rollout Plan

- **Database migrations**: [yes/no — if yes, describe the migration]
- **Breaking changes**: [yes/no — if yes, describe backward compatibility strategy]
- **Feature flag**: [yes/no — if yes, describe the flag and rollout strategy]
- **Rollback plan**: [How to revert if something goes wrong]

---

## Open Questions

| Question                    | Owner  | Status |
| --------------------------- | ------ | ------ |
| [Question about approach X] | [name] | Open   |

---

## References

- Related ADR: [link if architectural decision was needed]
- Related issues: #[number]
- Related PRs: #[number]
- External docs: [link]
