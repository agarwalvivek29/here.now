#!/usr/bin/env bash
# generate.sh — Regenerate all language bindings from proto files
#
# Runs: buf lint → buf breaking (vs main) → buf generate
# Output: generated/{ts,go,python}/
# Rust: not generated here — uses prost-build in each service's build.rs
#
# Requirements:
#   - buf: https://buf.build/docs/installation (brew install bufbuild/buf/buf)
#   - Go:  protoc-gen-go, protoc-gen-go-grpc (installed below if missing)
#   - Python: betterproto[compiler] (pip install betterproto[compiler])
#
# Usage: ./scripts/generate.sh [--skip-breaking] [--skip-lint]

set -euo pipefail

SCHEMA_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKIP_BREAKING=false
SKIP_LINT=false

# ─── Parse flags ─────────────────────────────────────────��────────────────────
for arg in "$@"; do
  case $arg in
    --skip-breaking) SKIP_BREAKING=true ;;
    --skip-lint) SKIP_LINT=true ;;
  esac
done

cd "$SCHEMA_DIR"

# ─── Colors ───────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'
BOLD='\033[1m'

step() { echo -e "\n${BOLD}▶ $1${NC}"; }
ok()   { echo -e "${GREEN}✓${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }
fail() { echo -e "${RED}✗${NC} $1" >&2; }

# ─── Check buf is installed ───────────────────────────────────────────────────
if ! command -v buf &> /dev/null; then
  fail "buf is not installed."
  echo ""
  echo "Install with: brew install bufbuild/buf/buf"
  echo "Or: curl -sSL https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m) -o /usr/local/bin/buf && chmod +x /usr/local/bin/buf"
  exit 1
fi

step "buf version: $(buf --version)"

# ─── Install Go plugins if missing ────────────────────────────────────────────
if command -v go &> /dev/null; then
  step "Installing Go protoc plugins..."
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ok "Go plugins installed"
else
  warn "Go not found — Go code generation will use buf remote plugins (requires buf login)"
fi

# ─── Check Python betterproto ─────────────────────────────────────────────────
if command -v python3 &> /dev/null; then
  if ! python3 -c "import betterproto" 2>/dev/null; then
    step "Installing betterproto..."
    pip install "betterproto[compiler]" --quiet
    ok "betterproto installed"
  fi
else
  warn "Python not found — Python code generation will be skipped"
fi

# ─── Lint ─────────────────────────────────────────────────────────────────────
if [ "$SKIP_LINT" = false ]; then
  step "Linting proto files..."
  buf lint
  ok "Proto lint passed"
fi

# ─── Breaking change detection ────────────────────────────────────────────────
if [ "$SKIP_BREAKING" = false ]; then
  step "Checking for breaking changes vs main..."
  # Only runs if we're on a branch (not main itself)
  CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
  if [ -n "$CURRENT_BRANCH" ] && [ "$CURRENT_BRANCH" != "main" ]; then
    if git show-ref --verify --quiet refs/heads/main 2>/dev/null || \
       git show-ref --verify --quiet refs/remotes/origin/main 2>/dev/null; then
      buf breaking --against '.git#branch=main,subdir=packages/schema' || {
        fail "Breaking changes detected vs main."
        echo ""
        echo "If this is intentional, create an ADR documenting the breaking change"
        echo "and update all affected service consumers before merging."
        echo ""
        echo "To skip this check: ./scripts/generate.sh --skip-breaking"
        exit 1
      }
      ok "No breaking changes detected"
    else
      warn "main branch not found locally — skipping breaking change check"
    fi
  else
    warn "On main branch — skipping breaking change check"
  fi
fi

# ─── Generate ─────────────────────────────────────────────────────────────────
step "Generating code from proto files..."
buf generate
ok "Code generation complete"

# ─── Summary ──────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}✓ Generation complete!${NC}"
echo ""
echo "Generated output:"
echo "  TypeScript: generated/ts/"
echo "  Go:         generated/go/"
echo "  Python:     generated/python/"
echo ""
echo -e "${YELLOW}Rust:${NC} Generated at build time via prost-build in each service's build.rs"
echo "  See examples/rust-build.rs for the pattern"
echo ""
echo "Next steps:"
echo "  1. Review the generated files"
echo "  2. Commit the changes: git add packages/schema/generated/ && git commit"
echo "  3. Update imports in services that use the changed types"
