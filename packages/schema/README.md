# packages/schema

The canonical source of truth for **all data models** across every service in this monorepo.

> **Schema-First Mandate**: Every data type — DB entities, API request/response shapes, enums, event payloads, and internal domain objects — MUST be defined here as a Protobuf message **before** any service code is written.

---

## Why Proto?

| Problem                                               | How proto solves it                        |
| ----------------------------------------------------- | ------------------------------------------ |
| Drift between services (different field names, types) | One definition, generated everywhere       |
| DB model ≠ API response ≠ event payload               | Derive all from the same message           |
| Adding a field breaks a consumer silently             | Breaking change detection (`buf breaking`) |
| "What shape does this event have?"                    | Read the `.proto` file                     |
| Type safety across TS/Go/Python/Rust                  | Generated types in each language           |

---

## Directory Structure

```
packages/schema/
├── buf.yaml                    # Buf workspace config (lint + breaking change rules)
├── buf.gen.yaml                # Code generation config
├── proto/                      # Source of truth — ONLY edit .proto files
│   ├── common/
│   │   └── v1/
│   │       ├── errors.proto    # ErrorResponse, FieldError
│   │       └── pagination.proto # PaginationRequest, PaginationMeta, SortOrder
│   └── [service-name]/         # One directory per service
│       └── v1/
│           └── [resource].proto
├── generated/                  # Auto-generated — DO NOT EDIT MANUALLY
│   ├── ts/                     # TypeScript (@bufbuild/protoc-gen-es)
│   ├── go/                     # Go (protoc-gen-go + protoc-gen-go-grpc)
│   └── python/                 # Python Pydantic-compatible (betterproto)
├── scripts/
│   └── generate.sh             # Run this after any .proto change
└── examples/
    └── rust-build.rs           # Copy to your Rust service as build.rs
```

---

## The Mandate: What Goes in Proto

**All of these must be proto-first:**

| What                      | Example                                                           |
| ------------------------- | ----------------------------------------------------------------- |
| DB entity                 | `message User { string id = 1; string email = 2; ... }`           |
| View model (API response) | `message UserPublic { ... }` (omits password_hash)                |
| API request body          | `message CreateUserRequest { ... }`                               |
| API response body         | `message GetUsersResponse { repeated UserPublic users = 1; ... }` |
| Enum / status             | `enum UserStatus { ACTIVE = 1; INACTIVE = 2; }`                   |
| Event payload             | `message UserCreatedEvent { string user_id = 1; ... }`            |
| Internal domain object    | `message UserDraft { ... }`                                       |

**What does NOT need to be in proto:**

- Framework configuration objects
- Infrastructure secrets/config
- Test fixtures (use the generated types directly)

---

## How to Add a New Type

### 1. Create the proto file

```bash
# For a new service "payments":
mkdir -p packages/schema/proto/payments/v1
touch packages/schema/proto/payments/v1/payment.proto
```

### 2. Write the proto definition

```protobuf
syntax = "proto3";

package payments.v1;

import "common/v1/pagination.proto";
import "google/protobuf/timestamp.proto";

option go_package = "github.com/agarwalvivek29/here.now/packages/schema/generated/go/payments/v1;paymentsv1";

enum PaymentStatus {
  PAYMENT_STATUS_UNSPECIFIED = 0;
  PAYMENT_STATUS_PENDING = 1;
  PAYMENT_STATUS_COMPLETED = 2;
  PAYMENT_STATUS_FAILED = 3;
}

message Payment {
  string id = 1;
  string user_id = 2;
  int64 amount_cents = 3;
  string currency = 4;
  PaymentStatus status = 5;
  google.protobuf.Timestamp created_at = 6;
}

message CreatePaymentRequest {
  string user_id = 1;
  int64 amount_cents = 2;
  string currency = 3;
}

message PaymentCompletedEvent {
  string payment_id = 1;
  string user_id = 2;
  int64 amount_cents = 3;
  google.protobuf.Timestamp occurred_at = 4;
}
```

