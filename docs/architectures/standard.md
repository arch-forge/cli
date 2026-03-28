# Standard Layout

## Overview

Standard Layout follows the community-accepted Go project layout with `cmd/`, `internal/`, and optionally `pkg/` at the top level. The internal structure uses `model`, `service`, `repository`, and `handler` packages — familiar naming that any Go developer can navigate immediately. There are no strict architectural rules about dependency direction; services may depend on repositories directly and handlers may depend on services. This pattern is ideal for smaller applications, rapid prototypes, internal tooling, or teams new to Go who want a conventional starting point before adopting a more structured architecture.

## Quick Start

```bash
arch_forge init myapp --arch standard --variant classic
arch_forge init myapp --arch standard --variant modular
```

## Generated Structure — Classic Variant

```
myapp/
├── cmd/
│   └── app/
│       └── main.go                       # Entry point; wires service and handler
├── internal/
│   ├── model/
│   │   └── user.go                       # Plain Go struct representing a User
│   ├── service/
│   │   └── user_service.go               # UserService interface + implementation
│   ├── repository/
│   │   └── user_repo.go                  # UserRepository interface + implementation
│   └── handler/
│       └── user_handler.go               # HTTP handler wired to user service
├── go.mod
├── .gitignore
└── archforge.yaml
```

## Generated Structure — Modular Variant

The modular variant organizes code by domain area and separates shared platform concerns.

```
myapp/
├── cmd/
│   └── app/
│       └── main.go               # Entry point
├── internal/
│   ├── <domain>/                 # One per domain area (e.g. users, orders)
│   │   ├── model/                # Structs scoped to this domain
│   │   ├── service/              # Service logic for this domain
│   │   ├── repository/           # Persistence for this domain
│   │   └── handler/              # HTTP handlers for this domain
│   └── platform/
│       └── config/
│           └── config.go         # Config loaded from environment variables
├── go.mod
├── .gitignore
└── archforge.yaml
```

## Layer Responsibilities

| Layer / Package | Belongs Here | Does NOT Belong Here |
|---|---|---|
| `internal/model` | Plain Go structs representing business entities | Business logic, HTTP types, SQL queries |
| `internal/service` | Business logic, service interface and implementation | SQL queries, HTTP request parsing |
| `internal/repository` | Repository interface and data-access implementation | Business logic, HTTP types |
| `internal/handler` | HTTP route registration, request parsing, response writing | Business logic, direct DB access |

## Compatible Modules

| Module | Purpose | Notes |
|---|---|---|
| `api` | REST router, middleware chain, request validation | Supports `chi` and `stdlib`; patches `cmd/*/main.go` via `arch_forge:routes` |
| `database` | Postgres connection pool and migrations | Patches `cmd/*/main.go` via `arch_forge:providers` |
| `auth` | JWT middleware and token management | Requires `api` and `logging` |
| `logging` | Structured `slog` logging with correlation IDs | Patches `cmd/*/main.go` via `arch_forge:providers` |
| `metrics` | Prometheus metrics and `/metrics` endpoint | Requires `api` |
| `tracing` | OpenTelemetry distributed tracing | Optional dependency on `logging` |
| `healthcheck` | `/health`, `/ready`, `/live` endpoints | Requires `api` |
| `cache` | Redis cache-aside client | Requires `logging` |
| `queue` | Message queue (RabbitMQ/NATS/in-memory) | Optional dependency on `logging` |
| `grpc` | gRPC server | Requires `logging` |
| `crud` | Full CRUD scaffold for an entity | Requires `api` and `database` |
| `cors` | CORS middleware | Requires `api` |
| `ratelimit` | Token-bucket rate limiting per IP | Requires `api` |
| `encryption` | AES-GCM encryption and key derivation | No dependencies |
| `storage` | Object storage client (S3/GCS/local) | No dependencies |
| `docker` | Multi-stage Dockerfile and docker-compose | Optional dependency on `database` |
| `ci` | GitHub Actions / GitLab CI pipelines | Optional dependency on `docker` |
| `k8s` | Kubernetes manifests | Requires `docker` |
| `testkit` | Test fixtures, factories, testcontainers | Optional dependency on `database` |
| `e2e` | End-to-end HTTP test scaffold | Requires `api`; optional `testkit` |
| `mocks` | mockery-generated test doubles from interfaces | No dependencies |
| `makefile` | Standard `make build/test/lint/run/migrate` targets | No dependencies |

## Architecture Rules (doctor checks)

Standard Layout applies a minimal set of checks to prevent the most common pitfalls:

- `internal/model` does not import `internal/service`, `internal/repository`, or `internal/handler`.
- `internal/repository` does not import `internal/handler`.
- `internal/service` does not import `internal/handler`.
- No circular imports between any `internal/` packages.

## Examples

### Adding a new endpoint (e.g., `CreateUser`)

1. Add a `CreateUser` method to the `UserService` interface in `internal/service/user_service.go`:
   ```go
   CreateUser(ctx context.Context, user *model.User) error
   ```
2. Implement the method in the `userService` struct.
3. Add a `Save` method to the `UserRepository` interface in `internal/repository/user_repo.go` and implement it.
4. Register the POST route in `internal/handler/user_handler.go`:
   ```go
   mux.HandleFunc("POST /users", h.createUser)
   ```

### Adding a module (e.g., `logging`)

```bash
arch_forge add logging
```

This creates the `slog`-based logger package and injects initialization code into `cmd/app/main.go` at the `// arch_forge:providers` anchor.

### Using the `crud` scaffold for a new entity

```bash
arch_forge add crud --option entity_name=Product --option table_name=products
```

This generates `internal/domain/product.go`, a migration file in `migrations/`, and wires up handler, service, and repository placeholders automatically.

## Migration from Standard Layout

Standard Layout is the starting point — there is no migration to it. To migrate _from_ Standard Layout to a more structured architecture:

- To **Hexagonal**: see [hexagonal.md](./hexagonal.md#migration-from-standard-layout)
- To **Clean Architecture**: see [clean.md](./clean.md#migration-from-standard-layout)
- To **DDD**: see [ddd.md](./ddd.md#migration-from-standard-layout)
- To **Microservice**: see [microservice.md](./microservice.md#migration-from-standard-layout)
- To **Modular Monolith**: see [modular_monolith.md](./modular_monolith.md#migration-from-standard-layout)
