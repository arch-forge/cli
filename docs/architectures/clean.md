# Clean Architecture

## Overview

Clean Architecture, popularized by Robert C. Martin, organizes code into concentric dependency rings: entities at the center, use cases around them, interface adapters in the next ring, and frameworks on the outside. The Dependency Rule states that source code dependencies can only point inward — outer layers know about inner layers, never the reverse. In arch_forge's implementation, enterprise business rules live in `internal/entity`, application business rules in `internal/usecase`, interface adapters in `internal/controller` and `internal/gateway`, and framework/driver code at the outermost ring. Choose Clean Architecture when you need clear enforcement of the Dependency Rule across a team, or when your application logic must be exercised entirely without a database or HTTP server in tests.

## Quick Start

```bash
arch_forge init myapp --arch clean --variant classic
arch_forge init myapp --arch clean --variant modular
```

## Generated Structure — Classic Variant

```
myapp/
├── cmd/
│   └── api/
│       └── main.go                       # Entry point; constructs and wires all layers
├── internal/
│   ├── entity/
│   │   └── user.go                       # Enterprise business rules — pure structs + Validate()
│   ├── usecase/
│   │   ├── port/
│   │   │   └── user_repo.go              # Repository interface (application-level port)
│   │   └── get_user.go                   # Use case struct with Execute() method
│   ├── controller/
│   │   └── http/
│   │       └── user_controller.go        # HTTP controller; adapts HTTP to use case input
│   └── gateway/
│       └── postgres/
│           └── user_gateway.go           # Postgres implementation of usecase/port.UserRepository
├── migrations/                           # SQL migration files
├── go.mod
├── .gitignore
└── archforge.yaml
```

## Generated Structure — Modular Variant

Each business domain is a self-contained directory under `internal/`. The shared infrastructure (database helpers, config) lives in `internal/shared/`.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go               # Entry point; bootstraps all domain modules
├── internal/
│   ├── <module>/                 # One per bounded domain (e.g. users, billing)
│   │   ├── domain/               # Entities scoped to this module
│   │   ├── port/                 # Use-case-level port interfaces
│   │   ├── app/                  # Use cases for this module
│   │   └── adapter/              # Controllers and gateways for this module
│   └── shared/
│       └── framework/
│           └── database.go       # Shared database connection and pool
├── go.mod
├── .gitignore
└── archforge.yaml
```

## Layer Responsibilities

| Layer / Package | Belongs Here | Does NOT Belong Here |
|---|---|---|
| `internal/entity` | Enterprise business rules; pure Go structs; `Validate()` methods | HTTP types, SQL, framework imports, use-case orchestration |
| `internal/usecase/port` | Repository and service interfaces used by use cases | Implementations, SQL, HTTP |
| `internal/usecase` | Application business logic; one struct per use case with `Execute()` | HTTP parsing, JSON, SQL, framework code |
| `internal/controller/http` | HTTP request parsing, response writing, routing | Business logic, direct DB access |
| `internal/gateway/postgres` | SQL queries, ORM calls implementing `usecase/port` interfaces | Business rules, HTTP types |

## Compatible Modules

| Module | Purpose | Notes |
|---|---|---|
| `api` | REST router, middleware chain, request validation | Supports `chi` and `stdlib`; patches `cmd/*/main.go` via `arch_forge:routes` |
| `database` | Postgres connection pool and migrations | Patches `cmd/*/main.go` via `arch_forge:providers` |
| `auth` | JWT middleware and token management | Requires `api` and `logging` |
| `logging` | Structured `slog` logging with correlation IDs | Patches `cmd/*/main.go` via `arch_forge:providers` |
| `metrics` | Prometheus counters, histograms, `/metrics` endpoint | Requires `api` |
| `tracing` | OpenTelemetry distributed tracing with OTLP export | Optional dependency on `logging` |
| `healthcheck` | `/health`, `/ready`, `/live` endpoints | Requires `api` |
| `cache` | Redis cache-aside client with TTL management | Requires `logging` |
| `queue` | RabbitMQ/NATS/in-memory message queue | Optional dependency on `logging` |
| `grpc` | gRPC server with buf config and interceptors | Requires `logging` |
| `crud` | Full CRUD scaffold for an entity | Requires `api` and `database` |
| `cors` | CORS middleware | Requires `api` |
| `ratelimit` | Token-bucket rate limiting per IP | Requires `api` |
| `docker` | Multi-stage Dockerfile and docker-compose | Optional dependency on `database` |
| `ci` | GitHub Actions / GitLab CI pipelines | Optional dependency on `docker` |
| `k8s` | Kubernetes manifests | Requires `docker` |
| `testkit` | Test fixtures, factories, testcontainers | Optional dependency on `database` |
| `e2e` | End-to-end HTTP test scaffold | Requires `api`; optional `testkit` |
| `mocks` | mockery-generated test doubles from interfaces | No dependencies |
| `makefile` | Standard `make build/test/lint/run/migrate` targets | No dependencies |

## Architecture Rules (doctor checks)

`arch_forge doctor` validates the following for Clean Architecture projects:

- `internal/entity` imports no `internal/*` packages; it is dependency-free.
- `internal/usecase/port` imports only `internal/entity`.
- `internal/usecase` imports `internal/entity` and `internal/usecase/port` only; never `internal/controller` or `internal/gateway`.
- `internal/controller` imports `internal/usecase` (use case structs or port interfaces) but not `internal/gateway`.
- `internal/gateway` imports `internal/entity` and `internal/usecase/port` only; never `internal/usecase` directly.
- No database packages appear in `internal/entity` or `internal/usecase`.
- No HTTP packages appear in `internal/entity`, `internal/usecase`, or `internal/gateway`.

## Examples

### Adding a new use case (e.g., `CreateUser`)

1. Add a `UserRepository.Save` method to `internal/usecase/port/user_repo.go` if it does not exist.
2. Create `internal/usecase/create_user.go` with a struct and `Execute(ctx, user)` method that calls the port.
3. Implement the route in `internal/controller/http/user_controller.go`:
   ```go
   mux.HandleFunc("POST /users", c.createUser)
   ```
4. Implement `Save` in `internal/gateway/postgres/user_gateway.go`.
5. Wire up the new use case in `cmd/api/main.go`.

### Adding a new gateway (e.g., an in-memory cache gateway)

1. Create `internal/gateway/memory/user_cache.go` implementing `usecase/port.UserRepository`.
2. In `cmd/api/main.go`, inject the cache gateway (or a decorator wrapping Postgres with the cache) into the use case.
3. Neither `internal/entity` nor `internal/usecase` require any modifications.

### Running use cases without HTTP (pure unit tests)

Because use cases depend only on the `usecase/port` interface, you can write:
```go
repo := memory.NewUserGateway()
uc := usecase.NewGetUserUseCase(repo)
result, err := uc.Execute(ctx, "user-123")
```
No HTTP server or database required.

## Migration from Standard Layout

1. **Rename `internal/model` to `internal/entity`.** Move business validation methods onto entity structs.
2. **Extract use-case structs.** Take the bodies of service methods and turn each one into a dedicated `Execute()` struct in `internal/usecase/`.
3. **Introduce `internal/usecase/port`.** Define the repository interface that use cases depend on.
4. **Rename `internal/repository` to `internal/gateway/postgres`.** Implementations satisfy the new port interface.
5. **Rename `internal/handler` to `internal/controller/http`.** Controllers call use cases, not service structs directly.
6. **Validate.** Run `arch_forge doctor --arch clean` to catch inward-pointing import violations.
