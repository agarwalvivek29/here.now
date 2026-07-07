#!/usr/bin/env bash
# new-app.sh — Scaffold a new frontend app
# Usage: ./scripts/new-app.sh
# Or with args: ./scripts/new-app.sh [name] [framework]

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APPS_DIR="$REPO_ROOT/apps"

# ─── Colors ───────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

print_step() { echo -e "\n${BLUE}▶${NC} ${BOLD}$1${NC}"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1" >&2; }

# ─── Collect inputs ───────────────────────────────────────────────────────────
print_step "New App Scaffold"
echo ""

# App name
if [ -n "${1:-}" ]; then
  APP_NAME="$1"
else
  read -rp "App name (kebab-case, e.g. web-app): " APP_NAME
fi

if [[ ! "$APP_NAME" =~ ^[a-z][a-z0-9-]*$ ]]; then
  print_error "App name must be lowercase kebab-case"
  exit 1
fi

if [ -d "$APPS_DIR/$APP_NAME" ]; then
  print_error "App '$APP_NAME' already exists at $APPS_DIR/$APP_NAME"
  exit 1
fi

# Framework
if [ -n "${2:-}" ]; then
  FRAMEWORK="$2"
else
  echo ""
  echo "Framework:"
  echo "  1) Next.js (React, SSR/SSG)"
  echo "  2) Vite + React (SPA)"
  echo "  3) Vite + Vue"
  echo "  4) Astro (content sites)"
  echo "  5) Remix"
  read -rp "Choose [1-5]: " FW_CHOICE
  case "$FW_CHOICE" in
    1) FRAMEWORK="nextjs" ;;
    2) FRAMEWORK="vite-react" ;;
    3) FRAMEWORK="vite-vue" ;;
    4) FRAMEWORK="astro" ;;
    5) FRAMEWORK="remix" ;;
    *) print_error "Invalid choice"; exit 1 ;;
  esac
fi

# GitHub Issue number
echo ""
read -rp "GitHub Issue number (press Enter to skip): " ISSUE_NUMBER
ISSUE_NUMBER="${ISSUE_NUMBER:-TBD}"

# ─── Confirm ──────────────────────────────────────────────────────────────────
echo ""
echo "─────────────────────────────────────"
echo "  App name:   ${BOLD}$APP_NAME${NC}"
echo "  Framework:  ${BOLD}$FRAMEWORK${NC}"
echo "  Issue:      ${BOLD}#$ISSUE_NUMBER${NC}"
echo "  Directory:  $APPS_DIR/$APP_NAME"
echo "─────────────────────────────────────"
echo ""
read -rp "Proceed? [y/N] " CONFIRM
if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
  echo "Aborted."
  exit 0
fi

# ─── Scaffold ─────────────────────────────────────────────────────────────────
print_step "Creating app directory..."
mkdir -p "$APPS_DIR/$APP_NAME"

# ─── Create AGENTS.md ────────────────────────────────────────────────────────
print_step "Creating AGENTS.md..."
TODAY=$(date +%Y-%m-%d)
cat > "$APPS_DIR/$APP_NAME/AGENTS.md" << EOF
# AGENTS.md — $APP_NAME

