# PRD — arch_forge

## Visión General

**arch_forge** es una herramienta CLI que genera estructuras de proyectos de software basadas en patrones arquitectónicos probados en la industria. Inicialmente enfocada en Go, permite crear proyectos completos con scaffolding inteligente y módulos adicionales que respetan los principios de cada arquitectura elegida.

No es un simple generador de boilerplate. arch_forge entiende la arquitectura que el usuario elige y garantiza que cada archivo, paquete y dependencia esté donde debe estar según los principios de esa arquitectura.

---

## Problema

1. **Inicio lento**: Los desarrolladores pierden horas configurando la estructura inicial de un proyecto, decidiendo dónde poner cada cosa.
2. **Arquitecturas mal implementadas**: Se elige una arquitectura (hexagonal, clean, etc.) pero se termina con una mezcla inconsistente por falta de guía concreta.
3. **Módulos desconectados**: Agregar un nuevo feature (auth, payments, notifications) requiere crear manualmente decenas de archivos respetando la estructura existente.
4. **Falta de convenciones**: Cada equipo reinventa la rueda con su propia estructura, dificultando la onboarding y la mantenibilidad.

---

## Solución

Una CLI que:

- Genera proyectos completos basados en arquitecturas específicas.
- Permite agregar módulos post-creación que se integran respetando la arquitectura elegida.
- Incluye configuración de CI/CD, Docker, linters, testing y documentación desde el día cero.
- Es extensible: los usuarios pueden crear y compartir sus propios templates y módulos.

---

## Arquitecturas Soportadas (v1 — Go)

Cada arquitectura soporta dos variantes que se seleccionan con el flag `--variant`:

| Variante | Filosofía | Cuándo elegirla |
|---|---|---|
| **`classic`** | Fiel a la definición original del libro/paper. Nomenclatura canónica, capas estrictas, separación purista. | Equipos que quieren implementar la arquitectura "by the book", proyectos académicos, cuando la trazabilidad hacia la teoría importa. |
| **`modular`** | Reorganiza la misma arquitectura en módulos de negocio autocontenidos. Cada módulo encapsula sus propias capas internamente. | Proyectos que crecerán con múltiples dominios, equipos grandes, preparación para eventual splitting a microservicios. |

```bash
# Selección de variante
arch_forge init myapp --arch=hexagonal --variant=classic
arch_forge init myapp --arch=hexagonal --variant=modular  # default
```

> **Default**: `modular`. La variante modular es el default porque escala mejor en proyectos reales. Los usuarios que quieran la versión purista del libro pueden optar por `classic` explícitamente.

---

### Hexagonal (Ports & Adapters)

Dominio aislado con puertos de entrada/salida y adaptadores intercambiables.

**Classic** — Basada en el paper original de Alistair Cockburn. Estructura por tipo de componente.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/                  # Entidades y reglas de negocio
│   │   ├── order.go
│   │   ├── product.go
│   │   └── user.go
│   ├── port/                    # Interfaces (driving & driven)
│   │   ├── input/               # Driving ports (casos de uso)
│   │   │   ├── order_service.go
│   │   │   └── user_service.go
│   │   └── output/              # Driven ports (repositorios, servicios externos)
│   │       ├── order_repository.go
│   │       └── notification_service.go
│   ├── app/                     # Implementación de los driving ports
│   │   ├── order_service.go
│   │   └── user_service.go
│   └── adapter/                 # Implementaciones de los driven ports
│       ├── inbound/             # Adaptadores de entrada
│       │   ├── http/
│       │   │   ├── router.go
│       │   │   ├── order_handler.go
│       │   │   └── user_handler.go
│       │   └── grpc/
│       └── outbound/            # Adaptadores de salida
│           ├── postgres/
│           │   ├── order_repo.go
│           │   └── user_repo.go
│           ├── redis/
│           └── smtp/
├── go.mod
└── archforge.yaml
```

**Modular** — Mismos principios hexagonales pero organizado por dominio de negocio. Cada módulo es un hexágono autocontenido.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── order/                   # Módulo de negocio: Order
│   │   ├── domain/
│   │   │   └── order.go
│   │   ├── port/
│   │   │   ├── order_service.go       # Driving port
│   │   │   └── order_repository.go    # Driven port
│   │   ├── app/
│   │   │   └── order_service.go       # Implementación del caso de uso
│   │   └── adapter/
│   │       ├── handler.go             # HTTP handler
│   │       └── postgres_repo.go       # Repositorio PostgreSQL
│   ├── user/                    # Módulo de negocio: User
│   │   ├── domain/
│   │   ├── port/
│   │   ├── app/
│   │   └── adapter/
│   └── shared/                  # Código compartido entre módulos
│       ├── domain/              # Value objects comunes
│       └── platform/            # Middleware, logging, config
├── go.mod
└── archforge.yaml
```

---

### Clean Architecture

Capas concéntricas: entities → use cases → interface adapters → frameworks.

**Classic** — Fiel al diagrama de Robert C. Martin. Capas concéntricas con dependency rule estricta.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── entity/                  # Capa 1: Enterprise Business Rules
│   │   ├── order.go
│   │   ├── product.go
│   │   └── user.go
│   ├── usecase/                 # Capa 2: Application Business Rules
│   │   ├── create_order.go
│   │   ├── get_user.go
│   │   └── port/               # Interfaces que los use cases necesitan
│   │       ├── order_repo.go
│   │       └── user_repo.go
│   ├── controller/              # Capa 3: Interface Adapters
│   │   ├── http/
│   │   │   ├── order_controller.go
│   │   │   └── presenter/
│   │   │       └── order_presenter.go
│   │   └── grpc/
│   ├── gateway/                 # Capa 3: Interface Adapters (data access)
│   │   ├── postgres/
│   │   │   └── order_gateway.go
│   │   └── redis/
│   └── framework/               # Capa 4: Frameworks & Drivers
│       ├── router.go
│       ├── database.go
│       └── config.go
├── go.mod
└── archforge.yaml
```

**Modular** — Cada módulo de negocio implementa internamente las 4 capas de Clean Architecture.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── order/
│   │   ├── entity/
│   │   │   └── order.go
│   │   ├── usecase/
│   │   │   ├── create_order.go
│   │   │   └── port/
│   │   │       └── order_repo.go
│   │   ├── controller/
│   │   │   └── http_handler.go
│   │   └── gateway/
│   │       └── postgres_repo.go
│   ├── user/
│   │   ├── entity/
│   │   ├── usecase/
│   │   ├── controller/
│   │   └── gateway/
│   └── shared/
│       └── framework/           # Router, DB connection, config
├── go.mod
└── archforge.yaml
```

