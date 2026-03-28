# arch_forge

> Generate production-ready Go project structures in seconds

![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-green?style=flat)
![Release](https://img.shields.io/github/v/release/arch-forge/cli?style=flat)
![Build](https://img.shields.io/github/actions/workflow/status/arch-forge/cli/release.yml?style=flat)

---

## What is arch_forge?

arch_forge is a CLI tool that generates complete, production-ready Go project skeletons based on proven architectural patterns. Instead of spending days setting up project structure, wiring boilerplate, and configuring tooling, you describe what you want and arch_forge builds it in seconds.

The tool is designed for Go developers who care about architecture. Whether you are starting a new microservice, a modular monolith, or a domain-driven application, arch_forge scaffolds the right structure with the right conventions from day one. It does not just create folders — it generates real, working code organized according to the chosen architectural style.

What sets arch_forge apart is its `doctor` command, which validates architecture compliance over time. As your codebase grows, arch_forge can detect layer violations, misplaced dependencies, and structural drift — keeping your architecture honest. The module system (25 built-in modules, extensible with custom local modules) means you can add infrastructure, observability, security, and DevOps tooling without writing boilerplate from scratch.

---

## Installation

**curl** (macOS / Linux)

```bash
curl -sSfL https://raw.githubusercontent.com/arch-forge/cli/main/install.sh | sh
```

Install a specific version or to a custom directory:

```bash
# Specific version
curl -sSfL https://raw.githubusercontent.com/arch-forge/cli/main/install.sh | sh -s -- v1.0.0

# Custom install directory (no sudo required)
INSTALL_DIR=$HOME/.local/bin curl -sSfL https://raw.githubusercontent.com/arch-forge/cli/main/install.sh | sh
```

**Homebrew** (macOS / Linux)

```bash
brew install arch-forge/tap/arch-forge
```

**Go install**

```bash
go install github.com/arch-forge/cli/cmd/archforge@latest
```

**Scoop** (Windows)

```powershell
scoop bucket add arch-forge https://github.com/arch-forge/scoop-bucket
scoop install arch-forge
```

**Build from source** (requires Go 1.23+, Git)

```bash
git clone https://github.com/arch-forge/cli.git
cd cli
make build
# binary is placed at ./bin/arch_forge

# install globally — pick one:
go install ./cmd/archforge
cp ./bin/arch_forge /usr/local/bin/arch_forge
```

Verify the installation:

```bash
arch_forge --version
```

---

## Quick Start

**Interactive wizard** — recommended for new users

```bash
arch_forge init myapp
# launches TUI wizard: choose architecture, variant, and modules
```

**One-liner with flags**

```bash
arch_forge init myapp --arch hexagonal --variant classic --modules api,database,logging,docker
```

**Using a preset**

```bash
arch_forge init myapp --preset production-api
```

### Generated structure

`arch_forge init myapp --arch hexagonal --variant classic`:

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
├── archforge.yaml
└── .gitignore
```

`arch_forge init myapp --arch hexagonal --variant modular` — same pattern, organized by bounded context:

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

| Architecture | Flag | Best For |
|---|---|---|
| Hexagonal | `--arch hexagonal` | APIs where adapters must be swappable |
| Clean Architecture | `--arch clean` | Large teams needing strict layer discipline |
| Domain-Driven Design | `--arch ddd` | Complex domains with rich bounded contexts |
| Standard Layout | `--arch standard` | Simple services, scripts, and CLIs |
| Modular Monolith | `--arch modular_monolith` | Monoliths designed for future extraction |
| CQRS + Event Sourcing | `--arch cqrs` | Audit-heavy, event-driven systems |
| Microservice | `--arch microservice` | Single production microservices |

### Classic vs Modular variants

Every architecture supports two variants controlled by `--variant`:

- **`classic`** — canonical layer naming as described in the original literature. Code is organized by technical layer across the entire project.
- **`modular`** — business-module-first organization. All layers for each bounded context live together under `internal/{name}/`.

**Example — Hexagonal classic** (layers first):

```
internal/
├── domain/
├── ports/
├── application/
└── adapters/
```

**Example — Hexagonal modular** (domains first, layers inside):

```
internal/
├── platform/
├── user/
│   ├── domain/
│   ├── ports/
│   ├── application/
│   └── adapters/
└── payment/
    ├── domain/
    ├── ports/
    ├── application/
    └── adapters/
```

Use `classic` when the team is small and the domain is simple. Use `modular` when you expect to grow multiple bounded contexts that may eventually need to be extracted.

---

## Modules

Modules are optional units of functionality added at init time or later. Each module generates real, working code integrated with your chosen architecture.

```bash
# add modules to an existing project
arch_forge add database logging metrics
```

### Available modules

| Category | Module | Description |
|---|---|---|
| Core | `api` | REST API with router, middleware chain, error handling, and request validation |
| | | |
| Infrastructure | `database` | Database connection, pool, migrations, and health check |
| Infrastructure | `cache` | Redis client with cache-aside pattern, TTL management, and connection pooling |
| Infrastructure | `queue` | Message queue client supporting RabbitMQ, NATS, and in-memory backends |
| Infrastructure | `storage` | Object storage client supporting S3, GCS, and local filesystem |
| Infrastructure | `search` | Full-text search client supporting Elasticsearch and Meilisearch |
| | | |
| Observability | `logging` | Structured logging with slog, log levels, and correlation IDs |
| Observability | `metrics` | Prometheus metrics: request counters, latency histograms, /metrics endpoint |
| Observability | `tracing` | OpenTelemetry distributed tracing with OTLP exporter |
| Observability | `healthcheck` | Liveness, readiness, and health HTTP endpoints for Kubernetes |
| | | |
| DevOps | `docker` | Multi-stage Dockerfile and docker-compose for development |
| DevOps | `makefile` | Makefile with standard targets: build, test, lint, run, migrate, generate |
| DevOps | `ci` | CI/CD pipeline for GitHub Actions, GitLab CI, or Bitbucket Pipelines |
| DevOps | `k8s` | Kubernetes manifests: Deployment, Service, Ingress, HPA, ConfigMap |
| DevOps | `terraform` | Terraform infrastructure for AWS, GCP, or Azure |
| | | |
| Testing | `testkit` | Test fixtures, factories, and testcontainers for integration testing |
| Testing | `e2e` | End-to-end test scaffold with HTTP test client |
| Testing | `mocks` | Mock generation with mockery from interfaces |
| | | |
| Security | `auth` | JWT authentication middleware, token generation and validation |
| Security | `cors` | CORS middleware with configurable allowed origins and methods |
| Security | `ratelimit` | Token-bucket rate limiting middleware per-IP and global |
| Security | `encryption` | AES-GCM symmetric encryption and PBKDF2 key derivation |
| | | |
| API | `grpc` | gRPC server setup with buf configuration and interceptors |
| | | |
| Scaffold | `crud` | Full CRUD scaffolding for an entity — handler, service, repository, migration |

---

## Domain Modules

The `domain add` command scaffolds a new bounded-context module inside an existing project. It respects the architecture declared in `archforge.yaml`.

**`add` vs `domain add`:**

- `arch_forge add <module>` installs infrastructure capabilities — databases, queues, auth, observability.
- `arch_forge domain add <name>` generates business domain code — entities, use cases, ports, and adapters for a named bounded context.

```bash
arch_forge domain add payment
arch_forge domain add order --dry-run
arch_forge domain add notification --project-dir ./myapp
```

### Generated structure

For a `hexagonal/modular` project, `arch_forge domain add payment` produces:

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
    │   └── service.go           # Application service implementing inbound port
    └── adapters/
        ├── inbound/
        │   └── http/
        │       └── handler.go
        └── outbound/
            └── postgres/
                └── repository.go
```

### Architecture-specific output

| Architecture | Generated structure under `internal/payment/` |
|---|---|
| `hexagonal` (classic) | `domain/`, `ports/inbound/`, `ports/outbound/`, `application/`, `adapters/inbound/`, `adapters/outbound/` |
| `hexagonal` (modular) | Same structure, but scoped inside `internal/payment/` |
| `clean` | `domain/`, `usecase/`, `ports/`, `adapters/` |
| `ddd` | `domain/` (includes repository interface), `application/`, `infrastructure/http/`, `infrastructure/persistence/postgres/` |
| `modular_monolith` | `domain/`, `service/`, `handler/` |
| `cqrs` | `domain/`, `command/`, `query/`, `infrastructure/` |
| `microservice` | `domain/`, `port/`, `app/`, `adapter/` |

### Flags

| Flag | Description |
|---|---|
| `--project-dir` | Project directory (default: current directory) |
| `--dry-run` | Preview files that would be created without writing to disk |

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
x Architecture compliance check failed
```

| Flag | Description |
|---|---|
| `--fix` | Print suggested fixes for each violation |
| `--threshold` | Minimum passing score for CI pipelines (default: 7.0) |

```bash
arch_forge doctor --threshold 9.0
arch_forge doctor --fix
```

---

## Dependency Graph

The `graph` command generates a visual dependency diagram of your project's internal packages.

```bash
# Mermaid diagram to stdout (embeds directly in GitHub Markdown)
arch_forge graph

# DOT format to file
arch_forge graph --format dot --output graph.dot

# Include external dependencies
arch_forge graph --include-external
```

---

## Custom Modules

The module system is extensible. You can create local modules that follow the same structure as built-in modules and share them across projects.

```bash
arch_forge module create mymodule --category infrastructure
arch_forge module dev mymodule      # live validation feedback
arch_forge module validate mymodule # validate before use
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

## Commands Reference

### Project

| Command | Description | Key Flags |
|---|---|---|
| `arch_forge init <name>` | Create a new project | `--arch`, `--variant`, `--modules`, `--preset`, `--module-path`, `--go-version`, `--dry-run` |
| `arch_forge inspect` | Show annotated project structure | — |
| `arch_forge doctor` | Validate architecture compliance | `--fix`, `--threshold` |
| `arch_forge graph` | Generate dependency diagram | `--format`, `--output`, `--include-external` |

### Modules

| Command | Description | Key Flags |
|---|---|---|
| `arch_forge add <module...>` | Add infrastructure modules to an existing project | `--project-dir`, `--dry-run` |
| `arch_forge list archs` | List supported architectures | — |
| `arch_forge list modules` | List available modules | — |
| `arch_forge list presets` | List available presets | — |

### Domains

| Command | Description | Key Flags |
|---|---|---|
| `arch_forge domain add <name>` | Add a bounded-context domain to a project | `--project-dir`, `--dry-run` |

### Custom Modules

| Command | Description | Key Flags |
|---|---|---|
| `arch_forge module create <name>` | Scaffold a new local module | `--category` |
| `arch_forge module validate <name>` | Validate a local module | — |
| `arch_forge module dev <name>` | Watch mode for module development | — |

### Utilities

| Command | Description | Key Flags |
|---|---|---|
| `arch_forge update` | Update arch_forge to the latest release | `--force` |
| `arch_forge completion <shell>` | Generate shell completion scripts | — |

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

## Project Config

After running `init`, arch_forge writes an `archforge.yaml` file to the project root. This file records the project's architecture decisions and is required by `doctor`, `add`, `inspect`, and `domain add`.

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

> Do not delete `archforge.yaml`. All post-init commands depend on it.

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

## Self-Update

```bash
arch_forge update           # check and install latest release
arch_forge update --force   # force reinstall even if already up to date
```

---

## Contributing

Contributions are welcome. To get started:

1. Fork and clone the repository.
2. Run `make test` to execute the full test suite.
3. Run `make lint` to check for linting issues.
4. Open a pull request with a clear description of the change.

Follow standard Go conventions (Effective Go, Go Proverbs). All new code requires appropriate tests. Wrap errors using `fmt.Errorf("context: %w", err)`.

---

## License

MIT — see [LICENSE](./LICENSE) for details.
