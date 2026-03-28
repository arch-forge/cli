# Modular Monolith

## Overview

A Modular Monolith deploys as a single binary while enforcing explicit module boundaries that mirror future microservice extraction points. Each business module owns its domain model, service logic, and HTTP handler, and modules communicate only through well-defined interfaces — never by importing each other's internal packages directly. Shared infrastructure (database connections, logging, config) lives in `internal/platform`. This pattern is ideal when you want the operational simplicity of a monolith today but plan to extract services as the system grows: the module boundaries you define now become the service boundaries later.

## Quick Start

```bash
arch_forge init myapp --arch modular_monolith --variant classic
arch_forge init myapp --arch modular_monolith --variant modular
```

## Generated Structure — Classic Variant

```
myapp/
├── cmd/
│   └── api/
│       └── main.go                               # Entry point; mounts all module routes
├── internal/
│   ├── module/
│   │   └── orders/                               # One directory per business module
│   │       ├── domain/
│   │       │   └── order.go                      # Module-scoped domain entity (Order aggregate)
│   │       ├── service/
│   │       │   └── order_service.go              # Module service with internal repository interface
│   │       └── handler/
│   │           └── http_handler.go               # HTTP handler for the orders module
│   └── platform/
│       └── database/
│           └── database.go                       # Shared Postgres connection and pool
├── migrations/                                   # SQL migration files
├── go.mod
├── .gitignore
└── archforge.yaml
```

## Generated Structure — Modular Variant

The modular variant nests each module under `internal/<module-name>/` and keeps shared concerns in `internal/shared/`.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go               # Entry point; bootstraps all modules
├── internal/
│   ├── <module>/                 # One per bounded module (e.g. orders, users, inventory)
│   │   ├── domain/               # Module-scoped entities
│   │   ├── port/                 # Module's inbound/outbound port interfaces
│   │   ├── app/                  # Module's application services
│   │   └── adapter/              # Module's HTTP handlers and repository implementations
│   └── shared/
│       └── platform/
│           └── config.go         # Shared configuration loaded from environment variables
├── go.mod
├── .gitignore
└── archforge.yaml
```

## Layer Responsibilities

| Layer / Package | Belongs Here | Does NOT Belong Here |
|---|---|---|
| `internal/module/<name>/domain` | Module-scoped entities, value objects, domain errors | HTTP types, SQL, imports from other modules' `domain` packages |
| `internal/module/<name>/service` | Business logic, repository interface (defined locally), orchestration | HTTP request parsing, direct SQL, platform-level concerns |
| `internal/module/<name>/handler` | HTTP route registration, request/response mapping | Business logic, direct DB access |
| `internal/platform` | Database connections, config, logging, shared middleware | Any module-specific business logic |

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
| `queue` | Message queue for cross-module events (RabbitMQ/NATS/in-memory) | Optional dependency on `logging` |
| `grpc` | gRPC server for module-to-module communication | Requires `logging` |
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

`arch_forge doctor` validates the following for Modular Monolith projects:

- No module imports another module's `domain`, `service`, or `handler` packages directly. Cross-module communication must go through explicit interfaces or the platform event bus.
- `internal/platform` does not import any `internal/module/<name>` package.
- Each module's `domain` package does not import `internal/platform`.
- HTTP handlers within a module do not access `internal/platform/database` directly; they receive a service via dependency injection.

## Examples

### Adding a new module (e.g., `inventory`)

```bash
arch_forge add --module inventory
```

This generates:
```
internal/module/inventory/
├── domain/inventory.go
├── service/inventory_service.go
└── handler/http_handler.go
```

Register the new handler in `cmd/api/main.go` at the `// arch_forge:routes` anchor.

### Enabling cross-module communication

Modules must not import each other's packages. Use one of these patterns:
- **Event bus (recommended):** The `orders` module publishes an `OrderPlaced` event to the `queue` module; the `inventory` module subscribes and adjusts stock.
- **Shared interface:** Define a minimal interface in `internal/platform` that one module implements and another calls through.

### Extracting a module into a microservice

Because each module has clear domain, service, and handler boundaries:
1. Copy `internal/module/orders/` into a new repository.
2. Add `cmd/api/main.go` and wire the module's handler directly.
3. Replace the in-process service call from the remaining monolith with an HTTP or gRPC client implementing the same service interface.

## Migration from Standard Layout

1. **Group code by business domain.** Move `internal/handler/order*.go`, `internal/service/order*.go`, and `internal/repository/order*.go` into `internal/module/orders/handler/`, `service/`, and a repository implementation respectively.
2. **Create a domain package per module.** Move the relevant model structs from `internal/model` into `internal/module/<name>/domain/`.
3. **Create `internal/platform`.** Move the database connection and shared config out of the service layer and into `internal/platform/database/`.
4. **Define explicit cross-module contracts.** Identify places where one service calls another and introduce interface types or event publishing.
5. **Validate.** Run `arch_forge doctor --arch modular_monolith` to surface direct cross-module imports.
