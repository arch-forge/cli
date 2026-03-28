# Hexagonal Architecture (Ports & Adapters)

## Overview

Hexagonal Architecture, also known as Ports & Adapters, isolates the core domain from all external concerns by defining explicit input and output ports as interfaces. The domain and application logic sit at the center and have zero knowledge of HTTP, databases, or any framework. Adapters implement ports and can be swapped independently — replacing a Postgres repository with an in-memory store requires no changes to domain or application code. This pattern excels when testability, long-term maintainability, and the ability to change infrastructure without touching business logic are priorities. Choose Hexagonal when you expect the external ecosystem (databases, messaging systems, HTTP frameworks) to evolve independently of your core rules.

## Quick Start

```bash
arch_forge init myapp --arch hexagonal --variant classic
arch_forge init myapp --arch hexagonal --variant modular
```

## Generated Structure — Classic Variant

```
myapp/
├── cmd/
│   └── api/
│       └── main.go               # Entry point; wires adapters to ports
├── internal/
│   ├── domain/
│   │   ├── user.go               # Core entity with business rules (Validate)
│   │   └── errors.go             # Sentinel errors (ErrInvalidUser, etc.)
│   ├── port/
│   │   ├── input/
│   │   │   └── user_service.go   # Driving port — interface consumed by inbound adapters
│   │   └── output/
│   │       └── user_repository.go # Driven port — interface implemented by outbound adapters
│   ├── app/
│   │   └── user_service.go       # Application service implementing input.UserService
│   └── adapter/
│       ├── inbound/
│       │   └── http/
│       │       └── user_handler.go # HTTP adapter; calls input.UserService port
│       └── outbound/
│           └── postgres/
│               └── user_repo.go  # Postgres adapter implementing output.UserRepository
├── migrations/                   # SQL migration files
├── go.mod
├── .gitignore
└── archforge.yaml                # arch_forge project manifest
```

## Generated Structure — Modular Variant