---

### DDD (Domain-Driven Design)

Bounded contexts, aggregates, value objects, repositories, domain events.

**Classic** — Estructura por building blocks tácticos de DDD (Evans, 2003).

```
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── model/
│   │   │   ├── aggregate/
│   │   │   │   ├── order.go         # Aggregate root
│   │   │   │   └── customer.go
│   │   │   ├── entity/
│   │   │   │   └── order_item.go
│   │   │   └── valueobject/
│   │   │       ├── money.go
│   │   │       ├── email.go
│   │   │       └── address.go
│   │   ├── event/
│   │   │   ├── order_placed.go
│   │   │   └── order_shipped.go
│   │   ├── repository/              # Interfaces
│   │   │   ├── order_repository.go
│   │   │   └── customer_repository.go
│   │   └── service/                 # Domain services
│   │       └── pricing_service.go
│   ├── application/
│   │   ├── command/
│   │   │   ├── place_order.go
│   │   │   └── handler/
│   │   │       └── place_order_handler.go
│   │   ├── query/
│   │   │   └── get_order.go
│   │   └── event_handler/
│   │       └── on_order_placed.go
│   └── infrastructure/
│       ├── persistence/
│       │   └── postgres/
│       │       └── order_repo.go
│       ├── messaging/
│       └── http/
│           ├── router.go
│           └── order_handler.go
├── go.mod
└── archforge.yaml
```

**Modular** — Organizado por Bounded Contexts. Cada contexto es un módulo independiente con su propio modelo de dominio.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── ordering/                    # Bounded Context: Ordering
│   │   ├── domain/
│   │   │   ├── order.go            # Aggregate root
│   │   │   ├── order_item.go       # Entity
│   │   │   ├── money.go            # Value object
│   │   │   ├── order_placed.go     # Domain event
│   │   │   └── order_repository.go # Repository interface
│   │   ├── application/
│   │   │   ├── place_order.go
│   │   │   └── get_order.go
│   │   └── infrastructure/
│   │       ├── postgres_repo.go
│   │       └── http_handler.go
│   ├── catalog/                     # Bounded Context: Catalog
│   │   ├── domain/
│   │   ├── application/
│   │   └── infrastructure/
│   ├── identity/                    # Bounded Context: Identity
│   │   ├── domain/
│   │   ├── application/
│   │   └── infrastructure/
│   └── shared/
│       └── kernel/                  # Shared Kernel entre contexts
│           ├── event_bus.go
│           └── types.go
├── go.mod
└── archforge.yaml
```

---

### Standard Layout

Estructura estándar de la comunidad Go (`cmd/`, `internal/`, `pkg/`).

**Classic** — Sigue el [golang-standards/project-layout](https://github.com/golang-standards/project-layout).

```
myapp/
├── cmd/
│   └── myapp/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   ├── order_handler.go
│   │   └── user_handler.go
│   ├── service/
│   │   ├── order_service.go
│   │   └── user_service.go
│   ├── repository/
│   │   ├── order_repo.go
│   │   └── user_repo.go
│   ├── model/
│   │   ├── order.go
│   │   └── user.go
│   └── middleware/
│       ├── auth.go
│       └── logging.go
├── pkg/                             # Código exportable/reutilizable
│   └── validator/
│       └── validator.go
├── go.mod
└── archforge.yaml
```

**Modular** — Standard layout pero con agrupación por feature.

```
myapp/
├── cmd/
│   └── myapp/
│       └── main.go
├── internal/
│   ├── order/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   ├── user/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   └── platform/
│       ├── config/
│       ├── middleware/
│       └── database/
├── pkg/
│   └── validator/
├── go.mod
└── archforge.yaml
```

---

### Modular Monolith

Módulos internos con boundaries claros, preparados para eventual extracción.

**Classic** — Monolito con módulos explícitos y comunicación a través de interfaces públicas.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── module/
│   │   ├── order/
│   │   │   ├── module.go            # Registro del módulo, expone API pública
│   │   │   ├── api.go               # Interfaz pública del módulo
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── model.go
│   │   ├── inventory/
│   │   │   ├── module.go
│   │   │   ├── api.go
│   │   │   ├── service.go
│   │   │   └── model.go
│   │   └── notification/
│   │       ├── module.go
│   │       ├── api.go
│   │       └── service.go
│   ├── registry/                    # Registro central de módulos
│   │   └── registry.go
│   └── platform/
│       ├── event_bus.go             # Comunicación inter-módulo
│       ├── config.go
│       └── database.go
├── go.mod
└── archforge.yaml
```

**Modular** — Cada módulo es completamente autónomo con su propia arquitectura interna (hexagonal por defecto).

```
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── order/
│   │   ├── module.go                # Init, routes, dependency wiring
│   │   ├── api.go                   # Interfaz pública (contrato con otros módulos)
│   │   ├── domain/
│   │   │   └── order.go
│   │   ├── port/
│   │   │   └── order_repository.go
│   │   ├── app/
│   │   │   └── order_service.go
│   │   ├── adapter/
│   │   │   ├── handler.go
│   │   │   └── postgres_repo.go
│   │   └── migrations/
│   │       └── 001_create_orders.sql
│   ├── inventory/
│   │   ├── module.go
│   │   ├── api.go
│   │   ├── domain/
│   │   ├── port/
│   │   ├── app/
│   │   └── adapter/
│   └── shared/
│       ├── event/                   # Event bus para comunicación async
│       │   ├── bus.go
│       │   └── events.go
│       └── platform/
├── go.mod
└── archforge.yaml
```

---

### CQRS + Event Sourcing

Separación de comandos y queries con store de eventos como fuente de verdad.