> Agent contract for the \`$APP_NAME\` frontend app.
> Read this before modifying anything in this app directory.

## App Overview

**Name**: \`$APP_NAME\`
**Framework**: $FRAMEWORK (TypeScript)
**Purpose**: [Describe what this app does]
**Owner**: [team or person]
**Created**: $TODAY
**Issue**: #$ISSUE_NUMBER
**Spec**: \`docs/specs/$ISSUE_NUMBER-$APP_NAME.md\`

---

## Tech Stack

- **Framework**: $FRAMEWORK
- **Language**: TypeScript (strict mode)
- **Package manager**: pnpm
- **Styling**: [Tailwind / CSS Modules / styled-components]
- **State management**: [Zustand / TanStack Query / Context / etc.]
- **Testing**: [Vitest / Jest + Testing Library / Playwright]

---

## Key Entry Points

- **Main**: \`src/app/page.tsx\` (Next.js) or \`src/main.tsx\` (Vite)
- **Config**: \`src/config/env.ts\` — validated env vars, import this not \`process.env\`
- **API client**: \`src/lib/api.ts\`

---

## Environment Variables

See \`.env.example\` for all required variables.

---

## Local Development

\`\`\`bash
cd apps/$APP_NAME
cp .env.example .env.local
pnpm install
pnpm dev
\`\`\`

---

## Forbidden Actions for Agents

- Changing the build output target or deployment config without approval
- Removing TypeScript strict mode
- Adding tracking scripts or analytics without product approval

---

## Agent Capabilities

Agents may use any available MCP servers, skills, and tools as needed.

### MCP Servers in Use

| MCP Server | Purpose | Added by |
|---|---|---|
| (none yet) | | |
EOF
print_success "AGENTS.md created"

# ─── Create .env.example ──────────────────────────────────────────────────────
print_step "Creating .env.example..."
cat > "$APPS_DIR/$APP_NAME/.env.example" << EOF
# Environment variables for $APP_NAME
# Copy to .env.local for Next.js or .env for Vite

# ─── API ───────────────────────────────────────────────────────────────────
NEXT_PUBLIC_API_URL=http://localhost:3001
# VITE_API_URL=http://localhost:3001

# ─── Auth ──────────────────────────────────────────────────────────────────
# NEXT_PUBLIC_AUTH_URL=http://localhost:3002

# ─── Feature flags ─────────────────────────────────────────────────────────
# NEXT_PUBLIC_FEATURE_X=false
EOF
print_success ".env.example created"

# ─── Create README.md ─────────────────────────────────────────────────────────
print_step "Creating README.md..."
cat > "$APPS_DIR/$APP_NAME/README.md" << EOF
# $APP_NAME

> Replace with a one-paragraph description of this app.

**Framework**: $FRAMEWORK (TypeScript)
**Issue**: #$ISSUE_NUMBER

---

## Local Development

\`\`\`bash
cd apps/$APP_NAME
cp .env.example .env.local  # or .env for Vite
pnpm install
pnpm dev
\`\`\`

---

## Testing

\`\`\`bash
pnpm test         # unit tests
pnpm test:e2e     # end-to-end tests (Playwright)
\`\`\`

---

## Build

\`\`\`bash
pnpm build
pnpm start        # preview production build
\`\`\`
EOF
print_success "README.md created"

# ─── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}✓ App '$APP_NAME' scaffolded successfully!${NC}"
echo ""
echo -e "${BOLD}Next steps:${NC}"
echo ""
echo "  1. ${YELLOW}Initialize the framework${NC}"

case "$FRAMEWORK" in
  nextjs)
    echo "     cd apps && pnpm create next-app $APP_NAME --typescript --tailwind --eslint --app --src-dir"
    ;;
  vite-react)
    echo "     cd apps && pnpm create vite $APP_NAME --template react-ts"
    ;;
  vite-vue)
    echo "     cd apps && pnpm create vite $APP_NAME --template vue-ts"
    ;;
  astro)
    echo "     cd apps && pnpm create astro $APP_NAME"
    ;;
  remix)
    echo "     cd apps && pnpm create remix $APP_NAME"
    ;;
esac

echo ""
echo "  2. ${YELLOW}Write the spec${NC}"
echo "     cp docs/specs/TEMPLATE.md docs/specs/${ISSUE_NUMBER}-${APP_NAME}.md"
echo ""
echo "  3. ${YELLOW}Customize AGENTS.md${NC}"
echo "     apps/$APP_NAME/AGENTS.md — fill in purpose, stack, entry points"
echo ""
echo "  4. ${YELLOW}Plan before coding${NC}"
echo "     Agents: use EnterPlanMode | Humans: post a plan on the issue"
echo ""
