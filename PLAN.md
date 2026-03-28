# PLAN.md — arch_forge Implementation Plan

## Overview

This document defines the phased implementation plan for **arch_forge**, a CLI tool that generates Go project structures based on proven architectural patterns. The plan is derived from the [PRD.md](./PRD.md) and organized into milestones aligned with the versioned roadmap.

---

## Architecture Decision

arch_forge itself follows **Hexagonal Architecture (classic variant)**. Dependency rule: `domain ← port ← app ← adapter`. Adapters depend on ports, never on each other.

```
cmd/archforge/main.go
internal/
  domain/      ← entities, value objects, rules
  port/        ← interfaces (generator, analyzer, template_repo)
  app/         ← use cases (init, add, doctor, inspect)
  adapter/
    cli/       ← Cobra commands + bubbletea TUI
    generator/ ← template engine
    analyzer/  ← AST-based architecture validator
    repository/← local template storage
templates/go/  ← embedded templates per arch/variant/module
```

---

## Milestone 0 — Project Scaffold ✅

**Goal**: Compilable skeleton with no business logic.

### Tasks

- [x] Initialize Go module: `github.com/archforge/cli`
- [x] Create directory structure per CLAUDE.md spec
- [x] Set up `cmd/archforge/main.go` entry point
- [x] Configure Cobra root command with version flag
- [x] Set up `Makefile` with targets: `build`, `test`, `lint`, `run`
- [x] Configure `.golangci.yml` for linting
- [x] Set up `go.sum` with initial dependencies (Cobra, Viper, bubbletea)
- [x] Configure `.goreleaser.yaml` for distribution
- [x] Add `archforge.yaml` schema definition

**Output**: `arch_forge --version` runs and prints version. ✅

---

## Milestone 1 — v0.1 MVP ✅

**Goal**: `arch_forge init` generates Standard Layout, Hexagonal, and Clean Architecture projects with core modules.

### 1.1 Domain Layer ✅

- [x] Define `Architecture` value object (enum: hexagonal, clean, standard, ddd, modular_monolith, cqrs, microservice)
- [x] Define `Variant` value object (enum: classic, modular)
- [x] Define `Module` entity with name, category, dependencies
- [x] Define `Project` aggregate (name, Go module path, architecture, variant, installed modules)
- [x] Define `TemplateContext` struct — full context passed to all templates
- [x] Define `Field` and `EntityInfo` structs — used by CRUD module
- [x] Define `ResolvedPaths` — pre-calculated paths for each arch+variant combo
- [x] Define domain errors (ErrArchNotSupported, ErrModuleNotFound, ErrMissingDependency, etc.)

### 1.2 Port Layer ✅

- [x] `port.Generator` interface — `Generate(ctx TemplateContext, fs afero.Fs) error`
- [x] `port.TemplateRepository` interface — `Load(arch, variant, module string) ([]TemplateFile, error)`
- [x] `port.ConfigReader` interface — `Read(path string) (*ProjectConfig, error)`, `Write(path string, cfg *ProjectConfig) error`
- [x] `port.Patcher` interface — `Apply(fs afero.Fs, patches []Patch) error`

### 1.3 App Layer — Use Cases ✅

- [x] `app.InitUseCase` — orchestrates project generation
  - Read options (arch, variant, modules)
  - Resolve module dependency graph
  - Build `TemplateContext`
  - Call `Generator`
  - Write `archforge.yaml`
- [x] `app.AddUseCase` — adds a module to an existing project
  - Read `archforge.yaml`
  - Validate module compatibility with project arch
  - Resolve missing dependencies
  - Generate + patch
- [x] `app.ListUseCase` — returns available architectures or modules
- [x] `app.ResolvePathsUseCase` — resolves `ResolvedPaths` for a given arch+variant

### 1.4 Adapter — Template Engine ✅