**Classic** — Separación estricta command/query side con event store centralizado.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── command/                     # Write side
│   │   ├── aggregate/
│   │   │   └── order.go            # Aggregate con Apply/Handle
│   │   ├── handler/
│   │   │   └── place_order_handler.go
│   │   └── command/
│   │       └── place_order.go
│   ├── query/                       # Read side
│   │   ├── handler/
│   │   │   └── get_order_handler.go
│   │   ├── query/
│   │   │   └── get_order.go
│   │   └── projection/
│   │       └── order_projection.go
│   ├── event/
│   │   ├── store/                   # Event store
│   │   │   └── postgres_store.go
│   │   ├── bus/
│   │   │   └── event_bus.go
│   │   └── events/
│   │       ├── order_placed.go
│   │       └── order_shipped.go
│   ├── projection/                  # Read model builders
│   │   ├── projector.go
│   │   └── order_projector.go
│   └── infrastructure/
│       ├── http/
│       └── database/
├── go.mod
└── archforge.yaml
```

**Modular** — CQRS por módulo de negocio. Cada módulo tiene su propio command/query side y eventos.

```
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── order/
│   │   ├── command/
│   │   │   ├── place_order.go
│   │   │   ├── handler.go
│   │   │   └── aggregate.go
│   │   ├── query/
│   │   │   ├── get_order.go
│   │   │   ├── handler.go
│   │   │   └── projection.go
│   │   └── event/
│   │       ├── order_placed.go
│   │       └── order_shipped.go
│   ├── inventory/
│   │   ├── command/
│   │   ├── query/
│   │   └── event/
│   └── shared/
│       ├── eventstore/              # Event store compartido
│       │   ├── store.go
│       │   └── postgres_store.go
│       ├── projector/               # Engine de proyecciones
│       └── platform/
├── go.mod
└── archforge.yaml
```

---

### Microservice (Single)

Servicio individual listo para producción con health checks, graceful shutdown, observability.

**Classic** — Estructura flat optimizada para un servicio con responsabilidad única.

```
myapp/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   └── order.go
│   ├── service/
│   │   └── order_service.go
│   ├── handler/
│   │   ├── http/
│   │   │   └── order_handler.go
│   │   └── grpc/
│   │       └── order_handler.go
│   ├── repository/
│   │   └── postgres_repo.go
│   └── client/                      # Clients para otros servicios
│       └── inventory_client.go
├── proto/
│   └── order/
│       └── v1/
│           └── order.proto
├── api/
│   └── openapi/
│       └── order.yaml
├── go.mod
└── archforge.yaml
```

**Modular** — Para microservicios más complejos que manejan múltiples subdominios.

```
myapp/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── order/
│   │   ├── domain.go
│   │   ├── service.go
│   │   ├── handler.go
│   │   └── repository.go
│   ├── fulfillment/
│   │   ├── domain.go
│   │   ├── service.go
│   │   └── handler.go
│   ├── client/
│   │   └── inventory_client.go
│   └── platform/
│       ├── config/
│       ├── server/                  # HTTP + gRPC server setup
│       ├── health/
│       └── observability/
├── proto/
│   └── order/
│       └── v1/
│           └── order.proto
├── go.mod
└── archforge.yaml
```

---

## Módulos Disponibles

Módulos que el usuario puede agregar a un proyecto existente. Cada módulo genera los archivos en la capa correcta según la arquitectura del proyecto.

### Core

| Módulo | Qué genera |
|---|---|
| `auth` | Autenticación JWT/OAuth2 con middleware, handlers, repositorio y migraciones. |
| `crud --entity=<name>` | CRUD completo para una entidad: handler, service, repository, modelo, migraciones, tests. |
| `api` | REST API con router, middleware chain, error handling, request validation, OpenAPI spec. |
| `grpc` | Servicio gRPC con proto files, server, interceptors y client stub. |
| `graphql` | Schema GraphQL, resolvers, dataloaders y playground. |
| `websocket` | Server WebSocket con rooms, broadcast y connection management. |

### Infraestructura

| Módulo | Qué genera |
|---|---|
| `database --driver=<postgres\|mysql\|sqlite\|mongo>` | Conexión, pool, migraciones, seeders y health check. |
| `cache --driver=<redis\|memcached\|in-memory>` | Cliente de caché con patrones cache-aside, write-through, TTL config. |
| `queue --driver=<rabbitmq\|kafka\|nats\|sqs>` | Producer/consumer con retry, DLQ y graceful shutdown. |
| `storage --driver=<s3\|gcs\|local>` | Abstracción de file storage con upload, download, presigned URLs. |
| `search --driver=<elasticsearch\|meilisearch>` | Indexación, búsqueda full-text y sincronización con source of truth. |

### Observabilidad

| Módulo | Qué genera |
|---|---|
| `logging` | Structured logging con slog, log levels, correlation IDs. |
| `metrics` | Métricas Prometheus con collectors custom, histogramas para latencia. |
| `tracing` | Distributed tracing con OpenTelemetry, span propagation, exporters. |
| `healthcheck` | Endpoints `/health`, `/ready`, `/live` con dependency checks. |

### DevOps

| Módulo | Qué genera |
|---|---|
| `docker` | Dockerfile multi-stage optimizado, docker-compose para desarrollo. |
| `ci --provider=<github\|gitlab\|bitbucket>` | Pipeline CI/CD con lint, test, build, release. |
| `k8s` | Manifiestos Kubernetes: Deployment, Service, Ingress, HPA, ConfigMap. |
| `terraform` | Infraestructura base para el cloud provider elegido. |
| `makefile` | Makefile con targets estándar: build, test, lint, run, migrate, generate. |

### Testing

| Módulo | Qué genera |
|---|---|
| `testkit` | Helpers de testing, fixtures, factories, testcontainers setup. |
| `e2e` | Suite de tests end-to-end con setup/teardown automatizado. |
| `benchmark` | Benchmarks para hot paths con reportes comparativos. |
| `mocks` | Generación automática de mocks para interfaces con mockery/moq. |

### Seguridad

| Módulo | Qué genera |
|---|---|
| `cors` | Configuración CORS con whitelist, preflight handling. |
| `ratelimit` | Rate limiting por IP/usuario con sliding window, token bucket. |
| `encryption` | Helpers de cifrado, hashing, key management. |
| `csrf` | Protección CSRF con token validation middleware. |

---

## Flujo de Uso

### Crear un proyecto nuevo

```bash
# Interactivo — wizard paso a paso
arch_forge init