### 3. Generate code

```bash
cd packages/schema
./scripts/generate.sh
```

### 4. Commit the generated code

```bash
git add packages/schema/
git commit -m "feat(schema): add Payment type and PaymentCompletedEvent"
```

---

## Using Generated Types

### TypeScript (services + frontend apps)

```typescript
// tsconfig.json — add path alias
// "@schema/*": ["../../packages/schema/generated/ts/*"]

import { User, UserStatus, CreateUserRequest } from '@schema/example/v1/user_pb'

// Use in route handler:
const req = new CreateUserRequest({ email: 'user@example.com', name: 'Alice' })

// Use as type:
function processUser(user: User): void { ... }
```

### Go

```go
import (
    userv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/example/v1"
    commonv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/common/v1"
)

// Use generated struct
user := &userv1.User{
    Id:    uuid.New().String(),
    Email: "user@example.com",
    Status: userv1.UserStatus_USER_STATUS_ACTIVE,
}

// Use generated enum in DB mapping
func statusToDB(s userv1.UserStatus) string {
    return s.String()
}
```

### Python (Pydantic-compatible via betterproto)

```python
# betterproto generates dataclasses that behave like Pydantic models
from schema.example.v1 import User, UserStatus, UserRole

# Works as a Pydantic model:
user = User(
    id="abc-123",
    email="user@example.com",
    status=UserStatus.USER_STATUS_ACTIVE,
)

# Serialize to dict (for DB or JSON):
user_dict = user.to_dict()

# Use as FastAPI request body:
from fastapi import FastAPI
app = FastAPI()

@app.post("/users")
async def create_user(body: CreateUserRequest) -> User:
    ...
```

### Rust (via build.rs + prost)

```rust
// In your service, add build.rs (copy from examples/rust-build.rs)
// Then in your code:

mod proto {
    include!(concat!(env!("OUT_DIR"), "/example.v1.rs"));
}

use proto::{User, UserStatus};

fn process_user(user: User) {
    if user.status == UserStatus::Active as i32 {
        // ...
    }
}
```

---

## Proto Conventions

### Naming

- **Package**: `[service-name].v1` (always versioned)
- **Messages**: `PascalCase` (e.g., `UserCreatedEvent`, not `UserEvent`)
- **Enums**: `PascalCase`, values in `UPPER_SNAKE_CASE` with type prefix (e.g., `USER_STATUS_ACTIVE`)
- **Fields**: `snake_case`
- **Services**: `PascalCase` + `Service` suffix (e.g., `UserService`)
- **RPCs**: `PascalCase` verb-noun (e.g., `CreateUser`, `GetUsers`, `UpdateUserStatus`)

### Breaking Changes

Avoid these — they break generated consumers in other services:

- Removing a field
- Changing a field's number
- Changing a field's type
- Removing an enum value
- Renaming a message or enum (use `option deprecated = true` first)

Safe changes:

- Adding new fields (always optional in proto3)
- Adding new enum values (add at end)
- Adding new messages
- Adding new RPC methods

### Field Numbering

- 1–15: frequently-used fields (1-byte tag encoding — slightly smaller on wire)
- 16–2047: less-used fields
- Never reuse a field number, even after removing a field

---

## Tooling Setup

```bash
# Install buf (macOS)
brew install bufbuild/buf/buf

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Install Python betterproto
pip install "betterproto[compiler]"

# Rust: prost-build is a build dependency — no system install needed
# Add to Cargo.toml [build-dependencies]: prost-build = "0.13"
```

---

## CI

`schema.yml` runs on every PR:

- `buf lint` — validates proto style
- `buf breaking` — detects breaking changes vs main (PRs only)
- Regenerates code and checks git diff — fails if generated code is stale

If CI fails because generated code is stale: run `./scripts/generate.sh` and commit.
