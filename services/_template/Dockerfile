# ─────────────────────────────────────────────────────────────────────────────
# PRODUCTION Dockerfile — [SERVICE_NAME]
#
# For local development with hot-reload, use Dockerfile.dev instead.
#
# Choose ONE language section and delete the others.
# Then remove [service_name] / [SERVICE_NAME] placeholders.
# ─────────────────────────────────────────────────────────────────────────────

# ═══════════════════ TYPESCRIPT ══════════════════════════════════════════════
FROM node:22-alpine AS ts-deps
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN corepack enable pnpm && pnpm install --frozen-lockfile

FROM ts-deps AS ts-build
COPY . .
# Import generated schema types from the monorepo root (mounted in CI)
RUN pnpm build

FROM node:22-alpine AS production
WORKDIR /app
RUN addgroup -S app && adduser -S app -G app
COPY --from=ts-build /app/dist ./dist
COPY --from=ts-deps /app/node_modules ./node_modules
COPY package.json ./
USER app
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:3000/health || exit 1
CMD ["node", "dist/main.js"]

# ═══════════════════ PYTHON ══════════════════════════════════════════════════
FROM python:3.12-slim AS production
WORKDIR /app
RUN addgroup --system app && adduser --system --ingroup app app
RUN pip install uv
COPY pyproject.toml uv.lock* ./
RUN uv sync --frozen --no-dev
COPY src/ ./src/
USER app
EXPOSE 8000
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/health')" || exit 1
CMD ["uv", "run", "python", "-m", "[service_name]"]

# ═══════════════════ GO ══════════════════════════════════════════════════════
FROM golang:1.23-alpine AS go-build
WORKDIR /app
# Install protobuf compiler (needed if using prost/proto in build)
RUN apk add --no-cache protobuf-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server

FROM gcr.io/distroless/static:nonroot AS production
COPY --from=go-build /app/server /server
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/server", "healthcheck"]
CMD ["/server"]

# ═══════════════════ RUST ════════════════════════════════════════════════════
FROM rust:1.82-alpine AS rust-build
RUN apk add --no-cache musl-dev protobuf-dev
WORKDIR /app
COPY Cargo.toml Cargo.lock ./
# Cache dependencies with a dummy main
RUN mkdir src && echo "fn main() {}" > src/main.rs \
  && cargo build --release \
  && rm -rf src
COPY . .
RUN touch src/main.rs && cargo build --release

FROM gcr.io/distroless/cc:nonroot AS production
COPY --from=rust-build /app/target/release/[service_name] /app
EXPOSE 8080
CMD ["/app"]
