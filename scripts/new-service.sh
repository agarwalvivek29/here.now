#!/usr/bin/env bash
# new-service.sh — Scaffold a new microservice from the _template
# Usage: ./scripts/new-service.sh
# Or with args: ./scripts/new-service.sh [name] [language] [type]

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEMPLATE_DIR="$REPO_ROOT/services/_template"
SERVICES_DIR="$REPO_ROOT/services"
COMPOSE_FILE="$REPO_ROOT/infra/docker-compose.yml"

# ─── Colors ───────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

print_step() { echo -e "\n${BLUE}▶${NC} ${BOLD}$1${NC}"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1" >&2; }

# ─── Validate environment ─────────────────────────────────────────────────────
if [ ! -d "$TEMPLATE_DIR" ]; then
  print_error "Template directory not found: $TEMPLATE_DIR"
  exit 1
fi

# ─── Collect inputs ───────────────────────────────────────────────────────────
print_step "New Service Scaffold"
echo "This script creates a new microservice from the _template."
echo "Read docs/CORE_RULES.md and docs/ONBOARDING.md before proceeding."
echo ""

# Service name
if [ -n "${1:-}" ]; then
  SERVICE_NAME="$1"
else
  read -rp "Service name (kebab-case, e.g. auth-service): " SERVICE_NAME
fi

# Validate service name
if [[ ! "$SERVICE_NAME" =~ ^[a-z][a-z0-9-]*$ ]]; then
  print_error "Service name must be lowercase kebab-case (e.g. my-service)"
  exit 1
fi

# Check if service already exists
if [ -d "$SERVICES_DIR/$SERVICE_NAME" ]; then
  print_error "Service '$SERVICE_NAME' already exists at $SERVICES_DIR/$SERVICE_NAME"
  exit 1
fi

# Language
if [ -n "${2:-}" ]; then
  LANGUAGE="$2"
else
  echo ""
  echo "Language:"
  echo "  1) TypeScript"
  echo "  2) Python"
  echo "  3) Go"
  echo "  4) Rust"
  read -rp "Choose [1-4]: " LANG_CHOICE
  case "$LANG_CHOICE" in
    1) LANGUAGE="typescript" ;;
    2) LANGUAGE="python" ;;
    3) LANGUAGE="go" ;;
    4) LANGUAGE="rust" ;;
    *) print_error "Invalid choice"; exit 1 ;;
  esac
fi

# Service type
if [ -n "${3:-}" ]; then
  SERVICE_TYPE="$3"
else
  echo ""
  echo "Service type:"
  echo "  1) REST API"
  echo "  2) GraphQL API"
  echo "  3) WebSocket"
  echo "  4) gRPC"
  echo "  5) Worker (queue consumer)"
  echo "  6) CLI tool"
  echo "  7) Agentic (AI agent service)"
  read -rp "Choose [1-7]: " TYPE_CHOICE
  case "$TYPE_CHOICE" in
    1) SERVICE_TYPE="rest-api" ;;
    2) SERVICE_TYPE="graphql" ;;
    3) SERVICE_TYPE="websocket" ;;
    4) SERVICE_TYPE="grpc" ;;
    5) SERVICE_TYPE="worker" ;;
    6) SERVICE_TYPE="cli" ;;
    7) SERVICE_TYPE="agentic" ;;
    *) print_error "Invalid choice"; exit 1 ;;
  esac
fi

# GitHub Issue number
echo ""
read -rp "GitHub Issue number (for spec linkage, press Enter to skip): " ISSUE_NUMBER
ISSUE_NUMBER="${ISSUE_NUMBER:-TBD}"

# ─── Confirm ──────────────────────────────────────────────────────────────────
echo ""
echo "─────────────────────────────────────"
echo "  Service name: ${BOLD}$SERVICE_NAME${NC}"
echo "  Language:     ${BOLD}$LANGUAGE${NC}"
echo "  Type:         ${BOLD}$SERVICE_TYPE${NC}"
echo "  Issue:        ${BOLD}#$ISSUE_NUMBER${NC}"
echo "  Directory:    $SERVICES_DIR/$SERVICE_NAME"
echo "─────────────────────────────────────"
echo ""
read -rp "Proceed? [y/N] " CONFIRM
if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
  echo "Aborted."
  exit 0
fi

# ─── Scaffold ─────────────────────────────────────────────────────────────────
print_step "Copying template..."
cp -r "$TEMPLATE_DIR" "$SERVICES_DIR/$SERVICE_NAME"
print_success "Copied template to $SERVICES_DIR/$SERVICE_NAME"

# ─── Populate AGENTS.md ───────────────────────────────────────────────────────
print_step "Populating AGENTS.md..."
TODAY=$(date +%Y-%m-%d)
AGENTS_FILE="$SERVICES_DIR/$SERVICE_NAME/AGENTS.md"

sed -i.bak \
  -e "s/\[SERVICE_NAME\]/$SERVICE_NAME/g" \
  -e "s/\[service_name\]/$SERVICE_NAME/g" \
  -e "s/\[ISSUE-NUMBER\]/$ISSUE_NUMBER/g" \
  -e "s/YYYY-MM-DD/$TODAY/g" \
  -e "s/\[TypeScript \/ Python \/ Go \/ Rust\]/$LANGUAGE/g" \
  -e "s/\[REST API \/ GraphQL API \/ WebSocket \/ gRPC \/ Worker \/ CLI \/ Agentic\]/$SERVICE_TYPE/g" \
  "$AGENTS_FILE"
rm -f "$AGENTS_FILE.bak"
print_success "AGENTS.md populated"

