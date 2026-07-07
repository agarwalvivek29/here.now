# Onboarding: Starting a New Project from This Template

This guide walks through how to take this template and bootstrap a new production project.

---

## Step 1: Use This Template

```bash
# On GitHub: click "Use this template" → "Create a new repository"
# Or clone directly:
git clone https://github.com/agarwalvivek29/here.now.git my-project
cd my-project
git remote set-url origin https://github.com/agarwalvivek29/here.now.git
git push -u origin main
```

---

## Step 2: Initialize the Repository

```bash
# Install git hooks (requires Node.js for husky)
npm install   # installs husky from package.json

# Make scripts executable
chmod +x scripts/*.sh

# Verify hooks are installed
ls .husky/
```

---

## Step 3: Configure Your Repository

### GitHub Settings

- Enable branch protection on `main`:
  - Require PR before merging
  - Require status checks to pass (lint, spec-check)
  - Require at least 1 approval
- Set up repository secrets for CI (see `infra/aws/README.md`)

### Update Root Files

- `README.md` — Replace template content with your project description
- `docs/CORE_RULES.md` — Adjust any project-specific rules (add don't remove)
- `docs/CONVENTIONS.md` — Remove language sections not used in your project

---

## Step 4: Create Your First Services

```bash
# For each backend service:
./scripts/new-service.sh
# Follow the prompts: name, language, type

# For each frontend app:
./scripts/new-app.sh
# Follow the prompts: name, framework
```

Each service gets its own `AGENTS.md`, `Dockerfile`, and `.env.example`.

---

## Step 5: Configure Local Infrastructure

Edit `infra/docker-compose.yml` to uncomment the infra services you need:

```yaml
# Uncomment PostgreSQL if needed
# Uncomment Redis if needed
# etc.
```

Then add your application service stubs.

```bash
# Verify the compose file is valid
docker compose -f infra/docker-compose.yml config

# Start local infra
docker compose -f infra/docker-compose.yml up -d
```

---

## Step 6: Record Your First ADR

Document the key decisions made when setting up this project:

```bash
# Copy the ADR template
cp docs/adr/0001-monorepo-structure.md docs/adr/0002-[your-first-decision].md
# Edit it to describe your decision
```

ADR 0001 (monorepo structure) already exists in the template — keep it or update it to match your project.

---

## Step 7: Your First Feature

Follow the full workflow:

1. **Open a GitHub Issue** using the `feature-request.md` template
2. **Create a spec**: `docs/specs/[ISSUE-NUMBER]-[name].md` from `docs/specs/TEMPLATE.md`
3. **Create an ADR** if the feature requires an architectural decision
4. **Create a branch**: `git checkout -b feat/[ISSUE-NUMBER]-[short-desc]`
5. **Plan** (agents: `EnterPlanMode`; humans: comment on the issue)
6. **Implement** following `docs/CONVENTIONS.md`
7. **PR** using the PR template — reference the issue and spec

---

## Repository Structure Quick Reference

```
.
├── CLAUDE.md          ← Claude Code agent contract
├── AGENTS.md          ← Generic AI agent rules
├── docs/
│   ├── CORE_RULES.md  ← The law
│   ├── CONVENTIONS.md ← Language standards
│   ├── adr/           ← Architecture decisions
│   └── specs/         ← Feature specs (one per issue)
├── apps/              ← Frontend applications
├── services/          ← Backend microservices
│   └── _template/     ← Copy this for new services
├── packages/          ← Shared libraries
├── infra/             ← docker-compose, AWS
└── scripts/           ← new-service.sh, new-app.sh
```

---

## Common Operations

| Task                      | Command                                            |
| ------------------------- | -------------------------------------------------- |
| Create a new service      | `./scripts/new-service.sh`                         |
| Create a new frontend app | `./scripts/new-app.sh`                             |
| Start local infra         | `docker compose -f infra/docker-compose.yml up -d` |
| Stop local infra          | `docker compose -f infra/docker-compose.yml down`  |
| Run all linters           | `npm run lint` (root)                              |

---

## Step 8: Set Up Agent Orchestrator (ao) — Recommended

`ao` manages parallel AI agent sessions: spawning agents on GitHub Issues, monitoring PR/CI/review state, and notifying you when your judgment is needed.

### Prerequisites

```bash
brew install tmux                                      # session runtime
npm install -g @anthropic-ai/claude-code               # AI coding agent
brew install gh && gh auth login                       # GitHub CLI (for tracker + SCM)
npm install -g @composio/ao-cli                        # the orchestrator itself
```

### Configure

```bash
# Copy the example config and fill in your repo + path
cp agent-orchestrator.yaml.example agent-orchestrator.yaml
# Edit agent-orchestrator.yaml:
#   - Set projects.my-project.repo to your GitHub repo (owner/repo)
#   - Set projects.my-project.path to the absolute local path
```

### Initialize

```bash
# Install workspace hooks (enables automatic PR/branch metadata updates)
ao init

# Verify everything is wired up
ao status
```

### Daily workflow

```bash
# Start the orchestrator (dashboard + lifecycle manager)
ao start

# Spawn an agent on a GitHub Issue
ao spawn my-project #42

# Spawn multiple agents in parallel
ao batch-spawn my-project #42 #43 #44

# Watch all sessions
ao status

# Send a message to a running session
ao send svc-1 "Please address the review comments"
```

The web dashboard runs at `http://localhost:3000` — live session cards, PR/CI status, attention zones for merge-ready and stuck agents.

**How it connects to your workflow:**

- Well-written GitHub Issues (using the issue templates) are the agent's task prompt — the more detail, the better the output
- `agentRulesFile: docs/CORE_RULES.md` in `agent-orchestrator.yaml` injects your binding rules into every agent's system prompt automatically
- `.claude` is symlinked into every worktree so agents share the same permissions and accumulated memory
- ao auto-forwards CI failures and review comments back to agents; you only get notified when human judgment is genuinely needed

---

## For AI Agents

If you are an AI agent starting work in this repository:

1. Read `CLAUDE.md` (if Claude Code) or `AGENTS.md` (all agents)
2. Read `docs/CORE_RULES.md`
3. Find your task in the linked GitHub Issue
4. Verify a spec exists in `docs/specs/`
5. Read the service's `AGENTS.md` before touching any service code

Do not write implementation code before completing these steps.