# Directo
arch_forge init myapp --arch=hexagonal --variant=modular --modules=api,database,docker,logging

# Con preset
arch_forge init myapp --preset=production-api
```

### Agregar módulos a un proyecto existente

```bash
cd myapp

# Agrega autenticación respetando la arquitectura hexagonal del proyecto
arch_forge add auth

# Genera CRUD completo para la entidad "product"
arch_forge add crud --entity=product --fields="name:string,price:float64,stock:int"

# Agrega múltiples módulos
arch_forge add database --driver=postgres cache --driver=redis queue --driver=nats
```

### Agregar un módulo de dominio a un proyecto existente

El comando `arch_forge domain add <name>` agrega un nuevo módulo de dominio (bounded context) a un proyecto existente, respetando la arquitectura declarada en `archforge.yaml`.

```bash
arch_forge domain add payment
arch_forge domain add order --dry-run
arch_forge domain add notification --project-dir ./myapp
```

**Flags disponibles:**

| Flag | Default | Descripción |
|---|---|---|
| `--project-dir` | directorio actual | Ruta al proyecto donde se agrega el dominio |
| `--dry-run` | `false` | Muestra los archivos que se generarían sin escribirlos |

**Estructura generada según arquitectura:**

La estructura de carpetas y archivos varía dependiendo de la arquitectura y variante declaradas en `archforge.yaml`:

| Arquitectura / Variante | Estructura generada |
|---|---|
| `hexagonal` / `modular` | `internal/{name}/{domain,ports/inbound,ports/outbound,application,adapters/inbound/http,adapters/outbound/postgres}` |
| `hexagonal` / `classic` | Agrega archivos a las capas existentes `internal/{domain,ports,application,adapters}` |
| `clean` / `modular` | `internal/{name}/{domain,usecase,ports,adapters/{http,postgres}}` |
| `clean` / `classic` | Agrega archivos a las capas existentes `internal/{domain,usecase,ports,adapters}` |
| `ddd` | `internal/{name}/{domain,application,infrastructure/{http,persistence/postgres}}` |
| `standard` / `modular` | `internal/{name}/{model,service,handler,repository}` |
| `modular_monolith` | `internal/{name}/{domain,service,handler}` |
| `cqrs` / `modular` | `internal/{name}/{domain,command,query,infrastructure/http}` |
| `microservice` | `internal/{name}/{domain,port,app,adapter/{http,postgres}}` |

---

### Otros comandos

```bash
# Ver la estructura del proyecto actual
arch_forge inspect

# Validar que la estructura sigue los principios de la arquitectura elegida
arch_forge doctor

# Listar arquitecturas y módulos disponibles
arch_forge list archs
arch_forge list modules

# Actualizar arch_forge a la última versión
arch_forge update
```

---

## Presets

Combinaciones predefinidas para casos de uso comunes:

| Preset | Arquitectura | Variante | Módulos incluidos |
|---|---|---|---|
| `starter` | Standard Layout | classic | api, logging, docker, makefile |
| `production-api` | Hexagonal | modular | api, database(postgres), auth, logging, metrics, tracing, healthcheck, docker, ci(github), makefile, testkit |
| `microservice` | Microservice | classic | grpc, database(postgres), queue(nats), logging, metrics, tracing, healthcheck, docker, k8s, makefile |
| `event-driven` | CQRS + Event Sourcing | modular | api, database(postgres), queue(kafka), logging, tracing, docker, makefile |
| `ddd-app` | DDD | modular | api, database(postgres), auth, logging, docker, makefile, testkit |
| `fullstack` | Modular Monolith | modular | api, graphql, database(postgres), cache(redis), auth, docker, ci(github), makefile |

---

## Archivo de Configuración — `archforge.yaml`

Cada proyecto generado contiene un archivo de configuración que arch_forge lee para entender el proyecto:

```yaml
project:
  name: myapp
  module: github.com/user/myapp
  version: 0.1.0

architecture: hexagonal
variant: modular

go:
  version: "1.23"
  linter: golangci-lint

modules:
  - name: api
    framework: chi
    port: 8080
  - name: database
    driver: postgres
    migrations: goose
  - name: auth
    strategy: jwt
  - name: logging
    library: slog
  - name: docker
  - name: makefile

```

---

## Comando `arch_forge doctor`

Analiza el proyecto y reporta violaciones arquitectónicas:

```
$ arch_forge doctor

🔍 Analyzing project structure...

Architecture: hexagonal
Modules: api, database, auth, logging

✓ Domain layer has no external imports
✓ Ports are defined as interfaces
✗ Adapter "user_repository.go" imports domain directly instead of through port
✗ Handler "order_handler.go" contains business logic — should be in service layer
✓ All use cases depend on abstractions
✓ Dependency injection configured correctly