- [x] `adapter/generator` — implements `port.Generator`
  - Register `templateFuncs` (camelCase, pascalCase, snakeCase, goType, sqlType, domainPath, portPath, etc.) — 19 functions
  - Walk `base/` + `variants/{arch}/{variant}/` template trees
  - Render each `.tmpl` file with `text/template`
  - Write output to `afero.Fs` at resolved destination path
- [x] `adapter/repository` — implements `port.TemplateRepository`
  - Embed `templates/go/` via `go:embed`
  - Serve embedded files by arch/variant/module key
- [x] `adapter/patcher` — implements `port.Patcher`
  - Parse anchor comments `// arch_forge:<name>`
  - Inject rendered hook content after anchor (inject_after, inject_before, replace)
  - Rewrite file with `go/format` (best-effort)
- [x] `adapter/config` — implements `port.ConfigReader`
  - YAML read/write using `gopkg.in/yaml.v3`
  - Error mapping to domain sentinel errors

### 1.5 Adapter — CLI ✅

- [x] Root command setup with Cobra (stateless, use-case injection)
- [x] `init` command
  - Flags: `--arch`, `--variant`, `--modules`, `--module-path`, `--go-version`, `--dry-run`
  - Delegates to `app.InitUseCase`
- [x] `add` command
  - Positional args: module names
  - Flags: `--project-dir`, `--dry-run`
  - Delegates to `app.AddUseCase`
- [x] `list` command
  - Subcommands: `archs`, `modules`
  - Output: tabwriter-formatted table
- [ ] Interactive wizard (bubbletea) — deferred to M2.4

### 1.6 Templates — v0.1 Architectures ✅

For each architecture, `templates/go/{arch}/{variant}/` created with 43 total template files:

| Architecture | classic | modular |
|---|---|---|
| `standard` | ✅ | ✅ |
| `hexagonal` | ✅ | ✅ |
| `clean` | ✅ | ✅ |

Each template tree includes:
- [x] `cmd/*/main.go.tmpl` — with arch_forge anchor comments
- [x] Core package stubs per arch layer
- [x] `go.mod.tmpl`
- [x] `archforge.yaml.tmpl`
- [x] `.gitignore.tmpl`

### 1.7 Templates — v0.1 Modules ✅

`templates/go/modules/{module}/` created for all 5 core modules:

| Module | Description | Status |
|---|---|---|
| `api` | chi router, middleware chain, error handling | ✅ |
| `database` | postgres connection, pool, health check | ✅ |
| `logging` | slog structured logging, correlation IDs | ✅ |
| `docker` | multi-stage Dockerfile, docker-compose | ✅ |
| `makefile` | standard Makefile targets | ✅ |

Each module has:
- [x] `module.yaml` — manifest
- [x] `prompts.yaml` — interactive questions
- [x] `base/` — arch-agnostic files
- [x] `variants/` — arch+variant-specific file directories (ready for overrides)
- [x] `hooks/` — patch templates for main.go injection

### 1.8 Testing — v0.1 ✅

- [x] Unit tests for domain logic (path resolution, dependency graph, value objects) — 14 tests
- [x] Unit tests for generator funcs (camelCase, pascalCase, snakeCase, kebabCase) — white-box
- [x] Integration tests for Engine using `afero.MemMapFs` — template rendering, context cancellation
- [x] Integration tests for FilePatcher — inject_after, inject_before, replace, optional/required anchors
- [x] Integration tests for ViperConfig — round-trip read/write, error mapping
- [x] Integration tests for InitUseCase — dry-run, real file generation, invalid inputs
- [x] **48 tests total, all passing** (`go test ./... -count=1`)
- [ ] Snapshot tests (go-golden) — deferred to post-M1
- [ ] Full add cycle integration test — deferred to post-M1

---

## Milestone 2 — v0.2 Developer Experience ✅

**Goal**: `doctor`, `inspect`, polished TUI, more modules, presets.

### 2.1 Doctor Command ✅

- [x] `adapter/analyzer` — implements AST-based architecture validator
  - Walk project Go files with `go/ast`
  - Check import directions per architecture rules
  - Detect business logic in wrong layers
  - Report violations with file:line references
