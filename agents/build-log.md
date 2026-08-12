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

- **F-1 (open) — `CanView` grants are not slug-scoped.** `domain.CanView` matches a grant on
  `grantee_sub` only, never comparing `grant.slug` to `artifact.slug`. **Not exploitable
  today** — the server passes `Store.Grants(slug)` (already slug-filtered) — but it is a
  missing defense-in-depth check. Documented by `TestCanView_notSlugScoped` (PR #3). Fixed in
  **W2-T1** (add slug check + flip the test).

## Waves

### Wave 1 — zero-dependency P0 foundation — **MERGED ✓**

| Task                                     | FR              | PR  | Status |
| ---------------------------------------- | --------------- | --- | ------ |
| W1-T1 `CanView` unit tests               | FR29            | #3  | merged |
| W1-T2 streaming `Blob` (`io.ReadCloser`) | FR23 (ADR-0005) | #1  | merged |
| W1-T3 atomic metadata writes             | FR24            | #2  | merged |
| W1-T4 env-var config (`ARTIFACTA_`)      | FR25 (partial)  | #4  | merged |

Integration: `go build ./...` + `go test ./...` green on merged `main` (config/domain/infra).

### Wave 2 — hardening + dashboard/data foundation (file-disjoint) — **in progress**

| Task                                           | FR        | Files                                                         | Infra-gated |
| ---------------------------------------------- | --------- | ------------------------------------------------------------- | ----------- |
| W2-T1 slug-scope `CanView` (fixes F-1)         | FR10/FR29 | `internal/domain/authz.go`, `authz_test.go`                   | no          |
| W2-T2 `ListByGrantee` + `ListVisibleTo`        | FR18 (D1) | `internal/infra/store.go`, `internal/domain/list.go` (new)    | no          |
| W2-T3 audit-integrity verifier CLI             | FR21      | `internal/infra/audit_verify.go` (new), `internal/cli/cli.go` | no          |
| W2-T4 upload size limit + rate-limit on `/raw` | FR31      | `internal/api/server.go` (+ new middleware)                   | no          |

### Later waves (indicative — composed off updated `main` as deps clear)

- OIDC browser SSO (FR6) + CLI loopback-PKCE (FR7) [infra-gated] → `share`/grants (FR12/13)
  → dashboard (FR18/19) → login page.
- Separate content origin (FR16) + render-parity harness (FR15) [infra-gated].
- Deploy: docker-compose + TLS + metrics/logging (FR25–27) [infra-gated].
- Publish API (FR1) + CLI (FR2) + Skill (FR4); security gate (FR30/32).

## Log

- **2026-08-12** — Reconciled branches: folded `astra-v2` into `main` (ADR-0003). Wrote ADRs
  0003–0012 + index.
- **2026-08-12** — **Wave 1 dispatched** (4 parallel worktree agents) → PRs #1–#4, all
  reviewed + squash-merged into `main`, integration green. Cleaned worktrees/branches.
  Recorded finding **F-1** (CanView slug-scoping). **Wave 2 dispatched** (W2-T1..T4).