# ─── Populate other template files ────────────────────────────────────────────
print_step "Updating README.md, .env.example, and Dockerfiles..."
for file in \
    "$SERVICES_DIR/$SERVICE_NAME/README.md" \
    "$SERVICES_DIR/$SERVICE_NAME/.env.example" \
    "$SERVICES_DIR/$SERVICE_NAME/Dockerfile" \
    "$SERVICES_DIR/$SERVICE_NAME/Dockerfile.dev"; do
  if [ -f "$file" ]; then
    sed -i.bak \
      -e "s/\[SERVICE_NAME\]/$SERVICE_NAME/g" \
      -e "s/\[service_name\]/$SERVICE_NAME/g" \
      -e "s/\[ISSUE-NUMBER\]/$ISSUE_NUMBER/g" \
      -e "s/\[TypeScript \/ Python \/ Go \/ Rust\]/$LANGUAGE/g" \
      -e "s/\[REST API \/ GraphQL API \/ WebSocket \/ gRPC \/ Worker \/ CLI \/ Agentic\]/$SERVICE_TYPE/g" \
      "$file"
    rm -f "$file.bak"
  fi
done
print_success "README.md, .env.example, and Dockerfiles updated"

# ─── Trim Dockerfiles to chosen language ──────────────────────────────────────
print_step "Trimming Dockerfiles for $LANGUAGE..."

# Remove every language section except the chosen one.
# Section headers are lines starting with '# ' and containing '═' characters.
# The section header line itself is also removed (redundant with only one language).
trim_dockerfile() {
  local file="$1"
  local lang="$2"

  awk -v lang="$lang" '
    /^#/ && /[═]/ {
      if      (/TYPESCRIPT/) { in_section=1; keep=(lang=="typescript") }
      else if (/PYTHON/)     { in_section=1; keep=(lang=="python") }
      else if (/ GO /)       { in_section=1; keep=(lang=="go") }
      else if (/RUST/)       { in_section=1; keep=(lang=="rust") }
      next
    }
    { if (!in_section || keep) print }
  ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"
}

trim_dockerfile "$SERVICES_DIR/$SERVICE_NAME/Dockerfile" "$LANGUAGE"
trim_dockerfile "$SERVICES_DIR/$SERVICE_NAME/Dockerfile.dev" "$LANGUAGE"
print_success "Dockerfiles trimmed to $LANGUAGE only"

# ─── Add docker-compose stub ──────────────────────────────────────────────────
print_step "Adding docker-compose stub..."

# Determine default port based on type
case "$SERVICE_TYPE" in
  grpc) DEFAULT_PORT="50051" ;;
  *) DEFAULT_PORT="3000" ;;
esac

# Convert kebab-case to UPPER_SNAKE for env var name
SERVICE_UPPER="${SERVICE_NAME^^}"
SERVICE_UPPER="${SERVICE_UPPER//-/_}"

# Append a commented stub to docker-compose.yml
cat >> "$COMPOSE_FILE" << EOF

  # ─── $SERVICE_NAME ──────────────────────────────────────────────────────────
  # Uncomment to enable. Uses Dockerfile.dev for hot-reload.
  # Dockerfile (production build) is NOT used in local dev.
  #
  # $SERVICE_NAME:
  #   build:
  #     context: ../services/$SERVICE_NAME
  #     dockerfile: Dockerfile.dev
  #   restart: unless-stopped
  #   env_file:
  #     - ../services/$SERVICE_NAME/.env
  #   ports:
  #     - "\${${SERVICE_UPPER}_PORT:-$DEFAULT_PORT}:$DEFAULT_PORT"
  #   volumes:
  #     - ../services/$SERVICE_NAME/src:/app/src              # hot reload
  #     - ../packages/schema/generated:/schema/generated      # schema types
  #   networks:
  #     - app-network
EOF

print_success "docker-compose stub added (commented out)"

# ─── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}✓ Service '$SERVICE_NAME' scaffolded successfully!${NC}"
echo ""
echo -e "${BOLD}Next steps:${NC}"
echo ""
echo "  1. ${YELLOW}Create a GitHub Issue${NC} (if not done already)"
echo "     Use the 'New Service' issue template."
echo ""
echo "  2. ${YELLOW}Write the spec${NC}"
echo "     cp docs/specs/TEMPLATE.md docs/specs/${ISSUE_NUMBER}-${SERVICE_NAME}.md"
echo "     # Fill it in before writing any code"
echo ""
echo "  3. ${YELLOW}Create an ADR${NC} (architectural decision required for new services)"
echo "     See docs/adr/README.md for the next ADR number."
echo ""
echo "  4. ${YELLOW}Customize AGENTS.md${NC}"
echo "     services/$SERVICE_NAME/AGENTS.md — fill in purpose, interfaces, constraints"
echo ""
echo "  5. ${YELLOW}Trim the Dockerfile${NC}"
echo "     Keep only the language-specific stages for $LANGUAGE"
echo ""
echo "  6. ${YELLOW}Enable in docker-compose${NC}"
echo "     Uncomment the $SERVICE_NAME stub in infra/docker-compose.yml"
echo ""
echo "  7. ${YELLOW}Define your data models in packages/schema${NC}"
echo "     Add .proto files for all entities, enums, events, and API shapes"
echo "     cd packages/schema && ./scripts/generate.sh"
echo "     Import generated types — never define types locally in the service"
echo ""
echo "  8. ${YELLOW}Plan before coding${NC}"
echo "     Agents: use EnterPlanMode | Humans: post a plan on the issue"
echo ""