Score: 8/10
Found 2 violations. Run `arch_forge doctor --fix` for suggestions.
```

---

## Stack Técnico para Construir arch_forge

### Core

| Tecnología | Propósito |
|---|---|
| **Go 1.23+** | Lenguaje del propio CLI. |
| **Cobra** | Framework CLI — subcommands, flags, completions. |
| **Viper** | Manejo de configuración (`archforge.yaml`). |
| **promptui / bubbletea** | UI interactiva para el wizard `init`. Terminal UI rica con bubbletea para selección de arquitectura, módulos, preview. |
| **text/template + embed** | Motor de templates con archivos embebidos en el binario. |
| **go:embed** | Templates distribuidos dentro del propio binario, zero dependencies externas al ejecutar. |

### Templates & Generación

| Tecnología | Propósito |
|---|---|
| **AST parsing (go/ast)** | Análisis estático del código generado para `doctor`. Inspección de imports y dependencias. |
| **jennifer** | Generación programática de código Go cuando los templates no son suficientes. |
| **goose / golang-migrate** | Generación de migraciones SQL. |
| **buf** | Generación de código desde proto files (módulo grpc). |
| **gqlgen** | Scaffolding de GraphQL. |

### Testing del CLI

| Tecnología | Propósito |
|---|---|
| **testify** | Assertions y test suites. |
| **testcontainers-go** | Tests de integración para módulos que requieren servicios externos. |
| **go-golden** | Snapshot testing para validar output de templates. |
| **afero** | Filesystem in-memory para tests de generación sin tocar disco. |

### Distribución

| Tecnología | Propósito |
|---|---|
| **GoReleaser** | Build multi-plataforma, changelogs automáticos, publicación. |
| **Homebrew tap** | `brew install arch_forge`. |
| **Scoop** | Instalación en Windows. |
| **Docker image** | `docker run archforge/cli init myapp`. |
| **GitHub Actions** | CI/CD del propio arch_forge. |
| **go install** | `go install github.com/archforge/cli@latest`. |

---

## Sistema de Módulos — Cómo se Crean y Generan

### Anatomía de un módulo

Cada módulo vive dentro de `templates/go/modules/<nombre>/` y tiene la siguiente estructura:

```
templates/go/modules/auth/
├── module.yaml                      # Manifiesto del módulo
├── prompts.yaml                     # Preguntas interactivas al usuario
├── base/                            # Archivos comunes a todas las arquitecturas
│   ├── migrations/
│   │   └── {{timestamp}}_create_users_table.sql.tmpl
│   └── testdata/
│       └── fixtures.go.tmpl
├── variants/                        # Archivos específicos por arquitectura + variante
│   ├── hexagonal/
│   │   ├── classic/
│   │   │   ├── port/
│   │   │   │   └── auth_service.go.tmpl
│   │   │   ├── port/
│   │   │   │   └── token_repository.go.tmpl
│   │   │   ├── app/
│   │   │   │   └── auth_service.go.tmpl
│   │   │   └── adapter/
│   │   │       ├── inbound/http/
│   │   │       │   └── auth_handler.go.tmpl
│   │   │       └── outbound/postgres/
│   │   │           └── token_repo.go.tmpl
│   │   └── modular/
│   │       ├── domain/
│   │       │   └── user.go.tmpl
│   │       ├── port/
│   │       │   ├── auth_service.go.tmpl
│   │       │   └── user_repository.go.tmpl
│   │       ├── app/
│   │       │   └── auth_service.go.tmpl
│   │       └── adapter/
│   │           ├── handler.go.tmpl
│   │           └── postgres_repo.go.tmpl
│   ├── clean/
│   │   ├── classic/
│   │   └── modular/
│   ├── ddd/
│   │   ├── classic/
│   │   └── modular/
│   └── standard/
│       ├── classic/
│       └── modular/
└── hooks/                           # Modificaciones a archivos existentes
    ├── wire.go.tmpl                 # Inyección de dependencias
    ├── router.go.tmpl               # Registro de rutas
    └── main.go.tmpl                 # Registro en main
```

---

### Manifiesto del módulo — `module.yaml`

El manifiesto define todo lo que arch_forge necesita saber sobre el módulo:

```yaml
name: auth
version: 1.0.0
description: "Authentication module with JWT and OAuth2 support"
category: core

# Arquitecturas soportadas por este módulo
architectures:
  - hexagonal
  - clean
  - ddd
  - standard
  - modular_monolith
  - cqrs
  - microservice

# Variantes soportadas
variants:
  - classic
  - modular

# Dependencias con otros módulos (se instalan automáticamente si faltan)
dependencies:
  required:
    - database                       # Necesita un repositorio de usuarios
  optional:
    - cache                          # Para almacenar sesiones en Redis
    - logging                        # Para audit logging

# Dependencias Go que el módulo necesita
go_dependencies:
  - package: github.com/golang-jwt/jwt/v5
    version: v5.2.1
  - package: golang.org/x/crypto
    version: v0.28.0
  - package: github.com/google/uuid
    version: v1.6.0

# Opciones configurables del módulo
options:
  - name: strategy
    type: enum
    values: [jwt, session, oauth2]
    default: jwt
    description: "Authentication strategy"

  - name: token_expiry
    type: duration
    default: "24h"
    description: "Token expiration time"

  - name: refresh_tokens
    type: bool
    default: true
    description: "Enable refresh token rotation"

  - name: oauth_providers
    type: list
    values: [google, github, apple, microsoft]
    default: []
    condition: "strategy == oauth2"
    description: "OAuth2 providers to configure"

# Archivos existentes que el módulo necesita modificar (hooks)
patches:
  - target: "cmd/*/main.go"
    action: inject
    anchor: "// arch_forge:modules"
    template: hooks/main.go.tmpl

  - target: "**/router.go"
    action: inject
    anchor: "// arch_forge:routes"
    template: hooks/router.go.tmpl

  - target: "**/wire.go"
    action: inject
    anchor: "// arch_forge:providers"
    template: hooks/wire.go.tmpl
    optional: true                   # No falla si wire.go no existe
```

---

### Prompts interactivos — `prompts.yaml`

Define las preguntas que el wizard le hace al usuario si ejecuta `arch_forge add auth` sin flags:

```yaml
prompts:
  - key: strategy
    type: select
    message: "Select authentication strategy"
    options:
      - label: "JWT (stateless tokens)"
        value: jwt
        description: "Best for APIs and microservices"
      - label: "Session-based (server-side)"
        value: session
        description: "Best for traditional web apps"
      - label: "OAuth2 (third-party providers)"
        value: oauth2
        description: "Login with Google, GitHub, etc."
    default: jwt

  - key: refresh_tokens
    type: confirm
    message: "Enable refresh token rotation?"
    default: true
    condition: "strategy == jwt"

  - key: oauth_providers
    type: multiselect
    message: "Select OAuth2 providers"
    options: [Google, GitHub, Apple, Microsoft]
    condition: "strategy == oauth2"
    min: 1
