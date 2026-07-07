# Coding Conventions

Language-specific standards for this monorepo. All contributions must follow these conventions.
Agents must read the relevant section before modifying or generating code.

---

## General (All Languages)

- **No magic numbers** — use named constants
- **No commented-out code** — delete dead code, use version control to recover it
- **No TODOs in merged code** — convert to GitHub Issues before merging
- **Descriptive names** — variable and function names must explain what they do
- **Single responsibility** — functions do one thing; files have a clear, single purpose
- **Error handling at the boundary** — validate at system entry points (API handlers, queue consumers); trust internal code

---

## TypeScript

### Tooling

- **Formatter**: Prettier (config in `.prettierrc`)
- **Linter**: ESLint with `@typescript-eslint` (config in `eslint.config.mjs`)
- **Runtime**: Node.js (LTS) or Bun
- **Package manager**: pnpm (workspace-level `pnpm-workspace.yaml`)

### Rules

- `strict: true` in `tsconfig.json` — no exceptions
- No `any` — use `unknown` and narrow types explicitly
- No `ts-ignore` or `ts-expect-error` without a comment explaining why
- Use `interface` for object shapes, `type` for unions/intersections/aliases
- Path aliases via `tsconfig.json` `paths` — no relative `../../../` chains beyond 2 levels
- Imports ordered: built-ins → external → internal (enforced by ESLint)
- Export from `index.ts` barrel files for public APIs only; avoid over-exporting internals
- Use `zod` or `valibot` for runtime validation of external data (API responses, env vars)

### Environment variables

```typescript
// Use a validated env module — never import process.env directly in business logic
import { env } from "@/config/env";
```

### API services

- OpenAPI spec in `openapi.yaml` at service root — generated or hand-authored, kept in sync
- Use `zod` schemas derived from OpenAPI spec for request/response validation
- `async/await` throughout — no raw `.then()` chains

### Agentic services

