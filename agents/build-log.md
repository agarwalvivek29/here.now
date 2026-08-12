# ArtifactA v2 — Build Log

> Durable record of the multi-agent build. **Target branch: `main`** (ADR-0003).
> PRD: `_bmad-output/planning-artifacts/prds/prd-ArtifactA-2026-08-09/prd.md`
> Task backlog: [tasks.md](tasks.md) · Dependency tree: [task-tree.md](task-tree.md)

## Orchestration model

- Each task = a short-lived **worktree branch off `main`** → **PR into `main`** → I review
  (FR acceptance criteria + tests green + clean/modular + concise comments) → **auto-merge**
  after review (per user, 2026-08-12). Dependent tasks re-branch off the updated `main`.
- **Infra-gated** items (OIDC end-to-end vs a real IdP, TLS/VPN deploy, render-parity vs real
  Claude artifacts) are built with **mocks/fixtures + flagged** for later infra validation.
- Worktree agents symlink `node_modules` + `.husky/_` so hooks run (commitlint = lower-case
  subjects; never `--no-verify`). `main` must always build.

## Decisions (source of truth: PRD memlog + ADRs)

- **ArtifactA** = productization of the `here.now` wedge. P0: **any AI assistant can publish**
  (ADR-0012). Auth: OIDC browser SSO + CLI loopback-PKCE; assistant rides the CLI session
  (ADR-0007). Deploy: **behind VPN**; RBAC + audit retained (ADR-0011). Visibility:
  PRIVATE / INVITED / ORG (ORG = any employee). Storage: single-file bundle (ADR-0004),
  streaming Blob (ADR-0005), S3 backend-only later (ADR-0006), Postgres later (ADR-0009).
  Viewer: separate content origin (ADR-0008), render-parity harness (ADR-0010).
- **CLI→API auth (decided W3-T3):** the CLI sends its OIDC `id_token` as a `Bearer`, verified
  server-side via the same JWKS (self-contained single-app). Token stored in the 0600 config
  file for now; **OS-keychain is a hardening follow-up** (ADR-0007).
- Phasing: P0 launch; fast-follow (MCP, workspaces/teams, Postgres+S3, TTL); later
  (collaboration loop, Helm). **Deploy track deferred** per user (2026-08-12).

## Findings

- **F-1 (RESOLVED, PR #5)** — `CanView` grants weren't slug-scoped; now matches grantee AND
  slug (defense-in-depth). Was never exploitable (server pre-filters).

## Waves

### Wave 1 — foundation — **MERGED ✓**

| Task                      | FR   | PR  |
| ------------------------- | ---- | --- |
| streaming `Blob`          | FR23 | #1  |
| atomic metadata writes    | FR24 | #2  |
| `CanView`/`NewSlug` tests | FR29 | #3  |
| env-var config            | FR25 | #4  |

### Wave 2 — hardening + data foundation — **MERGED ✓**

| Task                            | FR      | PR  |
| ------------------------------- | ------- | --- |
| slug-scope `CanView` (F-1)      | FR10/29 | #5  |
| `ListByGrantee`/`ListVisibleTo` | FR18    | #6  |
| audit-integrity verifier        | FR21    | #7  |
| rate-limit + body-size guard    | FR31    | #8  |

### Wave 3 — identity + publish core (infra-gated) — **in progress**

| Task                                             | FR      | PR  | Status   |
| ------------------------------------------------ | ------- | --- | -------- |
| `artifacta-publish` Skill                        | FR4     | #9  | merged ✓ |
| OIDC browser SSO (mock issuer)                   | FR6     | #10 | merged ✓ |
| publish API `POST /artifacts` (+ maxBytes wired) | FR1     | #11 | merged ✓ |
| CLI loopback login + publish-over-API (Bearer)   | FR2/FR7 | —   | building |

Also merged: ARCHITECTURE.md VPN constraint update (ADR-0011 action item).

### Remaining P0 (later waves)

- `share`/grants routes (FR12/13) → dashboard Mine/Shared/Org (FR18/19) → login page.
- Separate cookieless content origin wiring (FR16, ADR-0008).
- Render-parity harness (FR15, ADR-0010 — bundle-at-publish) [vs real artifacts].
- Deploy: docker-compose + TLS + real /metrics + logging (FR25–27) — **deferred (needs
  infra approval)**.
- Security gate: broaden e2e (FR30), `/review` or `/cso` pass (FR32).

## Log

- **2026-08-12** — Folded `astra-v2` → `main` (ADR-0003); wrote ADRs 0003–0012.
- **2026-08-12** — Wave 1 → PRs #1–#4 merged (green). F-1 recorded.
- **2026-08-12** — Wave 2 → PRs #5–#8 merged (green). F-1 resolved (#5).
- **2026-08-12** — ARCHITECTURE.md VPN update (a7eae4c). Wave 3: Skill #9, OIDC SSO #10,
  publish API #11 merged (each reviewed incl. independent build/test; auth code diff-reviewed
  for fail-closed + JWKS verify + empty-secret guard). maxBytes debt cleared (#11). CLI-client
  (W3-T3) dispatched with id_token-as-Bearer design.