```

---

### Templates — El motor de generación

Los templates usan `text/template` de Go con funciones helper adicionales registradas por arch_forge:

```go
// Ejemplo: templates/go/modules/auth/variants/hexagonal/modular/port/auth_service.go.tmpl

package port

import (
    "context"

    "{{ .Module }}/internal/{{ .ModuleName }}/domain"
)

// AuthService defines the driving port for authentication operations.
type AuthService interface {
    Register(ctx context.Context, req RegisterRequest) (*domain.User, error)
    Login(ctx context.Context, req LoginRequest) (*AuthTokens, error)
{{- if .Options.refresh_tokens }}
    RefreshToken(ctx context.Context, refreshToken string) (*AuthTokens, error)
{{- end }}
    Logout(ctx context.Context, userID string) error
}

type RegisterRequest struct {
    Email    string
    Password string
    Name     string
}

type LoginRequest struct {
    Email    string
    Password string
}

type AuthTokens struct {
    AccessToken  string
    TokenType    string
    ExpiresIn    int
{{- if .Options.refresh_tokens }}
    RefreshToken string
{{- end }}
}
```

**Funciones helper disponibles en templates:**

```go
// template_funcs.go — funciones registradas en el engine

var templateFuncs = template.FuncMap{
    // Naming
    "camelCase":    toCamelCase,     // orderItem → orderItem
    "pascalCase":   toPascalCase,    // orderItem → OrderItem
    "snakeCase":    toSnakeCase,     // OrderItem → order_item
    "kebabCase":    toKebabCase,     // OrderItem → order-item
    "plural":       toPlural,        // order → orders
    "singular":     toSingular,      // orders → order

    // Types
    "goType":       toGoType,        // string → string, uuid → uuid.UUID
    "sqlType":      toSQLType,       // string → VARCHAR(255), uuid → UUID
    "zeroValue":    toZeroValue,     // string → "", int → 0, bool → false
    "isNullable":   isNullable,      // *string → true

    // Paths — resuelven la ruta correcta según arquitectura + variante
    "domainPath":   resolveDomainPath,    // → "internal/domain" o "internal/order/domain"
    "portPath":     resolvePortPath,      // → "internal/port" o "internal/order/port"
    "adapterPath":  resolveAdapterPath,   // → "internal/adapter/..." o "internal/order/adapter/..."
    "servicePath":  resolveServicePath,

    // Imports
    "import":       resolveImport,        // Genera import path correcto para el módulo Go

    // Conditional
    "hasModule":    projectHasModule,     // Chequea si otro módulo está instalado
    "ifModule":     ifModuleInstalled,    // Bloque condicional basado en módulo

    // Timestamps
    "timestamp":    generateTimestamp,    // Para nombres de migraciones
    "now":          timeNow,
}
```

---

### Pipeline de generación — Qué pasa cuando ejecutás `arch_forge add auth`

```
┌─────────────────────────────────────────────────────────────────┐
│                     arch_forge add auth                         │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
               ┌───────────────────────┐
           1.  │  Leer archforge.yaml  │  Detectar arquitectura, variante,
               │                       │  módulo Go, módulos ya instalados
               └───────────┬───────────┘
                           │
                           ▼
               ┌───────────────────────┐
           2.  │  Cargar module.yaml   │  Leer manifiesto de auth:
               │                       │  dependencias, opciones, patches
               └───────────┬───────────┘
                           │
                           ▼
               ┌───────────────────────┐
           3.  │  Resolver dependencias│  ¿Falta "database"? → instalarlo
               │                       │  primero (recursivo)
               └───────────┬───────────┘
                           │
                           ▼
               ┌───────────────────────┐
           4.  │  Prompts / flags      │  Si es interactivo → prompts.yaml
               │                       │  Si tiene flags → validar opciones
               └───────────┬───────────┘
                           │
                           ▼
               ┌───────────────────────┐
           5.  │  Construir contexto   │  TemplateContext con toda la info:
               │  de template          │  proyecto, arquitectura, variante,
               │                       │  opciones, módulos existentes
               └───────────┬───────────┘
                           │
                           ▼
               ┌───────────────────────┐
           6.  │  Seleccionar archivos │  base/ + variants/{arch}/{variant}/
               │  de template          │  = lista de archivos a generar
               └───────────┬───────────┘
                           │
                           ▼
               ┌───────────────────────┐
           7.  │  Renderizar templates │  text/template.Execute() con el
               │                       │  contexto → código Go final
               └───────────┬───────────┘
                           │
                           ▼
               ┌───────────────────────┐
           8.  │  Calcular paths de    │  Resolver dónde va cada archivo
               │  destino              │  según arquitectura + variante
               └───────────┬───────────┘
                           │
                           ▼
               ┌───────────────────────┐
           9.  │  Aplicar patches      │  Modificar archivos existentes:
               │  (hooks)              │  main.go, router.go, wire.go
               └───────────┬───────────┘
                           │
                           ▼
               ┌───────────────────────┐
          10.  │  Dry-run / Preview    │  Si --dry-run → mostrar diff
               │                       │  Si no → pedir confirmación
               └───────────┬───────────┘
                           │
                           ▼
               ┌───────────────────────┐
          11.  │  Escribir a disco     │  Crear archivos nuevos,
               │                       │  aplicar patches a existentes
               └───────────┬───────────┘
                           │
                           ▼
               ┌───────────────────────┐
          12.  │  Post-generación      │  go mod tidy, go fmt,
               │                       │  actualizar archforge.yaml,
               │                       │  mostrar resumen
               └───────────────────────┘
```

---

### Contexto de template — `TemplateContext`

El objeto que reciben todos los templates durante la renderización:

```go
type TemplateContext struct {
    // Proyecto
    Project     ProjectInfo          // Nombre, módulo Go, versión
    Arch        string               // "hexagonal", "clean", "ddd", etc.
    Variant     string               // "classic", "modular"

    // Módulo actual
    ModuleName  string               // "auth"
    Options     map[string]any       // Opciones elegidas por el usuario

    // Estado del proyecto
    Modules     []string             // Módulos ya instalados: ["api", "database"]
    GoVersion   string               // "1.23"

    // Para módulo CRUD — info de la entidad
    Entity      *EntityInfo          // Nombre, fields, relaciones

    // Helpers pre-calculados
    Module      string               // "github.com/user/myapp" (Go module path)
    Paths       ResolvedPaths        // Paths resueltos para esta arch+variant
}

