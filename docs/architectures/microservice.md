# Microservice

## Overview

The Microservice pattern generates a single, self-contained service focused on one business capability. The structure is lightweight and production-ready out of the box: domain and port interfaces at the center, a thin application service, and adapters for HTTP and Postgres on the outside. Unlike Hexagonal Architecture's strict inbound/outbound port separation, Microservice collapses the adapter hierarchy into flat `internal/adapter/http` and `internal/adapter/postgres` packages for less ceremony. Choose this pattern when deploying a standalone service that will be operated, scaled, and released independently from other services in your platform.

## Quick Start

```bash
arch_forge init myapp --arch microservice --variant classic
arch_forge init myapp --arch microservice --variant modular
```

## Generated Structure — Classic Variant

```
myapp/
├── cmd/
│   └── api/
│       └── main.go                       # Entry point; wires service, sets up HTTP server
├── internal/
│   ├── domain/
│   │   ├── order.go                      # Primary aggregate (Order struct)
│   │   └── errors.go                     # Sentinel errors for domain failures
│   ├── port/
│   │   └── order_repository.go           # OrderRepository interface (persistence contract)
│   ├── app/
│   │   └── order_service.go              # Application service implementing business logic
│   └── adapter/
│       ├── http/
│       │   └── order_handler.go          # HTTP handler; routes and serializes responses
│       └── postgres/
│           └── order_repo.go             # Postgres implementation of port.OrderRepository
├── migrations/                           # SQL migration files
├── go.mod
├── .gitignore
└── archforge.yaml
```

## Generated Structure — Modular Variant