- [x] `app.DoctorUseCase` — orchestrates analysis, returns `Report`
- [x] `doctor` CLI command — renders report as formatted output
- [x] Score calculation (n_valid / total_rules * 10)

### 2.2 Inspect Command ✅

- [x] `app.InspectUseCase` — reads `archforge.yaml` + file tree, returns `ProjectSummary`
- [x] `inspect` CLI command — renders tree with layers highlighted
- [x] `adapter/scanner` — `OsScanner` for real filesystem traversal
- [x] `adapter/cli/tree_renderer.go` — box-drawing tree + JSON output

### 2.3 Graph Command ✅

- [x] Dependency graph builder using `go/ast` import analysis (`adapter/graph`)
- [x] Mermaid diagram generator (`RenderMermaid`)
- [x] DOT diagram generator (`RenderDOT`)
- [x] `graph` CLI command with `--format=mermaid|dot`, `--include-external`, `--output`

### 2.4 TUI Enhancement ✅

- [x] Replace simple prompts with full bubbletea model (`internal/adapter/cli/tui/`)
- [x] Architecture selection with description panel (bubbles/list)
- [x] Variant selection screen
- [x] Module multi-select with custom checkbox UI
- [x] Confirmation screen
- [x] Wizard invoked when `arch_forge init` called without `--arch` or `--preset`

### 2.5 Additional Modules — v0.2 ✅

| Module | Notes |
|---|---|
| `auth` | JWT token generation, middleware, login handler |
| `crud` | Entity scaffold, domain type, migration template |
| `grpc` | buf.yaml, proto file, server + interceptors |
| `cache` | Redis client with cache-aside pattern |

### 2.6 Presets — v0.2 ✅

- [x] Define preset registry in `internal/domain/preset.go`
- [x] `starter` — Standard Layout + api, logging, docker, makefile
- [x] `production-api` — Hexagonal + api, database, auth, logging, docker, makefile
- [x] `microservice` — Microservice + grpc, database, logging, docker, makefile
- [x] `--preset` flag on `init` command
- [x] `list presets` subcommand

### 2.7 Shell Completions ✅

- [x] Cobra completion for bash, zsh, fish, powershell (`completion` command)
- [x] Dynamic completion for `--arch` flag (from `domain.AllArchitectures()`)
- [x] Dynamic completion for `--preset` flag (from `domain.AllPresets()`)
- [x] Static completion for `--modules` flag and `add` positional args

---

## Milestone 3 — v0.3 Ecosystem ✅

**Goal**: Remaining architectures, more modules, doctor --fix, CI/K8s.

### 3.1 Additional Architectures ✅

| Architecture | classic | modular |
|---|---|---|
| `ddd` | ✅ | ✅ |
| `cqrs` | ✅ | ✅ |
| `modular_monolith` | ✅ | ✅ |
| `microservice` | ✅ | ✅ |

### 3.2 Additional Modules — v0.3 ✅

| Module | Category |
|---|---|
| `queue` | Infrastructure (NATS/RabbitMQ + in-memory backend) |
| `storage` | Infrastructure (local FS + Store interface) |
| `search` | Infrastructure (Meilisearch/Elasticsearch + in-memory) |
| `metrics` | Observability (Prometheus counter/histogram + middleware) |
| `tracing` | Observability (OpenTelemetry OTLP) |
| `healthcheck` | Observability (/health, /ready, /live) |
| `ci` | DevOps (GitHub Actions + GitLab CI) |
| `k8s` | DevOps (Deployment, Service, HPA) |
| `terraform` | DevOps (AWS ECS task definition) |
| `testkit` | Testing (fixtures, random helpers, test DB) |
| `e2e` | Testing (typed HTTP test client, TestMain scaffold) |
| `mocks` | Testing (mockery config + Makefile target) |
| `cors` | Security (CORS middleware) |
| `ratelimit` | Security (token-bucket per-IP limiter) |
| `encryption` | Security (AES-256-GCM + PBKDF2) |

