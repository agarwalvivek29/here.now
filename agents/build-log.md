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
- Code: clean, modular, concise comments explaining blocks. `main` must always build.

## Decisions (source of truth: PRD memlog + ADRs)

- Product: **ArtifactA** = productization of the `here.now` wedge. P0 north star: **any AI
  assistant can publish** (ADR-0012).
- Auth: OIDC browser SSO + CLI loopback-PKCE; assistant rides the CLI session (ADR-0007).
- Deploy: **behind VPN**; per-artifact RBAC + audit retained as defense-in-depth (ADR-0011).
- Visibility: PRIVATE / INVITED / ORG (ORG = any authenticated employee).
- Storage: single-file bundle now (ADR-0004); streaming Blob (ADR-0005); S3 backend-only
  later (ADR-0006); Postgres for shared-with-me at scale (ADR-0009).
- Viewer: separate cookieless content origin (ADR-0008); render-parity harness (ADR-0010).
- Phasing: **P0 launch** (FR1–FR32); fast-follow (MCP, workspaces/teams, Postgres+S3, TTL);
  later (collaboration loop, Helm).

## Waves

### Wave 1 — zero-dependency P0 foundation (file-disjoint, fully in-repo)

| Task                                     | FR              | Files                                                                     | Infra-gated | Status  | PR  |
| ---------------------------------------- | --------------- | ------------------------------------------------------------------------- | ----------- | ------- | --- |
| W1-T1 `CanView` unit tests               | FR29            | `internal/domain/authz_test.go` (new)                                     | no          | pending | —   |
| W1-T2 streaming `Blob` (`io.ReadCloser`) | FR23 (ADR-0005) | `internal/infra/blob.go`, `internal/api/server.go`, `internal/cli/cli.go` | no          | pending | —   |
| W1-T3 atomic metadata writes             | FR24            | `internal/infra/store.go`                                                 | no          | pending | —   |
| W1-T4 env config scaffolding             | FR25 (partial)  | `internal/config/config.go`                                               | no          | pending | —   |

### Later waves (indicative — composed off updated `main` as deps clear)

- Audit-integrity verifier CLI (FR21); upload size limit + rate-limit (FR31).
- `ListByGrantee` + `ListVisibleTo` (D1/FR18) → dashboard (FR18/19) → login page (FR19).
- OIDC browser SSO (FR6) + CLI loopback-PKCE (FR7) [infra-gated] → `share`/grants (FR12/13).
- Separate content origin (FR16) + render-parity harness (FR15) [infra-gated].
- Deploy: docker-compose + TLS + metrics/logging (FR25–27) [infra-gated].
- Publish API (FR1) + CLI (FR2) + Skill (FR4); security gate (FR30/32).

## Log

- **2026-08-12** — Reconciled branches: folded `astra-v2` into `main` (build target = `main`,
  ADR-0003). Wrote ADRs 0003–0012, updated ADR index. Committing ADRs + this build-log to
  `main`. Next: launch Wave 1.
