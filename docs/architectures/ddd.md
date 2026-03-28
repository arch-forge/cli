# Domain-Driven Design (DDD)

## Overview

Domain-Driven Design centers the entire project around the business domain, modeled through aggregates, value objects, domain events, and repository contracts defined inside the domain itself. The domain model owns its lifecycle transitions (e.g., `Order.Place()`, `Order.Confirm()`) and enforces invariants rather than delegating business decisions to service layers. Infrastructure (HTTP handlers, Postgres repositories) exists in `internal/infrastructure` and depends on the domain, never the other way around. Choose DDD when the problem domain is rich, complex, and collaborative with domain experts — for example, e-commerce, supply chain, or financial systems where the business rules themselves are the product's core value.

## Quick Start

```bash
arch_forge init myapp --arch ddd --variant classic
arch_forge init myapp --arch ddd --variant modular
```

## Generated Structure — Classic Variant

```
myapp/
├── cmd/
│   └── api/
│       └── main.go                                       # Entry point; wires infrastructure to domain
├── internal/
│   ├── domain/
│   │   ├── model/
│   │   │   ├── order.go                                  # Aggregate root with lifecycle methods (Place, Confirm)
│   │   │   └── errors.go                                 # Domain errors (ErrOrderEmpty, ErrInvalidTransition)
│   │   └── repository/
│   │       └── order_repository.go                       # Repository interface defined by the domain
│   ├── application/
│   │   └── order_service.go                              # Application service orchestrating domain use cases
│   └── infrastructure/
│       ├── http/
│       │   └── order_handler.go                          # HTTP handler; adapts HTTP to application service
│       └── persistence/
│           └── postgres/
│               └── order_repo.go                         # Postgres implementation of domain repository
├── migrations/                                           # SQL migration files
├── go.mod
├── .gitignore
└── archforge.yaml
```

## Generated Structure — Modular Variant

Each bounded context is its own module directory under `internal/`. Shared platform concerns live in `internal/shared/`.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go               # Bootstraps all bounded-context modules
├── internal/
│   ├── <context>/                # One per bounded context (e.g. orders, inventory, payments)
│   │   ├── domain/               # Aggregates, value objects, and repository interfaces for this context
│   │   ├── port/                 # Inbound application service interfaces (optional)
│   │   ├── app/                  # Application services coordinating domain operations
│   │   └── adapter/              # HTTP handlers and Postgres repositories for this context
│   └── shared/
│       └── platform/
│           └── config.go         # Shared application configuration
├── go.mod
├── .gitignore
└── archforge.yaml
```

## Layer Responsibilities

| Layer / Package | Belongs Here | Does NOT Belong Here |
|---|---|---|
| `internal/domain/model` | Aggregate roots, value objects, domain errors, lifecycle state machine methods | HTTP types, SQL, framework imports, orchestration logic |
| `internal/domain/repository` | Repository interfaces defined and owned by the domain | Implementations, SQL statements, ORM calls |
| `internal/application` | Application services that orchestrate domain objects via repository interfaces | HTTP parsing, SQL, direct infrastructure calls |
| `internal/infrastructure/http` | HTTP handlers, request parsing, response serialization | Business rules, direct DB access, domain mutation |
| `internal/infrastructure/persistence/postgres` | Postgres implementations of domain repository interfaces | Business rules, HTTP types, application orchestration |

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
| `grpc` | gRPC server with buf config and interceptors | Requires `logging` |
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

`arch_forge doctor` validates the following for DDD projects:

- `internal/domain/model` imports no `internal/infrastructure` or `internal/application` packages.
- `internal/domain/repository` imports only `internal/domain/model`; the interface is defined by and for the domain.
- `internal/application` imports `internal/domain/model` and `internal/domain/repository` only; it never imports `internal/infrastructure` directly.
- `internal/infrastructure/http` imports `internal/application` and/or `internal/domain/model`; not `internal/infrastructure/persistence`.
- `internal/infrastructure/persistence/postgres` imports `internal/domain/model` and `internal/domain/repository` only.
- Aggregate lifecycle transitions (state changes, invariant checks) live in `internal/domain/model`, not in application services.

## Examples

### Modeling a new aggregate (e.g., `Payment`)

1. Create `internal/domain/model/payment.go` with the `Payment` struct and state machine methods (`Authorize`, `Capture`, `Refund`).
2. Define `internal/domain/repository/payment_repository.go` with the `PaymentRepository` interface.
3. Implement the application use case in `internal/application/payment_service.go`.
4. Add the HTTP handler in `internal/infrastructure/http/payment_handler.go`.
5. Implement the Postgres repository in `internal/infrastructure/persistence/postgres/payment_repo.go`.

### Enforcing a domain invariant

Add the invariant check directly to the aggregate method, not the application service:
```go
// internal/domain/model/order.go
func (o *Order) Confirm() error {
    if o.Status != OrderStatusPending {
        return ErrInvalidTransition
    }
    o.Status = OrderStatusConfirmed
    return nil
}
```
The application service calls `order.Confirm()` and propagates the error up; it never inspects the status field itself.

### Adding a second bounded context (modular variant)

```bash
arch_forge add --module inventory
```

This creates `internal/inventory/` with its own domain model, repository interface, application service, and adapters — completely isolated from the `orders` context.

## Migration from Standard Layout

1. **Identify your aggregates.** Group related entities from `internal/model` into aggregate clusters with a single root.
2. **Move domain logic.** State changes currently in `internal/service` methods belong in aggregate methods. Move them.
3. **Define repository interfaces inside the domain.** Create `internal/domain/repository/` and declare interfaces there. Remove service-layer repository interfaces.
4. **Create the application service layer.** `internal/application` becomes a thin orchestrator that calls domain methods and persists via repository interfaces.
5. **Move repository implementations to `internal/infrastructure/persistence/postgres`.** Rename `internal/repository`.
6. **Move HTTP handlers to `internal/infrastructure/http`.** Rename `internal/handler`.
7. **Validate.** Run `arch_forge doctor --arch ddd` to confirm infrastructure does not bleed into the domain.