In the modular variant each business domain lives under `internal/<domain-name>/`. The shared platform code (config, DB connection, logging) is placed in `internal/shared/`.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go               # Entry point; bootstraps all modules
├── internal/
│   ├── <module>/                 # One directory per bounded domain (e.g. users, orders)
│   │   ├── domain/               # Entities and domain errors for this module
│   │   ├── port/                 # Input and output port interfaces
│   │   ├── app/                  # Application services for this module
│   │   ├── adapter/              # Inbound and outbound adapters (HTTP, Postgres, etc.)
│   │   └── migrations/           # Module-scoped SQL migrations
│   └── shared/
│       └── platform/
│           └── config.go         # Shared config loaded from environment variables
├── go.mod
├── .gitignore
└── archforge.yaml
```

## Layer Responsibilities

| Layer / Package | Belongs Here | Does NOT Belong Here |
|---|---|---|
| `internal/domain` | Entities, value objects, domain errors, invariant validation methods | HTTP types, SQL queries, framework imports, logging |
| `internal/port/input` | Driving port interfaces consumed by inbound adapters | Implementations, structs, business logic |
| `internal/port/output` | Driven port interfaces implemented by outbound adapters | Implementations, SQL, external SDK imports |
| `internal/app` | Application services that orchestrate domain logic via output ports | HTTP parsing, JSON marshalling, SQL, framework code |
| `internal/adapter/inbound/http` | HTTP handlers, request/response mapping, route registration | Business logic, direct DB access, domain mutation |
| `internal/adapter/outbound/postgres` | Postgres implementations of output ports, SQL queries | Business rules, HTTP-specific types |

## Compatible Modules

| Module | Purpose | Notes |
|---|---|---|
| `api` | REST router, middleware chain, request validation | Supports `chi` and `stdlib`; patches `cmd/*/main.go` via `arch_forge:routes` anchor |
| `database` | Postgres connection pool and migrations | Patches `cmd/*/main.go` via `arch_forge:providers` anchor |
| `auth` | JWT middleware and token management | Requires `api` and `logging` |
| `logging` | Structured `slog` logging with correlation IDs | Patches `cmd/*/main.go` via `arch_forge:providers` |
| `metrics` | Prometheus counters, histograms, `/metrics` endpoint | Requires `api` |
| `tracing` | OpenTelemetry distributed tracing with OTLP export | Optional dependency on `logging` |
| `healthcheck` | `/health`, `/ready`, `/live` endpoints | Requires `api`; patches routes anchor |
| `cache` | Redis cache-aside client with TTL management | Requires `logging` |
| `queue` | RabbitMQ/NATS/in-memory message queue client | Optional dependency on `logging` |
| `grpc` | gRPC server with buf config and interceptors | Requires `logging` |
| `crud` | Full CRUD scaffold for an entity | Requires `api` and `database` |
| `auth` | JWT authentication | Requires `api`, `logging` |
| `docker` | Multi-stage Dockerfile and docker-compose | Optional dependency on `database` |
| `ci` | GitHub Actions / GitLab CI pipelines | Optional dependency on `docker` |
| `k8s` | Kubernetes manifests (Deployment, Service, Ingress, HPA) | Requires `docker` |
| `testkit` | Test fixtures, factories, testcontainers | Optional dependency on `database` |
| `e2e` | End-to-end HTTP test scaffold | Requires `api`; optional `testkit` |
| `mocks` | mockery-generated test doubles from interfaces | No dependencies |
| `makefile` | Standard `make build/test/lint/run/migrate` targets | No dependencies |

## Architecture Rules (doctor checks)

`arch_forge doctor` validates the following for Hexagonal projects:

- `internal/domain` does not import any `internal/port`, `internal/app`, or `internal/adapter` package.
- `internal/port` imports only `internal/domain`.
- `internal/app` imports only `internal/domain` and `internal/port`; never `internal/adapter`.
- `internal/adapter` packages import `internal/port` and/or `internal/domain` but never other adapters directly.
- No HTTP framework packages (`net/http`, `github.com/go-chi/chi`, etc.) appear in `internal/domain` or `internal/app`.
- No database driver packages (`database/sql`, `github.com/jackc/pgx`, etc.) appear in `internal/domain`, `internal/port`, or `internal/app`.

## Examples

### Adding a new HTTP endpoint (e.g., `DeleteUser`)

1. Add the method signature to `internal/port/input/user_service.go`:
   ```go
   DeleteUser(ctx context.Context, id string) error
   ```
2. Implement it in `internal/app/user_service.go`, calling the output port:
   ```go
   func (s *userService) DeleteUser(ctx context.Context, id string) error { ... }
   ```
3. Add a route handler in `internal/adapter/inbound/http/user_handler.go`:
   ```go
   mux.HandleFunc("DELETE /users/{id}", h.deleteUser)
   ```
4. Run `arch_forge doctor` to confirm no layer violations were introduced.

### Adding a new outbound adapter (e.g., switching from Postgres to MySQL)

1. Create `internal/adapter/outbound/mysql/user_repo.go` implementing `output.UserRepository`.
2. In `cmd/api/main.go`, replace the Postgres constructor with the MySQL one.
3. No changes needed in `internal/domain`, `internal/port`, or `internal/app`.

### Adding the `logging` module to an existing project

```bash
arch_forge add logging
```

This injects a structured logger initialization block into `cmd/api/main.go` at the `// arch_forge:providers` anchor and adds the required slog wrapper package.

## Migration from Standard Layout

1. **Identify your layers.** Map `internal/service` to `internal/app`, `internal/repository` to a new `internal/port/output` interface plus an `internal/adapter/outbound/postgres` implementation, and `internal/handler` to `internal/adapter/inbound/http`.
2. **Extract port interfaces.** For each service and repository, create a matching interface in `internal/port/input` or `internal/port/output`.
3. **Isolate the domain.** Move plain structs from `internal/model` to `internal/domain`. Remove any non-domain imports.
4. **Rewrite constructors.** Wire through interfaces rather than concrete types so adapters depend on ports, not each other.
5. **Validate.** Run `arch_forge doctor --arch hexagonal` after migration to detect remaining import violations.
