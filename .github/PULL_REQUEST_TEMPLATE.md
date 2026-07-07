## Summary

<!-- 1-3 sentences describing what this PR does. -->

## Related Issue

Closes #[ISSUE-NUMBER]

## Spec

<!-- Link to the spec file. Required for feature PRs. -->

**Spec**: `docs/specs/[ISSUE-NUMBER]-[feature-name].md`

## ADR

<!-- Link to ADR if an architectural decision was made. -->

**ADR**: `docs/adr/[NNNN]-[title].md` (N/A if no architectural decision)

---

## Checklist

### Before Submitting

- [ ] Spec file exists and is linked above
- [ ] ADR created and linked (if architectural decision made)
- [ ] All commits follow `type(scope): description` format
- [ ] Branch name follows `feat/[ISSUE-NUMBER]-[desc]` or `fix/[ISSUE-NUMBER]-[desc]`

### Security

- [ ] All new routes are protected by auth middleware (no unauthenticated endpoints added without approval)
- [ ] No tokens, secrets, or API keys logged or returned in response bodies
- [ ] `JWT_SECRET` and `API_KEY` sourced from env — not hardcoded anywhere

### Schema

- [ ] New data types defined in `packages/schema/proto/` — NOT in service code
- [ ] `packages/schema/generated/` regenerated and committed if proto changed
- [ ] No service-local type/interface/struct/class/enum for business domain concepts

### Tests (backend services only)

- [ ] Unit tests written for new domain functions (`tests/unit/`)
- [ ] E2E tests written for new API endpoints / queue handlers (`tests/e2e/`)
- [ ] Test coverage not reduced

### Code Quality

- [ ] No `any` types (TypeScript), no bare `except` (Python), no `unwrap()` (Rust)
- [ ] No hardcoded secrets or credentials
- [ ] No `console.log` / `print()` / `fmt.Println` left in production code

### Service Changes

- [ ] `AGENTS.md` updated if service behavior or architecture changed
- [ ] `.env.example` updated if new environment variables added
- [ ] Migration files created for database schema changes
- [ ] `docker compose up` still works after changes

### For New Services

- [ ] Created using `scripts/new-service.sh` (not manually)
- [ ] `AGENTS.md` populated with service context
- [ ] `Dockerfile` builds successfully
- [ ] Service stub added to `infra/docker-compose.yml`
- [ ] `README.md` documents local dev instructions

---

## Testing Instructions

<!-- How should reviewers test this? Step-by-step. -->

1.
2.

## Screenshots / Recordings

<!-- For UI changes or complex flows. Optional. -->

## Notes for Reviewer

<!-- Anything the reviewer should pay special attention to, known limitations, or follow-up work. -->
