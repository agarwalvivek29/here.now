# CLAUDE.md — Agent Contract

> This file governs how Claude Code operates in this repository and all projects derived from it.
> It also applies to all other AI agents (Copilot, Codex, Gemini, etc.) — `AGENTS.md` is an alias to this file.
> Read this before taking **any** action.

---

## Project Context

Read these before every session:

- **[ARCHITECTURE.md](ARCHITECTURE.md)** — system overview, service map, data flow, tech stack, constraints
- **[PRODUCT.md](PRODUCT.md)** — product vision, target users, core features, roadmap

> **Missing file check**: If either `ARCHITECTURE.md` or `PRODUCT.md` does not exist, **stop immediately**.
> Tell the human: "Bootstrap was not completed — `ARCHITECTURE.md` and/or `PRODUCT.md` are missing.
> Please re-run the bootstrap process (restore `BOOTSTRAP.md` from git and follow it) before starting any work."
> Do not proceed with any task until both files exist.

---

## Quick Start

| Situation                         | Action                                | gstack skill                      |
| --------------------------------- | ------------------------------------- | --------------------------------- |
| Evaluating a new idea             | Read BOOTSTRAP.md Phase 2             | `/office-hours`                   |
| Writing a plan (>2 files)         | Spec Gate + Plan Gate → EnterPlanMode | `/plan-eng-review` before exiting |
| Architectural decision            | ADR Gate — create ADR first           | —                                 |
| Pre-landing code check            | Run before pushing                    | `/review`                         |
| Creating a PR                     | Full PR workflow                      | `/ship`                           |
| CI check failing                  | Fix root cause — never `--no-verify`  | `/debug`                          |
| Stuck / blocked                   | Diagnose root cause                   | `/debug`                          |
| New service                       | `scripts/new-service.sh`              | —                                 |
| Feature added / changed / removed | See CORE_RULES Rule 15                | —                                 |
| Unclear requirements              | `AskUserQuestion` — never assume      | —                                 |
| Potential secret in code          | Flag it, do not commit                | —                                 |

---

## Mandatory Pre-Flight Checks

Before starting any task, you MUST:

1. **Read `docs/CORE_RULES.md`** — The binding rules for this repository. No exceptions.
2. **Find the GitHub Issue** — Task context, requirements, and acceptance criteria come from the linked issue.
3. **Check for a spec** — Look in `docs/specs/` for a file matching the issue number. If none exists for a feature task, stop and create one before proceeding.
4. **Check relevant ADRs** — Scan `docs/adr/` for decisions that affect the area you're working in.
5. **Read `AGENTS.md` for the service** — If you're modifying a specific service, read its `AGENTS.md` before touching any code.
6. **Check `packages/schema/proto/`** — If your task involves any data type (entity, enum, event, request/response shape), verify it's defined in proto. If not, define it there first before writing any service code.
7. **Check gstack artifacts** — Look in `~/.gstack/projects/` for test plans (`*-test-plan-*.md`), design docs (`*-design-*.md`), and review logs from prior sessions. Don't redo work that's already been saved there.

---

## Workflow Gates

### Gate 1: Spec Gate

- **Non-trivial features require a spec.** If `docs/specs/[ISSUE-NUMBER]-*.md` does not exist, create it using `docs/specs/TEMPLATE.md` before writing any implementation code.
- Bug fixes, dependency updates, and documentation changes are exempt.

### Gate 2: Plan Gate

- **Changes touching more than 2 files or introducing new architecture require an approved plan.**
- Use `EnterPlanMode` to explore, design, and present your plan. Do not write production code during planning.
- **Before exiting plan mode, run `/plan-eng-review` on your written plan.** Exit only after the plan is approved by the user.

### Gate 3: ADR Gate

- **Any architectural decision requires an ADR.** See `docs/adr/README.md` for what triggers an ADR.
- Create the ADR in `docs/adr/[NNNN]-[title].md` before implementing the decision.

---

## File and Code Rules

- **Never create a new service manually.** Always use `scripts/new-service.sh`.
- **Prefer editing existing files over creating new ones.**
- **Never skip git hooks.** Do not use `--no-verify` or `--no-gpg-sign`.
- **Never add AI attribution.** Do not add `Co-Authored-By: Claude` or any mention of Claude, Anthropic, or AI tooling in commit messages or PR bodies.
- **Trunk-based on `main`.** This project maintains everything on `main` (see CORE_RULES Rule 5). Commit and push to `main` directly; keep commits small and green. Never force-push or rewrite published history.
- **Never hardcode secrets.** Use `.env` files (gitignored). Update `.env.example` with placeholder values.
- **Never modify `infra/docker-compose.yml` or AWS configs without explicit human approval.**

---

## Task Tracking Discipline

Use the built-in task system for any multi-step work:

```
TaskCreate  → when starting a complex task
TaskUpdate  → mark in_progress before beginning, completed when done
TaskList    → check for next task after completing one
```

Break large tasks into smaller, independently completable units. Never mark a task complete if tests are failing or implementation is partial.

---

## Memory Management

Maintain `.claude/memory/` to preserve context across sessions:

- `MEMORY.md` — high-level summary, always loaded (keep under 200 lines)
- Topic files (e.g., `auth.md`, `database.md`) — detailed notes linked from `MEMORY.md`

Write to memory when you discover:

- Stable architectural patterns in this repo
- Non-obvious conventions or gotchas
- Important file paths and entry points
- Solutions to recurring problems

Do NOT write: session-specific state, incomplete conclusions, or anything that duplicates `CORE_RULES.md`.

### gstack Artifacts

gstack saves workflow artifacts outside the repo. Check these before starting planning or QA work:

```bash
source <(~/.claude/skills/gstack/bin/gstack-slug 2>/dev/null)
# $SLUG is now set (e.g. agarwalvivek29-myproject)
```

| Path                                        | Contents                              |
| ------------------------------------------- | ------------------------------------- |
| `~/.gstack/projects/$SLUG/*-test-plan-*.md` | QA test plans from `/plan-eng-review` |
| `~/.gstack/projects/$SLUG/*-design-*.md`    | Design docs from `/office-hours`      |
| `~/.gstack/projects/$SLUG/review-log.jsonl` | Review history (eng, CEO, design)     |

If a test plan or design doc already exists for the current branch, read it before re-doing research.

---

## Capabilities and Tools

You may use any available gstack skills, MCP servers, and tools as needed to complete tasks.

### gstack Skills

| Skill                 | When to use                                                                       |
| --------------------- | --------------------------------------------------------------------------------- |
| `/office-hours`       | Evaluating a new idea before any planning                                         |
| `/plan-eng-review`    | Before exiting plan mode — reviews your plan for architecture, tests, performance |
| `/plan-ceo-review`    | For significant product direction changes                                         |
| `/plan-design-review` | When UI/UX changes are involved                                                   |
| `/review`             | Pre-landing code review — run before pushing                                      |
| `/ship`               | Creating the PR — handles VERSION, CHANGELOG, branch sync                         |
| `/qa`                 | Browser-based QA after implementation                                             |
| `/debug`              | Systematic root-cause analysis when stuck or CI is failing                        |
| `/retro`              | Weekly retrospective                                                              |
| `/document-release`   | Post-ship doc updates                                                             |

### MCP Servers

Common useful MCPs:

- `filesystem` — file operations
- `github` — issue/PR management, branch operations
- `postgres`/`mongo` — database inspection (read-only in production)
- Web search — for researching libraries and patterns

Document any MCP or tool you add to a workflow in the relevant service's `AGENTS.md`.

---

## Service Modification Checklist

When modifying a service:

- [ ] Read `services/[name]/AGENTS.md`
- [ ] Spec exists in `docs/specs/`
- [ ] ADR created if architectural decision needed
- [ ] Plan approved (EnterPlanMode for >2 files, `/plan-eng-review` before exiting)
- [ ] Auth middleware applied to all new routes (check it's not bypassed)
- [ ] `JWT_SECRET` and `API_KEY` present in `.env.example` with placeholder values
- [ ] New data types defined in `packages/schema/proto/` — NOT in service code
- [ ] `packages/schema/generated/` regenerated and committed if proto changed
- [ ] Unit tests written for new domain functions (`tests/unit/`)
- [ ] E2E tests written for new API endpoints (`tests/e2e/`)
- [ ] `.env.example` updated if new env vars added
- [ ] `AGENTS.md` updated if service behavior or architecture changed
- [ ] All scaffold/example/placeholder code removed from the diff (see Rule 14 in `docs/CORE_RULES.md`)
- [ ] All commits follow `type(scope): description` format
- [ ] `/review` run before pushing
- [ ] `/ship` used to create the PR

---

## Creating a New Service

```bash
# Always use the scaffold script
./scripts/new-service.sh

# Then follow the printed next steps:
# 1. Create GitHub Issue
# 2. Write spec in docs/specs/
# 3. Create ADR if needed
# 4. Plan → implement → PR
```

---

## Schema-First Rule (Critical)

Before writing any type definition in service code:

1. Check if the type exists in `packages/schema/proto/`
2. If not → create the `.proto` definition first
3. Run `cd packages/schema && ./scripts/generate.sh`
4. Commit the proto + generated files
5. Import the generated type in the service

Never define a `type`, `interface`, `struct`, `class`, `dataclass`, or `enum` for a business domain concept in service code. The generated types from `packages/schema` are the only source of truth.

---

## Multi-Agent Strategy

| Scenario                      | Tool                                                     |
| ----------------------------- | -------------------------------------------------------- |
| N GitHub Issues in parallel   | `EnterWorktree` per issue — each gets an isolated branch |
| 1 issue, N parallel sub-tasks | gstack Teams (`TeamCreate` + `Agent` tool)               |
| Sequential work (default)     | Neither needed                                           |

**For worktree work:**

- Each worktree is isolated from `main` and every other session
- Push your feature branch when done; open a PR via `/ship`
- `.claude/` is shared across worktrees — memory accumulates correctly
- Never touch another worktree's branch or `main` directly

Do NOT nest gstack Teams inside a worktree that's already doing the same parallel work — double overhead with no benefit.

---

## Behavior Boundaries

Without explicit human approval, you MUST NOT:

- Push to remote branches
- Create, close, or comment on GitHub Issues or PRs
- Modify CI/CD pipeline configurations
- Drop or migrate databases
- Change infrastructure (AWS, docker-compose services)
- Delete files that may represent in-progress work
- Force-push any branch

When in doubt: stop, explain what you were about to do, and ask.

---
