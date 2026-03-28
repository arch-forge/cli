# CLAUDE.md — arch_forge

## Project Overview

arch_forge is a CLI tool that generates Go project structures based on proven architectural patterns (Hexagonal, Clean Architecture, DDD, Standard Layout, Modular Monolith, CQRS+ES, Microservice). Each architecture supports two variants: `classic` (by-the-book) and `modular` (organized by business domain).


## Mandatory Rules

### 1. All Code Must Be Written in English

- All source code, variable names, function names, struct names, constants, errors, and comments MUST be in English.
- Commit messages MUST be in English.
- Only documentation files intended for Spanish-speaking audiences (like PRD.md) may be in Spanish.
- No exceptions. If translating an existing concept from the PRD, find the correct English term.

### 2. Use Specialized Agents for Every Task

Every implementation task MUST be executed by spawning the appropriate specialized agent. Do NOT write code inline without delegating to agents. The workflow is:

- **Planning**: Use the `Plan` agent to design the approach before writing any code.
- **Exploration**: Use the `Explore` agent to search and understand codebase context before making changes.
- **Implementation**: Use the `general-purpose` agent for writing code, running builds, and executing multi-step tasks.
- **Testing**: After writing code, spawn an agent to run tests and validate.
- **Parallel execution**: When tasks are independent, launch multiple agents concurrently.

Example workflow:
```
1. Explore agent → understand existing code
2. Plan agent → design the implementation
3. general-purpose agent(s) → implement (parallelize when possible)
4. general-purpose agent → run tests
```

### 3. Always Use context7 for Up-to-Date Documentation

Before writing any code that uses an external library, you MUST query context7 to get the latest documentation for that library. This applies to:

- **Cobra** — resolve library ID, then query for command setup, flags, completions.
- **Viper** — resolve library ID, then query for config loading, binding, file watching.
- **bubbletea** — resolve library ID, then query for model/update/view pattern, components.
- **testify** — resolve library ID, then query for assertions, suites, mocks.
- **afero** — resolve library ID, then query for filesystem abstraction, MemMapFs.
- **jennifer** — resolve library ID, then query for code generation API.
- **go-golden** — resolve library ID, then query for snapshot testing patterns.
- **Any other library** being used for the first time in this project.

The process is always:
```
1. mcp__context7__resolve-library-id  → get the library ID
2. mcp__context7__query-docs           → fetch relevant docs for your specific use case
3. Write code based on the ACTUAL current API, not from memory
```

Never assume you know the current API of a library. APIs change between versions. Always verify with context7 first.

---

## Tech Stack

- **Language**: Go 1.23+
- **CLI**: Cobra + Viper
- **TUI**: bubbletea (interactive wizard)
- **Templates**: text/template + go:embed
- **Code analysis**: go/ast (for `doctor` command)
- **Code generation**: jennifer (programmatic Go code gen)
- **Testing**: testify, afero (in-memory FS), go-golden (snapshots)
- **Distribution**: GoReleaser

## Project Structure

```
arch_forge/
├── cmd/archforge/main.go          # Entry point
├── internal/
│   ├── domain/                     # Entities, rules, value objects
│   ├── port/                       # Interfaces (generator, analyzer, template_repo)
│   ├── app/                        # Use cases (init, add, doctor, inspect)
│   └── adapter/
│       ├── cli/                    # Cobra commands + tui/
│       ├── generator/              # Template engine
│       ├── analyzer/               # AST analyzer
│       └── repository/             # Local template storage
├── templates/go/                   # Embedded templates per arch/variant/module
├── archforge.yaml                  # Project config
├── Makefile
└── .goreleaser.yaml
```

## Architecture

The project itself follows **Hexagonal Architecture (classic variant)**. Dependency rule: domain ← port ← app ← adapter. Adapters depend on ports, never on each other.

---

## Commands Reference

Key CLI commands being built:

| Command | Purpose |
|---|---|
| `arch_forge init` | Create new project (interactive wizard or flags) |
| `arch_forge add <module>` | Add module to existing project |
| `arch_forge list archs\|modules` | List available architectures/modules |
| `arch_forge doctor` | Validate architecture compliance |
| `arch_forge inspect` | Show project structure overview |
| `arch_forge module create\|dev\|validate` | Create and test custom local modules |
| `arch_forge domain add <name>` | Add a new bounded-context domain module to an existing project |
| `arch_forge update` | Update arch_forge to latest version |

## Module System

Modules live in `templates/go/modules/<name>/` with:
- `module.yaml` — manifest (dependencies, options, supported archs)
- `prompts.yaml` — interactive wizard questions
- `base/` — files common to all architectures
- `variants/{arch}/{variant}/` — architecture-specific templates
- `hooks/` — patches to inject into existing files via anchor comments (`// arch_forge:<anchor>`)

## Testing Strategy

- Unit tests: domain logic, template rendering, path resolution
- Snapshot tests (go-golden): validate generated file output against golden files
- Integration tests: full `init` + `add` cycle using afero in-memory filesystem
- All tests run with `make test`

## Code Style

- Follow standard Go conventions (Effective Go, Go Proverbs)
- Use `golangci-lint` with the project's `.golangci.yml` config
- Error handling: wrap errors with `fmt.Errorf("context: %w", err)`
- No global state — use dependency injection through constructor functions
- Interfaces in `port/` package, implementations in `adapter/` subpackages