type ProjectInfo struct {
    Name    string                   // "myapp"
    Module  string                   // "github.com/user/myapp"
    Version string                   // "0.1.0"
}

type EntityInfo struct {
    Name       string                // "Order"
    Fields     []Field               // [{Name: "Total", Type: "float64", ...}]
    Relations  []Relation            // [{Type: "belongs_to", Target: "Customer"}]
}

type Field struct {
    Name       string                // "Total"
    Type       string                // "float64"
    GoType     string                // "float64"
    SQLType    string                // "DECIMAL(10,2)"
    JSONName   string                // "total"
    DBName     string                // "total"
    Nullable   bool
    Validation string                // "required,gt=0"
}

type ResolvedPaths struct {
    Domain     string                // "internal/domain" o "internal/order/domain"
    Port       string                // "internal/port" o "internal/order/port"
    App        string                // "internal/app" o "internal/order/app"
    Adapter    string                // "internal/adapter" o "internal/order/adapter"
    Handler    string                // "internal/adapter/inbound/http" o "internal/order/adapter"
    Repository string                // "internal/adapter/outbound/postgres" o "internal/order/adapter"
    Migration  string                // "migrations/"
    Test       string                // Mirrors source path + _test.go
}
```

---

### Sistema de Patches — Cómo se modifican archivos existentes

El problema más difícil de un generador es **modificar archivos que ya existen** sin romperlos. arch_forge usa un sistema de **anchors** (comentarios marcador) que se insertan durante `init` y que los módulos buscan para inyectar código:

```go
// main.go generado por `arch_forge init` incluye anchors:

func main() {
    cfg := config.Load()
    db := database.Connect(cfg.DatabaseURL)

    // arch_forge:providers — modules inject their providers here

    router := chi.NewRouter()
    router.Use(middleware.Logger)

    // arch_forge:routes — modules inject their routes here

    // arch_forge:shutdown — modules inject graceful shutdown here

    server.Start(router, cfg.Port)
}
```

Cuando `arch_forge add auth` aplica su patch `hooks/main.go.tmpl`:

```go
// hooks/main.go.tmpl

// patch:anchor: arch_forge:providers
authRepo := postgresauth.NewRepository(db)
authService := authapp.NewService(authRepo, cfg.Auth)

// patch:anchor: arch_forge:routes
router.Mount("/auth", authhandler.NewRouter(authService))

// patch:anchor: arch_forge:shutdown
authService.Close()
```

**Resultado después del patch:**

```go
func main() {
    cfg := config.Load()
    db := database.Connect(cfg.DatabaseURL)

    // arch_forge:providers — modules inject their providers here
    authRepo := postgresauth.NewRepository(db)
    authService := authapp.NewService(authRepo, cfg.Auth)

    router := chi.NewRouter()
    router.Use(middleware.Logger)

    // arch_forge:routes — modules inject their routes here
    router.Mount("/auth", authhandler.NewRouter(authService))

    // arch_forge:shutdown — modules inject graceful shutdown here
    authService.Close()

    server.Start(router, cfg.Port)
}
```

Los imports se agregan automáticamente usando `go/ast` para parsear el archivo, agregar los imports necesarios, y reescribirlo con `go/format`.

---

### Crear módulos custom (local)

Un developer puede crear módulos propios para su proyecto:

```bash
# Scaffolding de un módulo nuevo
arch_forge module create my-payments

# Genera:
# templates/custom/modules/my-payments/
# ├── module.yaml        (manifiesto pre-llenado)
# ├── prompts.yaml       (vacío, listo para llenar)
# ├── base/
# ├── variants/
# │   └── hexagonal/     (genera solo la arch del proyecto actual como ejemplo)
# │       ├── classic/
# │       └── modular/
# └── hooks/

# Desarrollar el módulo probando contra el proyecto actual
arch_forge module dev my-payments

# Validar que el módulo es correcto
arch_forge module validate my-payments
```

**`arch_forge module validate`** verifica:

- `module.yaml` tiene todos los campos requeridos
- Todos los templates compilan sin errores de sintaxis
- Los paths resueltos no colisionan con otros módulos
- Las dependencias Go son válidas y resolubles
- Los anchors referenciados en patches existen en los templates de `init`
- Al menos una combinación de arquitectura + variante está soportada

> **Futuro**: en versiones posteriores se habilitará `arch_forge module publish` para compartir módulos al marketplace y `arch_forge template import/export` para cargar configuraciones desde templates externos.

---

## Características Diferenciadores

### 1. Architecture-Aware Code Generation

Cada módulo no genera código genérico. Genera código **adaptado a la arquitectura elegida**. El módulo `auth` en hexagonal genera puertos e interfaces; en clean architecture genera use cases y entities; en standard layout genera paquetes en `internal/`.

### 2. Live Architecture Validation (`doctor`)

No solo genera — vigila. El comando `doctor` actúa como un linter arquitectónico que detecta cuando el código se desvía de los principios de la arquitectura elegida.

### 3. Dependency Graph Visualization

```bash
arch_forge graph

# Genera un diagrama de dependencias entre capas/módulos
# Output: SVG, Mermaid, o abre en browser
```

Usa **graphviz** o genera diagramas **Mermaid** para visualizar las dependencias entre paquetes y validar que respetan la dirección permitida por la arquitectura.

### 4. Smart Scaffolding con Fields

```bash
arch_forge add crud --entity=order \
  --fields="customer_id:uuid,total:decimal,status:enum(pending,paid,shipped),items:[]OrderItem" \
  --relations="belongs_to:customer,has_many:order_items"
```

Genera modelo, validaciones, handlers, servicio, repositorio, migraciones SQL, tests y documentación OpenAPI. Todo en la capa correcta.

### 5. Interactive Diff Preview

Antes de escribir archivos, arch_forge muestra un diff interactivo de qué va a crear/modificar:

```bash
arch_forge add auth --dry-run