- **Framework**: [Mastra AI](https://mastra.ai) — the recommended TypeScript agent framework
- Agent definitions, tools, workflows, memory, and RAG all use Mastra primitives
- Tool input/output schemas must use `zod` — and the underlying types must come from `packages/schema`
- Agents are structured as Mastra `Agent` instances with named, typed tools
- See the `## AI / Agentic Services` section for cross-language conventions

---

## Python

### Tooling

- **Formatter**: ruff format (replaces black — use `ruff format` not `black`)
- **Linter**: ruff
- **Type checker**: mypy (strict mode)
- **Dependency management**: uv + `pyproject.toml` — the only approved package manager; never use pip, poetry, or conda directly
- **Test framework**: pytest

### Rules

- Type hints required on all function signatures (enforced by mypy)
- No bare `except:` — always catch specific exceptions
- Use `pydantic` for data models and config validation
- Use `structlog` or `loguru` for structured logging — no `print()`
- `async` with `asyncio` for I/O-bound services; use `fastapi` for HTTP
- Docstrings only on public APIs and non-obvious logic

### Project layout

```
service/
├── src/
│   └── [service_name]/
│       ├── __init__.py
│       ├── main.py          # entry point
│       ├── api/             # route handlers
│       ├── domain/          # business logic
│       ├── infra/           # DB, queue, external clients
│       └── config.py        # pydantic settings
├── tests/
├── pyproject.toml
└── Dockerfile
```

### Environment variables

```python
# Use pydantic-settings
from [service_name].config import settings
```

### Agentic services

- **Framework**: [Agno](https://docs.agno.com) — the recommended Python agent framework
- Agents are defined as `agno.Agent` instances with typed tools and structured output
- Tool parameters and return types use Pydantic models — sourced from `packages/schema` generated types
- Use Agno's built-in memory, knowledge base, and storage integrations
- See the `## AI / Agentic Services` section for cross-language conventions

---

## Go

### Tooling

- **Formatter**: gofmt (automatic)
- **Linter**: golangci-lint (config in `.golangci.yml`)
- **Dependency management**: Go modules (`go.mod`)
- **Test framework**: standard `testing` package + testify

### Rules

- Follow the [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- Error wrapping: `fmt.Errorf("context: %w", err)` — always add context
- No `panic` in library/service code — return errors
- Use `slog` (stdlib) for structured logging
- Use `context.Context` as first parameter for all I/O operations
- Interfaces defined at the consumer side, not the implementation side
- Use table-driven tests

### Project layout

```
service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   ├── domain/
│   └── infra/
├── pkg/              # exported utilities (if needed)
├── migrations/
├── go.mod
└── Dockerfile
```

---

## Rust

### Tooling

- **Formatter**: rustfmt (`rustfmt.toml`)
- **Linter**: clippy (deny on warnings in CI: `RUSTFLAGS="-D warnings"`)
- **Dependency management**: Cargo workspace (`Cargo.toml` at monorepo root)
- **Test framework**: built-in `#[test]`, integration tests in `tests/`
- **Async runtime**: tokio

### Rules

- No `unwrap()` or `expect()` in library/production code — propagate errors
- Use `thiserror` for defining errors, `anyhow` for application error handling
- Use `tracing` for structured logging/spans
- Prefer `Arc<Mutex<T>>` only when necessary; design for ownership
- All public items must have doc comments (`///`)
- Prefer iterators and functional style over explicit loops where readable

### Project layout

```
service/
├── src/
│   ├── main.rs
│   ├── api/
│   ├── domain/
│   └── infra/
├── tests/
├── Cargo.toml
└── Dockerfile
```

---

## Database

- **Migrations**: managed by migration tool — never apply raw SQL manually
  - TypeScript: `drizzle-kit` or `node-pg-migrate`
  - Python: `alembic`
  - Go: `goose`
  - Rust: `sqlx migrate`
- Migrations are append-only — never edit a committed migration
- Every migration has a corresponding down migration
- Migration files live in `migrations/` at the service root
- Index names follow: `idx_[table]_[columns]`
- Never use `SELECT *` in application code — always name columns explicitly

---

## API Design

### REST

- Follow RESTful resource naming: `GET /users/:id`, `POST /users`, `PATCH /users/:id`
- Use HTTP status codes correctly (200, 201, 400, 401, 403, 404, 409, 422, 500)
- Always return structured error responses: `{ "error": { "code": "...", "message": "..." } }`
- Version APIs via URL prefix: `/v1/...`
- OpenAPI spec required, kept in sync with implementation

### GraphQL

- Schema-first development — schema is the contract
- No N+1 queries — use DataLoader pattern
- Input validation via custom scalars and input types

### gRPC

- Proto files live in `packages/proto/`
- Follow [Google API Design Guide](https://cloud.google.com/apis/design)
- Generated code is committed (for discoverability); regenerate on proto change

### WebSocket

- Events follow: `{ "type": "[NOUN].[VERB]", "payload": {...}, "timestamp": "..." }`
- Authentication via initial handshake, not per-message

---

## Testing

Backend services are required to have unit and E2E tests. Frontend apps are exempt.

### TypeScript

```
tests/
├── unit/        # vitest — test domain logic in isolation
│   └── *.test.ts
└── e2e/         # supertest — test full HTTP handler stack
    └── *.e2e.ts
```

```typescript
// vitest unit test
import { describe, it, expect } from "vitest";
import { calculateTax } from "../src/domain/pricing";

describe("calculateTax", () => {
  it("applies 10% to the subtotal", () => {
    expect(calculateTax(100, "US")).toBe(10);
  });
});

// supertest e2e test
import request from "supertest";
import { app } from "../src/app";

describe("POST /v1/users", () => {
  it("creates a user and returns 201", async () => {
    const res = await request(app)
      .post("/v1/users")
      .send({ email: "test@example.com", name: "Alice" });
    expect(res.status).toBe(201);
    expect(res.body.email).toBe("test@example.com");
  });
});
```

### Python

```
tests/
├── unit/        # pytest — pure unit tests, no I/O
│   └── test_*.py
└── e2e/         # pytest + httpx — full HTTP stack
    └── test_*.py
```

```python
# pytest unit test
from src.domain.pricing import calculate_tax

def test_calculates_tax_for_us():
    assert calculate_tax(100, "US") == 10

# httpx e2e test
import pytest
import httpx
from src.main import app

@pytest.mark.anyio
async def test_create_user():
    async with httpx.AsyncClient(app=app, base_url="http://test") as client:
        response = await client.post("/v1/users", json={"email": "test@example.com", "name": "Alice"})
    assert response.status_code == 201
```

### Go

```
tests/
├── unit/        # standard testing + testify
└── e2e/         # testcontainers-go for DB/infra
```

```go
// Unit test (in same package as source, or _test package)
func TestCalculateTax(t *testing.T) {
    got := CalculateTax(100, "US")
    assert.Equal(t, 10.0, got)
}

// Table-driven tests (preferred in Go)
func TestCalculateTax(t *testing.T) {
    tests := []struct {
        name     string
        subtotal float64
        country  string
        want     float64
    }{
        {"US tax", 100, "US", 10},
        {"EU tax", 100, "DE", 19},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.want, CalculateTax(tt.subtotal, tt.country))
        })
    }
}
```

### Rust

```
# Unit tests: inline in src/ using #[cfg(test)]
# E2E tests: tests/ directory at crate root
```

```rust
// Unit test (inline)
fn calculate_tax(subtotal: f64, country: &str) -> f64 { ... }

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn calculates_us_tax() {
        assert_eq!(calculate_tax(100.0, "US"), 10.0);
    }
}

// E2E test (tests/api_test.rs)
#[tokio::test]
async fn creates_user_returns_201() {
    let app = create_test_app().await;
    let response = app.oneshot(
        Request::builder()
            .method("POST")
            .uri("/v1/users")
            .header("content-type", "application/json")
            .body(Body::from(r#"{"email":"test@example.com","name":"Alice"}"#))
            .unwrap()
    ).await.unwrap();
    assert_eq!(response.status(), StatusCode::CREATED);
}
```

---

## Schema / Protobuf

> All data models are defined in `packages/schema/proto/`. Read `packages/schema/README.md` first.

### Package Naming

```protobuf
// Format: [service-name].v[version]
package payments.v1;
package auth.v1;
package notifications.v1;

// Common shared types live in:
package common.v1;
```

### Message Naming

```protobuf
// Entities (DB models): noun
message User { ... }
message Payment { ... }

// View models (API responses): noun + descriptor
message UserPublic { ... }      // safe for external consumers
message UserInternal { ... }    // includes sensitive fields, internal only

// Requests: verb + noun + Request
message CreateUserRequest { ... }
message GetUserRequest { ... }
message UpdatePaymentStatusRequest { ... }

// Responses: verb + noun + Response (for list operations)
message GetUsersResponse { ... }

// Events: noun + past-tense verb + Event
message UserCreatedEvent { ... }
message PaymentCompletedEvent { ... }
message OrderStatusChangedEvent { ... }
```

### Enum Naming

```protobuf
// Enum type: PascalCase
// Values: UPPER_SNAKE_CASE with type prefix (buf enforces this)
enum UserStatus {
  USER_STATUS_UNSPECIFIED = 0;  // required: zero value is always UNSPECIFIED
  USER_STATUS_ACTIVE = 1;
  USER_STATUS_INACTIVE = 2;
}
```

### Field Numbering

- Fields 1–15 use 1-byte tags (use for the most common fields: id, status, timestamps)
- Fields 16–2047 use 2-byte tags
- Never reuse a field number, even after removing a field
- Mark removed fields as `reserved`: `reserved 5; reserved "old_field_name";`

### Importing Common Types

```protobuf
import "common/v1/pagination.proto";
import "common/v1/errors.proto";
import "google/protobuf/timestamp.proto";

// google/protobuf types available:
// timestamp.proto  → google.protobuf.Timestamp
// wrappers.proto   → google.protobuf.StringValue (nullable strings)
// struct.proto     → google.protobuf.Struct (arbitrary JSON)
```

### Workflow: Adding a New Type

1. Create/edit `packages/schema/proto/[service]/v1/[resource].proto`
2. Run `cd packages/schema && ./scripts/generate.sh`
3. Commit proto + generated files: `feat(schema): add Payment type`
4. Import the generated type in your service
5. Map to ORM/DB schema in the service's `infra/` layer

### Using Generated Types in ORM/DB Layers

**TypeScript (Drizzle example):**

```typescript
import { UserStatus } from "@schema/example/v1/user_pb";

// Map proto enum → DB column
export const usersTable = pgTable("users", {
  id: uuid("id").primaryKey().defaultRandom(),
  email: varchar("email", { length: 255 }).notNull().unique(),
  status: varchar("status", { length: 50 })
    .$type<UserStatus>()
    .notNull()
    .default(UserStatus[UserStatus.USER_STATUS_ACTIVE]),
});
```

**Python (SQLAlchemy example):**

```python
from schema.example.v1 import UserStatus
from sqlalchemy import Column, String, Enum as SAEnum
import enum

# Convert betterproto enum to Python enum for SQLAlchemy
class DBUserStatus(str, enum.Enum):
    ACTIVE = "USER_STATUS_ACTIVE"
    INACTIVE = "USER_STATUS_INACTIVE"

class UserModel(Base):
    __tablename__ = "users"
    id = Column(String, primary_key=True)
    email = Column(String, nullable=False, unique=True)
    status = Column(SAEnum(DBUserStatus), nullable=False, default=DBUserStatus.ACTIVE)
```

**Go (sqlx/GORM example):**

```go
import userv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/example/v1"

type UserRow struct {
    ID     string `db:"id"`
    Email  string `db:"email"`
    Status string `db:"status"`  // store enum name as string
}

// Convert: proto → DB
func toDB(u *userv1.User) UserRow {
    return UserRow{
        ID:     u.Id,
        Email:  u.Email,
        Status: u.Status.String(),
    }
}

// Convert: DB → proto
func toProto(row UserRow) *userv1.User {
    status := userv1.UserStatus_value[row.Status]
    return &userv1.User{
        Id:     row.ID,
        Email:  row.Email,
        Status: userv1.UserStatus(status),
    }
}
```

---

## Authentication & Middleware

> Every backend service endpoint must be protected by auth middleware. No exceptions except `/health` and `/metrics`.
> See `CORE_RULES.md → Rule 10` for the policy. This section covers per-language implementation.

### How it works

Every inbound request is checked for **one of two credentials** in this order:

1. `X-API-Key: <value>` — matches `API_KEY` env var → service-to-service identity, no user context
2. `Authorization: Bearer <jwt>` — verified with `JWT_SECRET` → user or service JWT with full claims

On success: the resolved `AuthClaims` is attached to the request context and available to all handlers.
On failure: `401 Unauthorized` with `common.v1.ErrorResponse` body.

### TypeScript (Express / Fastify)

```typescript
// src/middleware/auth.ts
import type { Request, Response, NextFunction } from "express";
import jwt from "jsonwebtoken";
import { env } from "../config/env";

// Extend express Request to carry resolved auth
declare global {
  namespace Express {
    interface Request {
      auth: {
        subject: string;
        method: string;
        role?: string;
        scopes?: string[];
      };
    }
  }
}

export function authMiddleware(
  req: Request,
  res: Response,
  next: NextFunction,
) {
  // 1. API Key — BE↔BE without user context
  const apiKey = req.headers["x-api-key"];
  if (typeof apiKey === "string" && apiKey === env.API_KEY) {
    req.auth = { subject: "service", method: "api-key" };
    return next();
  }

  // 2. JWT — FE→BE or BE↔BE with user/service token
  const authHeader = req.headers.authorization;
  if (authHeader?.startsWith("Bearer ")) {
    try {
      const claims = jwt.verify(
        authHeader.slice(7),
        env.JWT_SECRET,
      ) as jwt.JwtPayload;
      req.auth = {
        subject: claims.sub ?? "",
        method: "jwt",
        role: claims.role,
        scopes: claims.scopes,
      };
      return next();
    } catch {
      /* fall through */
    }
  }

  res.status(401).json({
    code: "UNAUTHORIZED",
    message: "Valid API key or Bearer token required",
    fieldErrors: [],
  });
}

// Wire up in app.ts:
//   app.get('/health', healthHandler)    ← BEFORE authMiddleware
//   app.get('/metrics', metricsHandler)  ← BEFORE authMiddleware
//   app.use(authMiddleware)              ← all routes after this are protected
```

```typescript
// src/config/env.ts — validate at startup, fail fast if missing
import { z } from "zod";
export const env = z
  .object({
    JWT_SECRET: z.string().min(32, "JWT_SECRET must be ≥32 chars"),
    JWT_EXPIRY_SECONDS: z.coerce.number().default(3600),
    API_KEY: z.string().min(16, "API_KEY must be ≥16 chars"),
  })
  .parse(process.env);
```

### Python (FastAPI dependency)

```python
# src/middleware/auth.py
from fastapi import HTTPException, Security, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer, APIKeyHeader
import jwt
from src.config import settings

_bearer = HTTPBearer(auto_error=False)
_api_key_header = APIKeyHeader(name="X-API-Key", auto_error=False)

async def require_auth(
    api_key: str | None = Security(_api_key_header),
    credentials: HTTPAuthorizationCredentials | None = Security(_bearer),
) -> dict:
    """FastAPI dependency — inject into any route or router."""
    if api_key and api_key == settings.api_key:
        return {"subject": "service", "method": "api-key"}

    if credentials:
        try:
            payload = jwt.decode(credentials.credentials, settings.jwt_secret, algorithms=["HS256"])
            return {"subject": payload.get("sub", ""), "method": "jwt",
                    "role": payload.get("role"), "scopes": payload.get("scopes", [])}
        except jwt.InvalidTokenError:
            pass

    raise HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail={"code": "UNAUTHORIZED", "message": "Valid API key or Bearer token required"},
    )
```

```python
# src/config.py
from pydantic import Field
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    jwt_secret: str = Field(min_length=32)
    jwt_expiry_seconds: int = 3600
    api_key: str = Field(min_length=16)

settings = Settings()
```

```python
# Apply to all routes via router dependency:
from fastapi import APIRouter, Depends
from src.middleware.auth import require_auth

router = APIRouter(dependencies=[Depends(require_auth)])

# Health/metrics — register on the app directly (no dependency):
# app.get("/health")(health_handler)
# app.get("/metrics")(metrics_handler)
```

### Go (net/http middleware)

```go
// internal/middleware/auth.go
package middleware

import (
    "context"
    "encoding/json"
    "net/http"
    "strings"
    "github.com/golang-jwt/jwt/v5"
)

type contextKey string
const AuthKey contextKey = "auth"

type AuthClaims struct {
    Subject string
    Method  string // "jwt" | "api-key"
    Role    string
    Scopes  []string
}

func Auth(jwtSecret, apiKey string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. API Key
            if key := r.Header.Get("X-API-Key"); key == apiKey {
                ctx := context.WithValue(r.Context(), AuthKey, &AuthClaims{Subject: "service", Method: "api-key"})
                next.ServeHTTP(w, r.WithContext(ctx))
                return
            }
            // 2. JWT
            if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
                token, err := jwt.Parse(strings.TrimPrefix(auth, "Bearer "),
                    func(t *jwt.Token) (any, error) { return []byte(jwtSecret), nil },
                    jwt.WithValidMethods([]string{"HS256"}),
                )
                if err == nil {
                    if mc, ok := token.Claims.(jwt.MapClaims); ok {
                        claims := &AuthClaims{Subject: fmt.Sprint(mc["sub"]), Method: "jwt"}
                        next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), AuthKey, claims)))
                        return
                    }
                }
            }
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(map[string]any{"code": "UNAUTHORIZED", "message": "Valid API key or Bearer token required"})
        })
    }
}

// GetAuth retrieves claims from context inside handlers.
func GetAuth(ctx context.Context) *AuthClaims {
    v, _ := ctx.Value(AuthKey).(*AuthClaims)
    return v
}
```

```go
// cmd/server/main.go
mux := http.NewServeMux()
mux.HandleFunc("GET /health", healthHandler)   // exempt — registered first
mux.HandleFunc("GET /metrics", metricsHandler) // exempt
mux.Handle("/", middleware.Auth(cfg.JWTSecret, cfg.APIKey)(apiRouter)) // protected
```

### Rust (axum extractor)

```rust
// src/middleware/auth.rs
use axum::{extract::FromRequestParts, http::{request::Parts, StatusCode}, Json};
use jsonwebtoken::{decode, DecodingKey, Validation};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JwtClaims { pub sub: String, pub role: Option<String>, pub scopes: Option<Vec<String>> }

pub struct AuthClaims { pub subject: String, pub method: &'static str }

#[axum::async_trait]
impl<S: Send + Sync> FromRequestParts<S> for AuthClaims {
    type Rejection = (StatusCode, Json<serde_json::Value>);

    async fn from_request_parts(parts: &mut Parts, _: &S) -> Result<Self, Self::Rejection> {
        let cfg = parts.extensions.get::<crate::config::Config>().expect("Config missing");

        if let Some(key) = parts.headers.get("x-api-key") {
            if key.to_str().ok() == Some(&cfg.api_key) {
                return Ok(AuthClaims { subject: "service".into(), method: "api-key" });
            }
        }
        if let Some(auth) = parts.headers.get("authorization") {
            if let Some(token) = auth.to_str().ok().and_then(|s| s.strip_prefix("Bearer ")) {
                let key = DecodingKey::from_secret(cfg.jwt_secret.as_bytes());
                if let Ok(data) = decode::<JwtClaims>(token, &key, &Validation::default()) {
                    return Ok(AuthClaims { subject: data.claims.sub, method: "jwt" });
                }
            }
        }
        Err((StatusCode::UNAUTHORIZED, Json(serde_json::json!({
            "code": "UNAUTHORIZED", "message": "Valid API key or Bearer token required"
        }))))
    }
}
// Usage in handlers: async fn list_users(auth: AuthClaims, ...) { ... }
// Health/metrics routes: don't include AuthClaims extractor in the handler signature
```

### Generating service-to-service tokens

When a BE service calls another BE service with user context (not just API key):

```typescript
// TypeScript
import jwt from "jsonwebtoken";
import { env } from "../config/env";

export function signServiceToken(
  sub: string,
  role?: string,
  scopes?: string[],
): string {
  return jwt.sign({ sub, type: "service", role, scopes }, env.JWT_SECRET, {
    expiresIn: env.JWT_EXPIRY_SECONDS,
  });
}
```

```python
# Python
import jwt
from src.config import settings

def sign_service_token(sub: str, role: str = "", scopes: list[str] | None = None) -> str:
    return jwt.encode(
        {"sub": sub, "type": "service", "role": role, "scopes": scopes or []},
        settings.jwt_secret, algorithm="HS256",
    )
```

### Exempt paths (always unauthenticated)

| Path       | Method | Reason                         |
| ---------- | ------ | ------------------------------ |
| `/health`  | GET    | Container orchestration probes |
| `/metrics` | GET    | Prometheus scraper             |

Everything else → auth required.

---

## AI / Agentic Services

> Agentic services are backend services whose primary logic involves orchestrating LLM calls, tool use, memory, and multi-step reasoning. They follow all standard service conventions PLUS the rules below.

### Recommended Libraries

| Language   | Framework                      | Purpose                                                  |
| ---------- | ------------------------------ | -------------------------------------------------------- |
| TypeScript | [Mastra AI](https://mastra.ai) | Agent orchestration, tools, workflows, memory, RAG       |
| Python     | [Agno](https://docs.agno.com)  | Agent orchestration, tools, structured output, knowledge |

### TypeScript — Mastra AI

```typescript
import { Agent, createTool } from "@mastra/core";
import { z } from "zod";
// Schema types always come from packages/schema
import { User } from "@schema/example/v1/user_pb";

// Tools: input/output schemas defined with zod, underlying types from proto
const getUserTool = createTool({
  id: "get-user",
  description: "Fetch a user by ID",
  inputSchema: z.object({ userId: z.string().uuid() }),
  outputSchema: z.object({ user: z.custom<User>() }),
  execute: async ({ context }) => {
    const user = await userRepository.findById(context.userId);
    return { user };
  },
});

// Agent definition
export const myAgent = new Agent({
  name: "my-agent",
  instructions: `You are a helpful assistant. Always use the provided tools.`,
  model: {
    provider: "ANTHROPIC",
    name: "claude-sonnet-4-6",
  },
  tools: { getUserTool },
});
```

**Mastra conventions:**

- One agent per file in `src/agents/`
- Tools live in `src/tools/` — one file per tool domain
- Workflows (multi-step, branching) live in `src/workflows/`
- Memory and storage configured in `src/mastra.ts` (the Mastra instance)
- All tool schemas use `zod` for runtime validation
- Never hardcode model names — use `env.MODEL_NAME` from config

### Python — Agno

```python
from agno.agent import Agent
from agno.models.anthropic import Claude
from agno.tools import tool
from pydantic import BaseModel
# Schema types come from packages/schema generated output
from schema.example.v1 import User, UserStatus

# Tool input/output as Pydantic models (sourced from or extending proto types)
class GetUserInput(BaseModel):
    user_id: str

# Tool definition
@tool(description="Fetch a user by ID")
async def get_user(user_id: str) -> User:
    return await user_repository.find_by_id(user_id)

# Agent definition
agent = Agent(
    name="my-agent",
    model=Claude(id="claude-sonnet-4-6"),
    tools=[get_user],
    instructions="You are a helpful assistant. Always use the provided tools.",
    markdown=True,
    # Optional: structured output enforced
    response_model=User,
)
```

**Agno conventions:**

- One agent per file in `src/agents/`
- Tools live in `src/tools/` — one file per tool domain
- Agent teams (multi-agent) defined in `src/teams/`
- Knowledge bases and storage configured in `src/knowledge/`
- Tool parameters use Pydantic models — extend or import from `packages/schema` generated types
- Never hardcode model IDs — use `settings.model_name` from config

### Cross-Language Rules for Agentic Services

1. **Schema-first applies equally** — all data types passed to/from tools must be in `packages/schema`
2. **Structured output** — agents must return typed, structured responses (never raw strings for data)
3. **Tool idempotency** — tools must be safe to retry; avoid side effects that can't be rolled back
4. **Observability** — every agent run must emit a trace ID; log tool calls and LLM responses at DEBUG level
5. **Model config** — model name, provider, and parameters come from environment variables, not hardcoded
6. **No direct LLM calls** — always go through the framework (Mastra / Agno); never call the Anthropic SDK directly from business logic
7. **Cost guardrails** — set `max_tokens` and timeout limits on every agent invocation
8. **Human-in-the-loop** — any agent action that is irreversible (send email, delete data, charge card) must have an explicit approval step

---

## Logging and Observability

- Structured logging only (JSON in production, pretty in development)
- Log levels: `DEBUG` (dev only), `INFO` (normal operations), `WARN` (unexpected but handled), `ERROR` (needs attention)
- Never log secrets, PII, or full request bodies
- Use trace IDs for request correlation across services
- Every service exposes `/health` and `/metrics` endpoints

---

## Docker

- Multi-stage Dockerfiles: `dev` target (with hot reload) and `prod` target (minimal image)
- Base images: use official images, pin to minor versions (e.g., `node:22-alpine`, not `node:latest`)
- Non-root user in production containers
- `.dockerignore` excludes `node_modules`, `.env`, `dist`, `__pycache__`, etc.
