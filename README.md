# arch_forge

> Generate production-ready Go project structures in seconds

![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-green?style=flat)
![Release](https://img.shields.io/github/v/release/archforge/cli?style=flat)
![Build](https://img.shields.io/github/actions/workflow/status/archforge/cli/release.yml?style=flat)

---

## Overview

arch_forge is a CLI tool that generates complete, production-ready Go project skeletons based on proven architectural patterns. Instead of spending days setting up project structure, wiring boilerplate, and configuring tooling, you describe what you want and arch_forge builds it in seconds.

The tool is designed for Go developers who care about architecture. Whether you are starting a new microservice, a modular monolith, or a domain-driven application, arch_forge scaffolds the right structure with the right conventions from day one. It does not just create folders — it generates real, working code organized according to the chosen architectural style.

What sets arch_forge apart is its `doctor` command, which validates your project's architecture compliance over time. As your codebase grows, arch_forge can detect layer violations, misplaced dependencies, and structural drift — keeping your architecture honest. The module system (25 built-in modules, extensible with custom local modules) means you can add infrastructure, observability, security, and DevOps tooling without writing boilerplate from scratch.

---

## Installation

### Homebrew (macOS / Linux)

```bash
brew install archforge/tap/arch-forge
```

### Go install

```bash
go install github.com/archforge/cli/cmd/archforge@latest
```

### Docker

```bash
docker run --rm -v $(pwd):/workspace archforge/cli:latest init myapp --arch hexagonal
```

### Scoop (Windows)

```powershell
scoop bucket add archforge https://github.com/archforge/scoop-bucket
scoop install arch-forge
```

### Build from source

Requirements: Go 1.25+, Git.

```bash
git clone https://github.com/archforge/cli.git
cd cli
make build
```

The binary is placed at `./bin/arch_forge`. To make it available globally:

```bash
# Option A — install via go install
go install ./cmd/archforge

# Option B — copy to a directory in your PATH
cp ./bin/arch_forge /usr/local/bin/arch_forge
```

Verify the installation:

```bash
arch_forge --version
```

---

## Quick Start

### Interactive wizard (recommended for new users)

```bash
arch_forge init myapp
# launches TUI wizard to choose architecture, variant, and modules
```

### One-liner with flags

```bash
arch_forge init myapp --arch hexagonal --variant classic --modules api,database,logging,docker
```

### Using a preset

```bash
arch_forge init myapp --preset production-api
```

### Generated project structure

After running `arch_forge init myapp --arch hexagonal --variant classic`, the generated project looks like this:

```
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/              # Business logic — no external deps
│   ├── ports/
│   │   ├── inbound/         # Driving ports (use case interfaces)
│   │   └── outbound/        # Driven ports (repository interfaces)
│   ├── application/         # Use cases — orchestrate domain + ports
│   └── adapters/
│       ├── inbound/
│       │   └── http/        # HTTP handlers
│       └── outbound/
│           └── postgres/    # Database implementations
├── go.mod
├── archforge.yaml           # Project config
└── .gitignore
```

With `--variant modular`, code is organized by bounded context instead of technical layer:

```
myapp/
├── cmd/api/main.go
├── internal/
│   ├── platform/            # Shared config and infrastructure
│   └── user/                # Bounded context
│       ├── domain/
│       ├── ports/
│       │   ├── inbound/
│       │   └── outbound/
│       ├── application/
│       └── adapters/
│           ├── inbound/http/
│           └── outbound/postgres/
├── go.mod
└── archforge.yaml
```

---

## Architectures

arch_forge supports 7 architectural patterns. Choose the one that matches your project's complexity, team size, and domain requirements.

| Architecture | Flag | Best For | Classic structure | Modular structure |
|---|---|---|---|---|
| Hexagonal | `--arch hexagonal` | APIs where adapters must be swappable | `domain/`, `ports/`, `application/`, `adapters/` | `internal/{name}/{domain,ports,application,adapters}` |
| Clean Architecture | `--arch clean` | Large teams needing strict layer discipline | `domain/`, `usecase/`, `ports/`, `adapters/` | `internal/{name}/{domain,usecase,ports,adapters}` |
| Domain-Driven Design | `--arch ddd` | Complex domains with bounded contexts | `internal/{name}/{domain,application,infrastructure}` | `internal/{name}/{domain,application,infrastructure}` |
| Standard Layout | `--arch standard` | Simple services, scripts, CLIs | `model/`, `service/`, `handler/`, `repository/` | `internal/{name}/{model,service,handler,repository}` |
| Modular Monolith | `--arch modular_monolith` | Monoliths designed for future extraction | `internal/module/{name}/` | `internal/{name}/{domain,service,handler}` |
| CQRS + Event Sourcing | `--arch cqrs` | Audit-heavy, event-driven systems | `domain/`, `command/`, `query/`, `infrastructure/` | `internal/{name}/{domain,command,query,infrastructure}` |
| Microservice | `--arch microservice` | Single production microservices | `domain/`, `port/`, `app/`, `adapter/` | `internal/{name}/{domain,port,app,adapter}` |

Each architecture supports two variants:

- `--variant classic` — by-the-book, canonical layer naming as described in the original literature
- `--variant modular` — business-module-first organization, grouping all layers for each bounded context under `internal/{name}/`

---

## Modules

Modules are optional units of functionality that can be added to any project at init time or later. Each module generates real, working code integrated with your chosen architecture.

### Adding modules to an existing project

```bash
arch_forge add database logging metrics
```

### Available modules

#### Core

| Module | Description |
|---|---|
| `api` | REST API with router, middleware chain, error handling, and request validation |

#### Infrastructure

| Module | Description |
|---|---|
| `database` | Database connection, pool, migrations, and health check |
| `cache` | Redis client with cache-aside pattern, TTL management, and connection pooling |
| `queue` | Message queue client supporting RabbitMQ, NATS, and in-memory backends |
| `storage` | Object storage client supporting S3, GCS, and local filesystem |
| `search` | Full-text search client supporting Elasticsearch and Meilisearch |

#### Observability

| Module | Description |
|---|---|
| `logging` | Structured logging with slog, log levels, and correlation IDs |
| `metrics` | Prometheus metrics: request counters, latency histograms, /metrics endpoint |
| `tracing` | OpenTelemetry distributed tracing with OTLP exporter |
| `healthcheck` | Liveness, readiness, and health HTTP endpoints for Kubernetes |

#### DevOps

| Module | Description |
|---|---|
| `docker` | Multi-stage Dockerfile and docker-compose for development |
| `makefile` | Makefile with standard targets: build, test, lint, run, migrate, generate |
| `ci` | CI/CD pipeline for GitHub Actions, GitLab CI, or Bitbucket Pipelines |
| `k8s` | Kubernetes manifests: Deployment, Service, Ingress, HPA, ConfigMap |
| `terraform` | Terraform infrastructure for AWS, GCP, or Azure |

#### Testing

| Module | Description |
|---|---|
| `testkit` | Test fixtures, factories, and testcontainers for integration testing |
| `e2e` | End-to-end test scaffold with HTTP test client |
| `mocks` | Mock generation with mockery from interfaces |

#### Security

| Module | Description |
|---|---|
| `auth` | JWT authentication middleware, token generation and validation |
| `cors` | CORS middleware with configurable allowed origins and methods |
| `ratelimit` | Token-bucket rate limiting middleware per-IP and global |
| `encryption` | AES-GCM symmetric encryption and PBKDF2 key derivation |

#### API

| Module | Description |
|---|---|
| `grpc` | gRPC server setup with buf configuration and interceptors |

#### Scaffold

| Module | Description |
|---|---|
| `crud` | Full CRUD scaffolding for an entity — handler, service, repository, migration |

---

## Adding Domain Modules

The `domain add` command scaffolds a new bounded-context module inside an existing project, respecting the architecture declared in `archforge.yaml`. Unlike `add` (which installs infrastructure capabilities like databases or auth), `domain add` generates business domain code: entities, use cases, ports, and adapters.

```bash
arch_forge domain add payment
arch_forge domain add order --dry-run
arch_forge domain add notification --project-dir ./myapp
```

The generated structure depends on the project's architecture. For a `hexagonal/modular` project:

```
internal/
└── payment/
    ├── domain/
    │   ├── payment.go           # Core entity + business rules
    │   └── errors.go            # Domain error sentinels
    ├── ports/
    │   ├── inbound/
    │   │   └── service.go       # Use case interface (driving port)
    │   └── outbound/
    │       └── repository.go    # Persistence interface (driven port)
    ├── application/
    │   └── service.go           # Application service implementing the inbound port
    └── adapters/
        ├── inbound/
        │   └── http/
        │       └── handler.go   # HTTP handler
        └── outbound/
            └── postgres/
                └── repository.go # Postgres implementation
```

For a `ddd` project (`arch_forge domain add payment` on a DDD project):

```
internal/
└── payment/
    ├── domain/
    │   ├── payment.go
    │   ├── payment_repository.go  # Repository interface lives in domain
    │   └── errors.go
    ├── application/
    │   ├── create_payment.go
    │   └── get_payment.go
    └── infrastructure/
        ├── http/
        │   └── payment_handler.go
        └── persistence/
            └── postgres/
                └── payment_repository.go
```

| Flag | Description |
|---|---|
| `--project-dir` | Project directory (default: current directory) |
| `--dry-run` | Preview files that would be created without writing to disk |

---

## Presets

Presets are curated combinations of architecture, variant, and modules for common project types.

| Preset | Architecture | Modules Included |
|---|---|---|
| `starter` | Standard Layout (modular) | api, logging, docker, makefile |
| `production-api` | Hexagonal (classic) | api, database, auth, logging, docker, makefile |
| `microservice` | Microservice (classic) | grpc, database, logging, docker, makefile |

```bash
arch_forge init myservice --preset microservice
```

---

## Commands Reference

```
arch_forge init <name>            Create a new project
arch_forge add <module...>        Add infrastructure/capability modules to a project
arch_forge domain add <name>      Add a new bounded-context domain to a project
arch_forge list archs             List supported architectures
arch_forge list modules           List available modules
arch_forge list presets           List available presets
arch_forge doctor                 Validate architecture compliance
arch_forge inspect                Show annotated project structure
arch_forge graph                  Generate a dependency diagram (Mermaid or DOT)
arch_forge module create <name>   Scaffold a new local module
arch_forge module validate <name> Validate a local module
arch_forge module dev <name>      Watch mode for module development
arch_forge update                 Update arch_forge to the latest version
arch_forge completion <shell>     Generate shell completion scripts
```

### Key flags for `init`

```
--arch          Architecture pattern (default: hexagonal)
--variant       Architecture variant: classic | modular (default: modular)
--modules       Comma-separated list of modules to include
--preset        Named preset (starter, production-api, microservice)
--module-path   Go module path (default: github.com/your-org/{name})
--go-version    Go version (default: 1.23)
--dry-run       Preview generated files without writing to disk
```

---

## Architecture Compliance

The `doctor` command analyzes your project against the rules of its declared architecture and reports violations.

```bash
arch_forge doctor
```

Example output:

```
arch_forge doctor — Analyzing ./myapp
Architecture: hexagonal (classic)

ERRORS (1)
  internal/domain/user.go:12    domain-no-adapter    domain imports adapter package

Score: 8.5 / 10.0  [threshold: 7.0]
✗ Architecture compliance check failed
```

Use `--fix` to receive suggestions on how to resolve each violation. Use `--threshold` to tune the minimum passing score for CI pipelines:

```bash
arch_forge doctor --threshold 9.0
arch_forge doctor --fix
```

---

## Dependency Graph

The `graph` command generates a visual dependency diagram of your project's internal packages.

```bash
# Output Mermaid diagram to stdout
arch_forge graph

# Save DOT format to file
arch_forge graph --format dot --output graph.dot

# Include external dependencies
arch_forge graph --include-external
```

The Mermaid output can be embedded directly into GitHub Markdown files and rendered inline.

---

## Shell Completions

```bash
# Bash
arch_forge completion bash > ~/.bash_completion.d/arch_forge

# Zsh
arch_forge completion zsh > "${fpath[1]}/_arch_forge"

# Fish
arch_forge completion fish > ~/.config/fish/completions/arch_forge.fish

# PowerShell
arch_forge completion powershell | Out-String | Invoke-Expression
```

---

## Custom Modules

The module system is extensible. You can create local modules that follow the same structure as built-in modules and share them across projects in your organization.

```bash
# Scaffold a new local module
arch_forge module create mymodule --category infrastructure

# Develop with live validation feedback
arch_forge module dev mymodule

# Validate the module before use
arch_forge module validate mymodule
```

Each module is defined by a `module.yaml` manifest:

```yaml
name: mymodule
category: infrastructure
description: Custom infrastructure module
version: "1.0.0"

dependencies:
  - api

supported_archs:
  - hexagonal
  - clean
  - ddd

options:
  backend:
    type: select
    choices: [redis, memcached]
    default: redis
```

Templates live under `base/` for files common to all architectures, and `variants/{arch}/{variant}/` for architecture-specific files. Use `// arch_forge:<anchor>` comments in existing files to define injection points for hooks.

---

## Self-Update

```bash
arch_forge update           # Check and install latest release
arch_forge update --force   # Force reinstall even if already up to date
```

---

## Project Config

After running `init`, arch_forge writes an `archforge.yaml` file to the project root. This file records the project's architecture decisions and is used by `doctor`, `add`, and `inspect` commands.

```yaml
project:
  name: myapp
  module: github.com/acme/myapp

architecture: hexagonal
variant: classic

go:
  version: "1.23"

modules:
  - api
  - database
  - logging
```

> Do not delete `archforge.yaml`. It is required for all post-init commands to function correctly.

---

## Contributing

Contributions are welcome. To get started:

1. Fork and clone the repository.
2. Run `make test` to execute the full test suite.
3. Run `make lint` to check for linting issues.
4. Open a pull request with a clear description of the change.

Please follow standard Go conventions (Effective Go, Go Proverbs) and ensure all new code includes appropriate tests. Error messages must wrap underlying errors using `fmt.Errorf("context: %w", err)`.

---

## License

MIT — see [LICENSE](./LICENSE) for details.