# Muestra:
# + internal/auth/port/auth_service.go
# + internal/auth/port/token_service.go
# + internal/auth/adapter/jwt_token_service.go
# + internal/auth/adapter/handler/auth_handler.go
# ~ cmd/api/main.go (modified: adds auth middleware registration)
# + migrations/20260326_create_users_table.sql
#
# Apply changes? [Y/n/diff]
```

### 6. Multi-Language Ready (Arquitectura interna)

La arquitectura interna de arch_forge está diseñada para ser language-agnostic. Los templates y la lógica de generación están separados por lenguaje, preparando el camino para soporte futuro:

```
templates/
├── go/              # v1
│   ├── hexagonal/
│   ├── clean/
│   └── modules/
├── rust/            # futuro
├── typescript/      # futuro
└── python/          # futuro
```

> **v1 solo soporta Go.** El soporte multi-lenguaje se habilitará en versiones posteriores.

---

## Roadmap

### v0.1 — MVP

- [ ] Comando `init` con wizard interactivo
- [ ] Arquitecturas: Standard Layout, Hexagonal, Clean Architecture
- [ ] Módulos core: api (chi), database (postgres), logging (slog), docker, makefile
- [ ] Archivo `archforge.yaml`
- [ ] Comando `add` para módulos post-creación
- [ ] Comando `list` para listar arquitecturas y módulos
- [ ] Tests con snapshot testing

### v0.2 — Developer Experience

- [ ] Comando `doctor` — validación arquitectónica básica
- [ ] Comando `inspect` — visualización de estructura
- [ ] Terminal UI mejorada con bubbletea
- [ ] Módulos: auth (JWT), crud scaffolding, grpc, cache (redis)
- [ ] Presets: starter, production-api, microservice
- [ ] Shell completions (bash, zsh, fish, powershell)
- [ ] Comando `graph` — dependency visualization

### v0.3 — Ecosystem

- [ ] Arquitecturas: DDD, CQRS, Modular Monolith
- [ ] Módulos: queue, storage, search, metrics, tracing, healthcheck
- [ ] Comando `doctor --fix` con auto-corrección
- [ ] CI module (GitHub Actions, GitLab CI)
- [ ] Kubernetes manifests module

### v1.0 — Production Ready

- [ ] Documentación completa con ejemplos
- [ ] Estabilidad de API de templates
- [ ] Homebrew, Scoop, Docker distribution
- [ ] `arch_forge module validate` para módulos custom locales

### v2.0 — Ecosystem Abierto

- [ ] Templates remotos — import/export
- [ ] Plugin system (nuevas arquitecturas, módulos, validaciones)
- [ ] Template marketplace (publish, search, trending)
- [ ] Community templates curados

### v3.0 — Multi-Language

- [ ] Soporte para Rust
- [ ] Soporte para TypeScript/Node.js
- [ ] Language-specific `doctor` rules

---

## Estructura del Proyecto arch_forge (Dogfooding)

arch_forge usará su propia arquitectura hexagonal:

```
arch_forge/
├── cmd/
│   └── archforge/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── architecture.go      # Architecture, Module, Template entities
│   │   ├── project.go           # Project aggregate
│   │   └── validation.go        # Architecture rules
│   ├── port/
│   │   ├── generator.go         # Port: code generation
│   │   ├── analyzer.go          # Port: code analysis (doctor)
│   │   └── template_repo.go     # Port: template storage
│   ├── app/
│   │   ├── init_project.go      # Use case: initialize project
│   │   ├── add_module.go        # Use case: add module
│   │   ├── diagnose.go          # Use case: run doctor
│   │   └── inspect.go           # Use case: inspect project
│   └── adapter/
│       ├── cli/
│       │   ├── root.go          # Cobra root command
│       │   ├── init.go          # init subcommand
│       │   ├── add.go           # add subcommand
│       │   ├── doctor.go        # doctor subcommand
│       │   └── tui/             # bubbletea components
│       ├── generator/
│       │   └── template_engine.go
│       ├── analyzer/
│       │   └── ast_analyzer.go
│       ├── repository/
│       │   └── local_templates.go
├── templates/
│   └── go/
│       ├── hexagonal/
│       ├── clean/
│       ├── ddd/
│       ├── standard/
│       ├── cqrs/
│       ├── modular_monolith/
│       ├── microservice/
│       └── modules/
│           ├── api/
│           ├── auth/
│           ├── crud/
│           ├── database/
│           ├── docker/
│           └── ...
├── archforge.yaml
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── .goreleaser.yaml
├── .golangci.yml
└── README.md
```

---

## Métricas de Éxito

| Métrica | Target (6 meses post-launch) |
|---|---|
| GitHub stars | 1,000+ |
| Descargas mensuales | 5,000+ |
| Arquitecturas soportadas | 7+ |
| Módulos built-in disponibles | 25+ |
| Contributors | 20+ |

---

## Competidores y Diferenciación

| Herramienta | Limitación | Ventaja arch_forge |
|---|---|---|
| `go-blueprint` | Solo genera estructura inicial, no entiende arquitecturas. | Architecture-aware, módulos post-creación, doctor. |
| `cookiecutter` | Templates estáticos sin contexto arquitectónico. | Templates dinámicos que adaptan output a la arquitectura. |
| `goxygen` | Solo fullstack Go+frontend, una sola estructura. | Múltiples arquitecturas, módulos composables. |
| `micro` | Solo microservicios, framework-locked. | Agnóstico de framework, múltiples patrones. |

---

## Principios de Diseño

1. **Convention over Configuration**: Defaults inteligentes. Cero config para empezar, config granular para personalizar.
2. **Architecture First**: Cada decisión de generación está informada por la arquitectura elegida.
3. **Composable**: Los módulos son independientes y combinables. Agregar uno no rompe otro.
4. **No Lock-in**: El código generado es tuyo. No hay runtime dependency con arch_forge.
5. **Idiomatic**: El código generado sigue las convenciones idiomáticas del lenguaje (effective Go, proverbs).
6. **Testable by Default**: Todo lo generado incluye tests. La estructura facilita testing en todos los niveles.
7. **Progressive Disclosure**: Simple para empezar, poderoso cuando lo necesitas.