### 3.3 Doctor --fix ✅

- [x] Suggestion engine (`adapter/analyzer/suggestion_engine.go`) — maps violations to fix suggestions
- [x] `FixSuggestion` domain type with `FixKind` (move_file, remove_import, manual)
- [x] `--fix` flag on doctor command — displays suggestions after report
- [x] `--yes` flag — auto-applies `AutoApplicable` fixes (all manual for v0.3)

### 3.4 Custom Module System ✅

- [x] `module create <name>` — scaffold new module directory with manifest, prompts, base/, hooks/
- [x] `module dev <name>` — watch mode: re-validates on interval (default 2s), SIGINT for clean stop
- [x] `module validate <name>` — validates manifest fields + parses all .tmpl files

---

## Milestone 4 — v1.0 Production Ready ✅

- [x] Complete documentation with examples per architecture
- [x] Template API stability guarantee
- [x] Homebrew tap (`brew install archforge/tap/arch-forge`)
- [x] Scoop manifest for Windows
- [x] Docker image (`docker run archforge/cli init myapp`)
- [x] GoReleaser full pipeline (multi-arch binaries, changelogs, GitHub releases)
- [x] `arch_forge update` self-update command

---

## Implementation Order (Critical Path)

```
✅ Milestone 0 (scaffold)
    ↓
✅ Domain layer (M1.1)
    ↓
✅ Port layer (M1.2)
    ↓
✅ App — InitUseCase + ResolvePathsUseCase (M1.3)
    ↓
✅ Template engine + repository (M1.4)          ← parallelizable with M1.3
    ↓
✅ Templates: standard layout both variants (M1.6)
    ↓
✅ Templates: core modules (api, logging, docker, makefile) (M1.7)
    ↓
✅ CLI init/add/list commands (M1.5)
    ↓
✅ App — AddUseCase (M1.3)
    ↓
✅ Templates: hexagonal + clean (M1.6)
    ↓
✅ Patcher system / hooks (M1.4)
    ↓
✅ Tests: unit + integration — 48 passing (M1.8)
    ↓
[v0.1 RELEASE] ← next
    ↓
✅ Doctor + Inspect (M2.1, M2.2)
    ↓
✅ Auth + CRUD modules (M2.5)
    ↓
✅ TUI polish + presets (M2.4, M2.6)
    ↓
✅ Graph + Shell Completions (M2.3, M2.7)
    ↓
[v0.2 RELEASE] ← next
```

---

## Key Technical Decisions

| Decision | Choice | Reason |
|---|---|---|
| Filesystem abstraction | `afero` | Enables in-memory FS for tests without touching disk |
| Template engine | `text/template` + `go:embed` | Zero runtime dependencies, templates inside binary |
| Code generation | `jennifer` | Programmatic Go AST when templates are insufficient |
| Snapshot testing | `go-golden` | Deterministic output validation for generated files |
| CLI framework | `cobra` + `viper` | Industry standard for Go CLIs, flag/config integration |
| TUI | `bubbletea` | Elm-architecture TUI, composable components |
| AST analysis | `go/ast` + `go/format` | For doctor command and patch import injection |

---

## Testing Strategy Summary

| Test Type | Tool | Scope |
|---|---|---|
| Unit | testify | Domain logic, path resolution, dep graph |
| Snapshot | go-golden | Template rendering per arch+variant+module |
| Integration | afero + testify | Full init/add cycles in-memory |
| CLI E2E | Cobra test utils | Command invocation → file tree assertions |

All tests run with `make test`. CI runs on every PR via GitHub Actions.

---

## File Naming Conventions

- All Go source files: `snake_case.go`
- Template files: `{target_filename}.tmpl` (e.g., `order_service.go.tmpl`)
- Test files: `{source_file}_test.go`
- Golden files: `testdata/golden/{arch}/{variant}/{module}/{filename}`
- Hook templates: `templates/go/modules/{module}/hooks/{target}.go.tmpl`
