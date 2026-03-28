# CQRS + Event Sourcing

## Overview

Command Query Responsibility Segregation (CQRS) separates the write side (commands that change state) from the read side (queries that return projections). Combined with Event Sourcing, the write side never mutates a database row directly; instead, it applies domain events to an aggregate, and those events are stored as the source of truth. The read side maintains denormalized projections optimized for query patterns. This combination excels when audit trails are mandatory, when read and write workloads scale independently, or when a complete history of state transitions is a business requirement — for example, financial ledgers, order management systems, or compliance-sensitive platforms.

## Quick Start

```bash
arch_forge init myapp --arch cqrs --variant classic
arch_forge init myapp --arch cqrs --variant modular
```

## Generated Structure — Classic Variant

```
myapp/
├── cmd/
│   └── api/
│       └── main.go                       # Entry point; wires command/query handlers to HTTP
├── internal/
│   ├── domain/
│   │   ├── order.go                      # Aggregate root with Apply(Event) and PendingEvents()
│   │   └── events.go                     # Domain event types (OrderPlaced, OrderConfirmed, etc.)
│   ├── command/
│   │   └── place_order.go                # Command struct, write-side repository interface, handler
│   ├── query/
│   │   └── get_order.go                  # Query struct, read-model view, read-side repo interface, handler
│   └── infrastructure/
│       ├── http/
│       │   └── order_handler.go          # HTTP adapter; dispatches to command or query handler
│       └── postgres/                     # Write-side event store and read-side projection store
├── migrations/                           # SQL migration files
├── go.mod
├── .gitignore
└── archforge.yaml
```

## Generated Structure — Modular Variant

Each bounded context is isolated under `internal/<context>/`. Shared infrastructure (event bus, config) lives in `internal/shared/`.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go               # Bootstraps all bounded-context modules
├── internal/
│   ├── <context>/                # One per bounded context (e.g. orders, shipping)
│   │   ├── domain/               # Aggregate and event types for this context
│   │   ├── port/                 # Write-side and read-side repository interfaces
│   │   ├── app/                  # Command and query handlers
│   │   └── adapter/              # HTTP adapter and Postgres implementations
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
| `internal/domain` | Aggregate root, `Apply(Event)` method, `PendingEvents()`, event type definitions, optimistic concurrency `Version` field | HTTP types, SQL queries, projection logic, command/query structs |
| `internal/command` | Command structs, write-side repository interface (`Load`/`Save` with events), command handler | Read-model views, SQL SELECT queries, HTTP parsing |
| `internal/query` | Query structs, read-model view structs, read-side repository interface (`FindByID`), query handler | Write-side repository, aggregate mutation, SQL INSERT/UPDATE |
| `internal/infrastructure/http` | HTTP dispatcher routing requests to command or query handler | Business logic, direct DB access |
| `internal/infrastructure/postgres` | Write-side event store implementation; read-side projection store implementation | Command/query handler logic, domain event application |

## Compatible Modules

| Module | Purpose | Notes |
|---|---|---|
| `api` | REST router, middleware chain, request validation | Supports `chi` and `stdlib`; patches `cmd/*/main.go` via `arch_forge:routes` |
| `database` | Postgres connection pool and migrations | Patches `cmd/*/main.go` via `arch_forge:providers` |
| `queue` | Message queue for event publishing (RabbitMQ/NATS/in-memory) | Recommended for publishing events to read-side projectors; optional `logging` |
| `logging` | Structured `slog` logging with correlation IDs | Patches `cmd/*/main.go` via `arch_forge:providers` |
| `auth` | JWT middleware and token management | Requires `api` and `logging` |
| `metrics` | Prometheus metrics and `/metrics` endpoint | Requires `api` |
| `tracing` | OpenTelemetry distributed tracing | Optional dependency on `logging` |
| `healthcheck` | `/health`, `/ready`, `/live` endpoints | Requires `api` |
| `cache` | Redis cache-aside client for read-model acceleration | Requires `logging` |
| `grpc` | gRPC server for command/query dispatch | Requires `logging` |
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

`arch_forge doctor` validates the following for CQRS projects:

- `internal/domain` does not import `internal/command`, `internal/query`, or `internal/infrastructure`.
- `internal/command` does not import `internal/query`; the write side is fully independent of the read side.
- `internal/query` does not import `internal/command`; the read side is fully independent of the write side.
- Command handlers mutate state only through `aggregate.Apply(event)` then `repo.Save(ctx, agg, events)`.
- Query handlers interact only with read-side projections; they never call write-side repositories.
- The aggregate's `PendingEvents()` method is the only mechanism for passing events out of the aggregate for persistence.
- `internal/infrastructure` packages import `internal/domain`, `internal/command`, or `internal/query`; never the reverse.

## Examples

### Adding a new command (e.g., `ConfirmOrder`)

1. Add the `OrderConfirmed` event to `internal/domain/events.go` if not already present.
2. Add the `Confirm()` transition to the aggregate's `Apply` switch in `internal/domain/order.go`.
3. Create `internal/command/confirm_order.go`:
   ```go
   type ConfirmOrderCommand struct { OrderID string }
   type ConfirmOrderHandler struct { repo OrderWriteRepository }
   func (h *ConfirmOrderHandler) Handle(ctx context.Context, cmd ConfirmOrderCommand) error {
       order, _ := h.repo.Load(ctx, cmd.OrderID)
       order.Apply(domain.OrderConfirmed{OrderID: cmd.OrderID, OccurredAt: time.Now().UTC()})
       return h.repo.Save(ctx, order, order.PendingEvents())
   }
   ```
4. Add the HTTP route to `internal/infrastructure/http/order_handler.go`.

### Adding a new query (e.g., `ListOrdersByCustomer`)

1. Add a `ListByCustomerID` method to the `OrderReadRepository` interface in `internal/query/`.
2. Create a new query handler struct with a `Handle` method that calls the read-side repository.
3. Implement the SQL SELECT in `internal/infrastructure/postgres/`.
4. Add the HTTP GET route; the handler only calls the query handler, never a command handler.

### Replaying events to rebuild a projection

Because events are stored in the event store (write side), you can rebuild any read-model projection:
```go
events, _ := eventStore.LoadAll(ctx, aggregateID)
for _, e := range events {
    projector.Project(e)
}
```
No aggregate business logic is re-executed — only `Apply` is replayed.

## Migration from Standard Layout

1. **Identify write vs. read operations.** Every `service.Create/Update/Delete` becomes a command; every `service.Get/List` becomes a query.
2. **Introduce the aggregate.** Move business state into an aggregate struct with an `Apply(Event)` method. Remove direct field mutation from service code.
3. **Define event types.** Create one struct per business fact (`OrderPlaced`, `OrderConfirmed`). Implement the `Event` interface.
4. **Create command handlers.** Replace service write methods with command handler structs in `internal/command/`.
5. **Create query handlers.** Replace service read methods with query handlers in `internal/query/` backed by denormalized read-model tables.
6. **Update persistence.** Create a write-side event store table and a read-side projection table in the database.
7. **Validate.** Run `arch_forge doctor --arch cqrs` to confirm command and query sides do not import each other.
