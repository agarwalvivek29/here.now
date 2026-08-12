# ArtifactA v2 — Build Log

> Durable record of the multi-agent build. **Target branch: `main`** (ADR-0003).
> PRD: `_bmad-output/planning-artifacts/prds/prd-ArtifactA-2026-08-09/prd.md`
> Task backlog: [tasks.md](tasks.md) · Dependency tree: [task-tree.md](task-tree.md)

## Orchestration model

- Each task = a short-lived **worktree branch off `main`** → **PR into `main`** → I review
  (FR acceptance criteria + tests green + clean/modular + concise comments) → **auto-merge**
  after review (per user, 2026-08-12).
- Dependent tasks re-branch off the **updated `main`** after their prerequisites merge.
- **Infra-gated** items (OIDC end-to-end vs a real IdP, TLS/VPN deploy, render-parity vs real
  Claude artifacts) are built with **mocks/fixtures + flagged** for later infra validation.
- Worktree agents symlink `node_modules` + `.husky/_` from the repo root so hooks run
  (commitlint requires lower-case subjects; never `--no-verify`).
- Code: clean, modular, concise comments. `main` must always build.

## Decisions (source of truth: PRD memlog + ADRs)

- Product **ArtifactA** = productization of the `here.now` wedge. P0: **any AI assistant can
  publish** (ADR-0012). Auth: OIDC browser SSO + CLI loopback-PKCE; assistant rides the CLI
  session (ADR-0007). Deploy: **behind VPN**; RBAC + audit retained (ADR-0011). Visibility:
  PRIVATE / INVITED / ORG (ORG = any employee). Storage: single-file bundle (ADR-0004),
  streaming Blob (ADR-0005), S3 backend-only later (ADR-0006), Postgres for shared-with-me
  (ADR-0009). Viewer: separate content origin (ADR-0008), render-parity harness (ADR-0010).
- Phasing: **P0 launch** (FR1–FR32); fast-follow (MCP, workspaces/teams, Postgres+S3, TTL);
  later (collaboration loop, Helm).

## Findings

- **F-1 (RESOLVED, PR #5)** — `CanView` grants were not slug-scoped. Fixed: `CanView` now
  matches grantee AND slug (defense-in-depth). Was never exploitable (server pre-filters).

## Open notes / carried debt

- **maxBytes** middleware (PR #8) is defined + tested but **not yet wired to a route** — no
  upload route exists until the publish API. Wire it there (FR1). Rate-limit IS live on
  `/a/{slug}` + `/a/{slug}/raw` (120/min per IP).

## Waves

### Wave 1 — zero-dependency P0 foundation — **MERGED ✓**

| Task                                     | FR             | PR  |
| ---------------------------------------- | -------------- | --- |
| W1-T1 `CanView` + `NewSlug` unit tests   | FR29           | #3  |
| W1-T2 streaming `Blob` (`io.ReadCloser`) | FR23           | #1  |
| W1-T3 atomic metadata writes             | FR24           | #2  |
| W1-T4 env-var config (`ARTIFACTA_`)      | FR25 (partial) | #4  |

### Wave 2 — hardening + dashboard/data foundation — **MERGED ✓**

| Task                                              | FR        | PR  |
| ------------------------------------------------- | --------- | --- |
| W2-T1 slug-scope `CanView` (fixes F-1)            | FR10/FR29 | #5  |
| W2-T2 `ListByGrantee` + `ListVisibleTo`           | FR18 (D1) | #6  |
| W2-T3 audit-integrity verifier (`audit verify`)   | FR21      | #7  |
| W2-T4 rate-limit content routes + body-size guard | FR31      | #8  |

Integration after each wave: `go build ./...` + `go test ./...` green (api/config/domain/infra).

### Wave 3 — infra-gated P0 core (candidate, pending sequencing) — **not started**

- Publish API `POST /artifacts` (FR1) + CLI publish-over-API + Skill (FR2/FR4) [mock auth].
- OIDC browser SSO scaffolding (FR6) + CLI loopback-PKCE (FR7) [mock issuer/JWKS].
- `share`/grants routes (FR12/13) → dashboard Mine/Shared/Org (FR18/19) → login page.
- Separate cookieless content origin wiring (FR16, ADR-0008).
- Render-parity harness (FR15, ADR-0010 — bundle-at-publish leading) [vs real artifacts].
- Deploy: docker-compose + TLS + real /metrics + structured logging (FR25–27).
- ARCHITECTURE.md constraint update for VPN (ADR-0011 action item).

## Log

- **2026-08-12** — Folded `astra-v2` into `main` (ADR-0003); wrote ADRs 0003–0012 + index.
- **2026-08-12** — **Wave 1** dispatched (4 worktree agents) → PRs #1–#4, reviewed +
  squash-merged, integration green. Finding F-1 recorded.
- **2026-08-12** — **Wave 2** dispatched → PRs #5–#8, reviewed + squash-merged, integration
  green. F-1 resolved (#5). Wave 3 (infra-gated core) planned; paused for sequencing.
