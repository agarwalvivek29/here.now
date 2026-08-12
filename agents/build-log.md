# ArtifactA v2 — Build Log

> Durable record of the multi-agent build. **Target branch: `main`** (ADR-0003).
> PRD: `_bmad-output/planning-artifacts/prds/prd-ArtifactA-2026-08-09/prd.md`
> Task backlog: [tasks.md](tasks.md) · Dependency tree: [task-tree.md](task-tree.md)

## Orchestration model

- Each task = a short-lived **worktree branch off `main`** → **PR into `main`** → I review
  (FR acceptance criteria + tests green + clean/modular + concise comments) → **auto-merge**
  after review (per user, 2026-08-12). Dependent tasks re-branch off the updated `main`.
- **Infra-gated** items built with **mocks/fixtures + flagged** for later infra validation.
- Worktree agents symlink `node_modules` + `.husky/_` so hooks run (commitlint = lower-case
  subjects; never `--no-verify`). `main` must always build.

## Decisions (source of truth: PRD memlog + ADRs)

- **ArtifactA** = productization of the `here.now` wedge. P0: **any AI assistant can publish**
  (ADR-0012). Auth: OIDC browser SSO + CLI loopback-PKCE; assistant rides the CLI session
  (ADR-0007). Deploy: **behind VPN**; RBAC + audit retained (ADR-0011). Visibility:
  PRIVATE / INVITED / ORG (ORG = any employee).
- **CLI→API auth (W3-T3):** CLI sends its OIDC `id_token` as `Bearer`, JWKS-verified
  server-side. Token in 0600 config file; OS-keychain = hardening follow-up (ADR-0007).
- Phasing: P0 launch; fast-follow (MCP, workspaces/teams, Postgres+S3, TTL); later
  (collaboration loop, Helm). **Deploy track deferred** per user.

## Findings

- **F-1 (RESOLVED, PR #5)** — `CanView` grants now match grantee AND slug (defense-in-depth).

## Waves

### Wave 1 — foundation — **MERGED ✓** (PRs #1–#4)

streaming Blob (#1), atomic writes (#2), CanView/NewSlug tests (#3), env config (#4).

### Wave 2 — hardening + data foundation — **MERGED ✓** (PRs #5–#8)

slug-scope CanView (#5), ListByGrantee/ListVisibleTo (#6), audit verifier (#7), rate-limit +
body-size guard (#8).

### Wave 3 — identity + publish core — **MERGED ✓** (PRs #9–#12)

publish Skill (#9), OIDC browser SSO (#10), publish API + maxBytes wired (#11), CLI loopback
login + publish-over-API Bearer (#12). Plus ARCHITECTURE.md VPN update (ADR-0011).

### Wave 4 — UI + RBAC — **MERGED ✓** (PRs #13–#14)

| Task                                            | FR      | PR  |
| ----------------------------------------------- | ------- | --- |
| share/grants + set-visibility + `herenow share` | FR12/13 | #13 |
| dashboard (Mine/Shared/Org) + sign-in page      | FR18/19 | #14 |

## Remaining P0 — both DESIGN-GATED (paused for user)

- **Content origin (FR16, ADR-0008)** — turned out non-trivial: serving `/raw` from a
  cookieless SEPARATE origin breaks cookie-based auth for private artifacts. Needs a decision:
  (A) signed per-view capability token, (B) parent-domain session cookie + distinct subdomain,
  or (C) defer (current null-origin sandboxed iframe is already decent). **Awaiting user call.**
- **Render-parity harness (FR15, ADR-0010)** — the make-or-break. Decision: bundle-at-publish
  (leading) vs pinned import-map+SRI; needs real sample artifacts to validate. **Awaiting call.**

## Also remaining (not design-gated)

- Broaden e2e (FR30) + security-gate pass (FR32, `/review` or `/cso`).
- Deploy (FR25–27) — deferred (needs infra approval).

## Log

- **2026-08-12** — Waves 1–2 (PRs #1–#8) merged green. F-1 recorded + resolved (#5).
- **2026-08-12** — Wave 3 (PRs #9–#12) merged green: identity + publish core. ARCHITECTURE.md
  VPN update. maxBytes debt cleared (#11). Each auth diff independently built + reviewed.
- **2026-08-12** — Wave 4 (PRs #13–#14) merged green: share/grants + set-visibility, dashboard
  - sign-in (html/template escaping). **Paused:** content-origin (FR16) revealed a cross-origin
    auth design decision; surfaced to user alongside render-parity (FR15).