The modular variant is typically used when a service manages more than one internal concern. Each concern gets its own directory under `internal/`. Shared infrastructure lives in `internal/shared/`.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go               # Entry point
├── internal/
│   ├── <concern>/                # One per internal concern (e.g. orders, notifications)
│   │   ├── domain/               # Entity and errors scoped to this concern
│   │   ├── port/                 # Repository/service interface for this concern
│   │   ├── app/                  # Application service for this concern
│   │   └── adapter/              # HTTP and Postgres adapters for this concern
│   └── shared/
│       └── platform/
│           └── config.go         # Shared config from environment variables
├── go.mod
├── .gitignore
└── archforge.yaml
```

## Layer Responsibilities

| Layer / Package | Belongs Here | Does NOT Belong Here |
|---|---|---|
| `internal/domain` | Core entity structs, sentinel errors | HTTP types, SQL, framework imports, business orchestration |
| `internal/port` | Repository and external service interface definitions | Implementations, SQL, HTTP |
| `internal/app` | Application service orchestrating domain and port calls | HTTP parsing, JSON marshalling, SQL, direct DB calls |
| `internal/adapter/http` | HTTP route registration, request binding, response serialization | Business logic, direct DB access |
| `internal/adapter/postgres` | SQL queries, Postgres-specific implementation of port interfaces | Business rules, HTTP types |

## Compatible Modules

| Module | Purpose | Notes |
|---|---|---|
| `api` | REST router, middleware chain, request validation | Supports `chi` and `stdlib`; patches `cmd/*/main.go` via `arch_forge:routes` |
| `database` | Postgres connection pool and migrations | Patches `cmd/*/main.go` via `arch_forge:providers` |
| `auth` | JWT middleware and token management | Requires `api` and `logging` |
| `logging` | Structured `slog` logging with correlation IDs | Patches `cmd/*/main.go` via `arch_forge:providers` |
| `metrics` | Prometheus metrics and `/metrics` endpoint | Requires `api` |
| `tracing` | OpenTelemetry distributed tracing with OTLP export | Optional dependency on `logging` |
| `healthcheck` | `/health`, `/ready`, `/live` endpoints for Kubernetes | Requires `api` |
| `cache` | Redis cache-aside client | Requires `logging` |
| `queue` | Message queue (RabbitMQ/NATS/in-memory) for event publishing | Optional dependency on `logging` |
| `grpc` | gRPC server for service-to-service calls | Requires `logging` |
| `cors` | CORS middleware | Requires `api` |
| `ratelimit` | Token-bucket rate limiting per IP | Requires `api` |
| `encryption` | AES-GCM encryption and PBKDF2 key derivation | No dependencies |
| `storage` | Object storage client (S3/GCS/local) | No dependencies |
| `search` | Full-text search (Elasticsearch/Meilisearch) | Optional dependency on `logging` |
| `docker` | Multi-stage Dockerfile and docker-compose | Optional dependency on `database` |
| `ci` | GitHub Actions / GitLab CI pipelines | Optional dependency on `docker` |
| `k8s` | Kubernetes manifests (Deployment, Service, HPA, Ingress) | Requires `docker` |
| `terraform` | Terraform infrastructure for AWS/GCP/Azure | No dependencies |
| `testkit` | Test fixtures, factories, testcontainers | Optional dependency on `database` |
| `e2e` | End-to-end HTTP test scaffold | Requires `api`; optional `testkit` |
| `mocks` | mockery-generated test doubles from interfaces | No dependencies |
| `makefile` | Standard `make build/test/lint/run/migrate` targets | No dependencies |

## Architecture Rules (doctor checks)

`arch_forge doctor` validates the following for Microservice projects:

- `internal/domain` imports no `internal/port`, `internal/app`, or `internal/adapter` packages.
- `internal/port` imports only `internal/domain`.
- `internal/app` imports `internal/domain` and `internal/port` only; never `internal/adapter` directly.
- `internal/adapter/http` does not import `internal/adapter/postgres` and vice versa.
- No database packages appear in `internal/domain`, `internal/port`, or `internal/app`.
- No HTTP packages appear in `internal/domain`, `internal/port`, or `internal/app`.

## Examples

### Adding a new HTTP endpoint (e.g., `CreateOrder`)

1. Add a `Save` method to `internal/port/order_repository.go` if not present.
2. Add a `CreateOrder` method to `internal/app/order_service.go`:
   ```go
   func (s *OrderService) CreateOrder(ctx context.Context, order *domain.Order) error {
       return s.repo.Save(ctx, order)
   }
   ```
3. Register the route in `internal/adapter/http/order_handler.go`:
   ```go
   mux.HandleFunc("POST /orders", h.createOrder)
   ```
4. Implement `Save` in `internal/adapter/postgres/order_repo.go`.
5. Wire the new handler in `cmd/api/main.go`.

### Adding Kubernetes and Docker support

```bash
arch_forge add docker
arch_forge add k8s
```

This generates a multi-stage `Dockerfile`, a `docker-compose.yml` for local development, and Kubernetes manifests including `Deployment`, `Service`, `HPA`, and `ConfigMap`.

### Writing a unit test for the application service

```go
func TestOrderService_GetOrder(t *testing.T) {
    repo := mocks.NewOrderRepository(t)
    repo.On("FindByID", mock.Anything, "order-1").Return(&domain.Order{ID: "order-1"}, nil)

    svc := app.NewOrderService(repo)
    order, err := svc.GetOrder(context.Background(), "order-1")
    require.NoError(t, err)
    assert.Equal(t, "order-1", order.ID)
}
```

The port interface makes the repository trivially mockable — no database needed.

## Migration from Standard Layout

1. **Keep the flat package layout but introduce ports.** Create `internal/port/` and define repository interfaces there.
2. **Rename `internal/model` to `internal/domain`.** Keep domain types pure — remove any service or repository imports.
3. **Rename `internal/service` to `internal/app`.** Update the service to depend on the new `internal/port` interfaces.
4. **Move `internal/repository` to `internal/adapter/postgres`.** Update the implementation to satisfy the new port interfaces.
5. **Move `internal/handler` to `internal/adapter/http`.** Update handlers to call the `app` layer instead of the repository directly.
6. **Validate.** Run `arch_forge doctor --arch microservice` to catch remaining layer violations.
