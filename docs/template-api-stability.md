# Template API Stability Guarantee

This document defines which parts of the arch_forge template API are stable across releases and which are subject to change. It is intended for developers who write custom templates, extend arch_forge modules, or build tooling on top of arch_forge's generated output.

---

## Stability Policy

The template API is **stable from v1.0.0**. Once a field, function, anchor, or structural contract is documented in this file and shipped in a v1.x release, it will not be removed or changed in an incompatible way within the v1 major version.

Breaking changes require a **major version bump** (v1 → v2). Breaking changes are:
- Removing a `TemplateContext` field.
- Renaming a `TemplateContext` field.
- Changing the type or return type of a template function.
- Removing a template function.
- Changing the anchor comment format.
- Removing or renaming `ResolvedPaths` fields.

Non-breaking changes that do not require a major version bump:
- Adding new `TemplateContext` fields (existing templates receive the zero value).
- Adding new template functions.
- Adding new anchor types.
- Changing internal adapter implementations.
- Changing port interfaces that users only call, not implement.

---

## What is Stable

The following items are stable from v1.0.0 and are covered by this guarantee:

- All fields of `TemplateContext` listed in the [TemplateContext Reference](#templatecontext-reference) below.
- All fields of `domain.ResolvedPaths` listed in the [ResolvedPaths Reference](#resolvedpaths-reference) below.
- All template function signatures listed in the [Template Function Reference](#template-function-reference) below.
- The anchor comment format `// arch_forge:<anchor>` as documented in the [Anchor Comment Reference](#anchor-comment-reference) below.
- The `.tmpl` file suffix convention: all template source files end in `.tmpl` and are rendered to the same path with the suffix stripped.

---

## What is NOT Stable

The following items may change between minor versions without notice:

- The internal implementation of `internal/adapter/generator/Engine` (the rendering pipeline, error formatting, file writing logic).
- Port interfaces in `internal/port/` — these are consumed by the internal application layer and are not part of the public extension API.
- The layout of `internal/adapter/`, `internal/app/`, and `internal/domain/` packages within arch_forge itself.
- The exact content of generated boilerplate inside `.tmpl` files (non-structural changes such as comment wording or example values).
- The `archforge.yaml` schema beyond the fields documented in the project manifest specification.
- Module `module.yaml` fields beyond `name`, `version`, `description`, `category`, `architectures`, `variants`, `dependencies`, `go_dependencies`, `options`, and `patches`.

---

## Versioning

arch_forge uses [Semantic Versioning](https://semver.org):

- **Patch** (`v1.0.x`): Bug fixes, security patches, documentation corrections. No API changes.
- **Minor** (`v1.x.0`): New architectures, new modules, new template functions, new `TemplateContext` fields. Fully backward-compatible.
- **Major** (`vX.0.0`): Breaking changes to the template API as described above.

All template API changes are recorded in `CHANGELOG.md` under the relevant version heading with a `Template API` subsection.

---

## TemplateContext Reference

`TemplateContext` is the root data value passed to every `.tmpl` file. Access fields using `{{ .FieldName }}` in templates.

| Field | Type | Description |
|---|---|---|
| `Project` | `ProjectInfo` | Identifying information for the project being generated. |
| `Project.Name` | `string` | The project name as provided to `arch_forge init`. Example: `myapp`. |
| `Project.ModulePath` | `string` | The Go module path. Example: `github.com/acme/myapp`. |
| `Arch` | `Architecture` | The architecture being generated. One of: `hexagonal`, `clean`, `ddd`, `standard`, `modular_monolith`, `cqrs`, `microservice`. |
| `Variant` | `Variant` | The structural variant. Either `classic` or `modular`. |
| `ModuleName` | `string` | When adding a module with `arch_forge add`, the module slug (e.g., `logging`). Empty during `arch_forge init`. |
| `Options` | `map[string]any` | Key-value map of options declared in `module.yaml` and provided by the user at generation time. |
| `Modules` | `[]string` | Slice of module slugs already installed in the project. Use with the `hasModule` template function. |
| `GoVersion` | `string` | The Go toolchain version string recorded in the project. Example: `1.23`. |
| `Entity` | `*EntityInfo` | When generating entity-centric scaffolding (e.g., `crud` module), holds entity metadata. `nil` when not applicable. |
| `Entity.Name` | `string` | Entity name in PascalCase. Example: `Product`. |
| `Entity.Fields` | `[]Field` | Slice of field descriptors for the entity. |
| `Entity.Relations` | `[]Relation` | Slice of relation descriptors for the entity. |
| `Module` | `string` | Convenience alias for `Project.ModulePath`. Identical value; provided for shorter template syntax: `{{ .Module }}`. |
| `Paths` | `ResolvedPaths` | Pre-computed canonical directory paths for this architecture/variant combination. |

### EntityInfo.Field Reference

| Sub-field | Type | Description |
|---|---|---|
| `Name` | `string` | Field name in PascalCase. Example: `CreatedAt`. |
| `Type` | `string` | Declared type identifier. Example: `string`, `int`, `time.Time`. |
| `GoType` | `string` | Full Go type for use in generated code. |
| `SQLType` | `string` | SQL column type. Example: `TEXT`, `TIMESTAMPTZ`. |
| `JSONName` | `string` | JSON serialization key in snake_case. Example: `created_at`. |
| `DBName` | `string` | Database column name. Example: `created_at`. |
| `Nullable` | `bool` | Whether the field is nullable in the database. |
| `Validation` | `string` | Optional validation tag or rule string. |

### EntityInfo.Relation Reference

| Sub-field | Type | Description |
|---|---|---|
| `Kind` | `RelationKind` | One of `belongs_to`, `has_many`, `has_one`. |
| `Target` | `string` | Name of the related entity in PascalCase. |

---

## ResolvedPaths Reference

`ResolvedPaths` is available as `{{ .Paths }}` and provides the canonical directory path for each architectural layer. Paths are relative to the project root.

| Field | Type | Description |
|---|---|---|
| `Domain` | `string` | Directory containing domain entities and core business rules. |
| `Port` | `string` | Directory containing port interfaces (input/output or repository). |
| `App` | `string` | Directory containing application services or use cases. |
| `Adapter` | `string` | Root directory for all adapter implementations. |
| `Handler` | `string` | Directory for inbound HTTP handler adapters. |
| `Repository` | `string` | Directory for outbound database adapter implementations. |
| `Migration` | `string` | Directory for SQL migration files. Always `migrations` for classic variant; `internal/<module>/migrations` for modular variant. |
| `Test` | `string` | Canonical directory for test files co-located with domain code. Equals `Domain` by convention. |

Resolved values by architecture and variant:

| Architecture | Variant | `Domain` | `Port` | `App` | `Handler` | `Repository` |
|---|---|---|---|---|---|---|
| `hexagonal` | `classic` | `internal/domain` | `internal/port` | `internal/app` | `internal/adapter/inbound/http` | `internal/adapter/outbound/postgres` |
| `clean` | `classic` | `internal/entity` | `internal/usecase/port` | `internal/usecase` | `internal/controller/http` | `internal/gateway/postgres` |
| `ddd` | `classic` | `internal/domain/model` | `internal/domain/repository` | `internal/application` | `internal/infrastructure/http` | `internal/infrastructure/persistence/postgres` |
| `standard` | `classic` | `internal/model` | `internal/service` | `internal/service` | `internal/handler` | `internal/repository` |
| `cqrs` | `classic` | `internal/domain` | `internal/event` | `internal/command` | `internal/infrastructure/http` | `internal/infrastructure/postgres` |
| `modular_monolith` | `classic` | `internal/module` | `internal/module` | `internal/module` | `internal/platform` | `internal/platform` |
| `microservice` | `classic` | `internal/domain` | `internal/port` | `internal/app` | `internal/adapter/http` | `internal/adapter/postgres` |
| any | `modular` | `internal/<module>/domain` | `internal/<module>/port` | `internal/<module>/app` | `internal/<module>/adapter` | `internal/<module>/adapter` |

---

## Template Function Reference

All functions below are registered in `template.FuncMap` by `internal/adapter/generator/buildFuncMap()` and are available in every `.tmpl` file.

### String Case Conversion

| Function | Signature | Description | Example |
|---|---|---|---|
| `camelCase` | `camelCase(s string) string` | Converts `snake_case`, `kebab-case`, or `PascalCase` to `camelCase`. | `camelCase "user_name"` → `"userName"` |
| `pascalCase` | `pascalCase(s string) string` | Converts any casing to `PascalCase`. | `pascalCase "user_name"` → `"UserName"` |
| `snakeCase` | `snakeCase(s string) string` | Converts `PascalCase` or `camelCase` to `snake_case`. | `snakeCase "UserName"` → `"user_name"` |
| `kebabCase` | `kebabCase(s string) string` | Converts `PascalCase` or `camelCase` to `kebab-case`. | `kebabCase "UserName"` → `"user-name"` |
| `upperCase` | `upperCase(s string) string` | Converts string to `UPPER CASE`. Alias for `strings.ToUpper`. | `upperCase "hello"` → `"HELLO"` |
| `lowerCase` | `lowerCase(s string) string` | Converts string to `lower case`. Alias for `strings.ToLower`. | `lowerCase "HELLO"` → `"hello"` |

### Path Helpers

These functions accept a `domain.ResolvedPaths` value (i.e., `{{ .Paths }}`) and return the string path for that layer.

| Function | Signature | Description |
|---|---|---|
| `domainPath` | `domainPath(p ResolvedPaths) string` | Returns `p.Domain`. |
| `portPath` | `portPath(p ResolvedPaths) string` | Returns `p.Port`. |
| `appPath` | `appPath(p ResolvedPaths) string` | Returns `p.App`. |
| `adapterPath` | `adapterPath(p ResolvedPaths) string` | Returns `p.Adapter`. |
| `handlerPath` | `handlerPath(p ResolvedPaths) string` | Returns `p.Handler`. |
| `repoPath` | `repoPath(p ResolvedPaths) string` | Returns `p.Repository`. |
| `joinPath` | `joinPath(elems ...string) string` | Joins path segments using `path.Join`. |

### Module and Package Helpers

| Function | Signature | Description | Example |
|---|---|---|---|
| `hasModule` | `hasModule(modules []string, name string) bool` | Reports whether `name` is present in the `modules` slice. Use with `{{ .Modules }}`. | `{{ if hasModule .Modules "logging" }}` |
| `goPackage` | `goPackage(p string) string` | Returns the last path component of `p` as a valid Go identifier (hyphens replaced with underscores). | `goPackage "internal/adapter/http"` → `"http"` |

### Utility

| Function | Signature | Description | Example |
|---|---|---|---|
| `currentYear` | `currentYear() int` | Returns the current calendar year as an integer. | `{{ currentYear }}` → `2026` |
| `quote` | `quote(s string) string` | Wraps `s` in double-quote characters. Useful for generating Go string literals. | `{{ quote .Module }}` → `"github.com/acme/myapp"` |
| `trimSuffix` | `trimSuffix(s, suffix string) string` | Removes `suffix` from the end of `s`. Alias for `strings.TrimSuffix`. | `trimSuffix "main.go.tmpl" ".tmpl"` → `"main.go"` |
| `replace` | `replace(s, old, new string) string` | Replaces all occurrences of `old` with `new` in `s`. Alias for `strings.ReplaceAll`. | `replace .Module "/" "_"` |

---

## Anchor Comment Reference

Anchor comments allow modules to inject code into files generated by `arch_forge init` without requiring the user to manually edit those files. The `arch_forge add` command applies patches declared in `module.yaml` by locating anchor comments in the target file and inserting rendered hook templates immediately after them.

### Syntax

```go
// arch_forge:<anchor-name>
```

The comment must appear on its own line. The colon (`:`) and anchor name are required. Leading whitespace is preserved.

### Registered Anchors

| Anchor | Location | Purpose |
|---|---|---|
| `arch_forge:providers` | `cmd/*/main.go` | Injection point for infrastructure initialization (database connections, logger setup, etc.) |
| `arch_forge:routes` | `cmd/*/main.go` | Injection point for HTTP route registration (mounting routers, registering handlers) |

### Example — anchor in generated `main.go`

```go
func main() {
    ctx := context.Background()

    // arch_forge:providers
    // arch_forge:routes

    http.ListenAndServe(":8080", mux)
}
```

When `arch_forge add database` is run, it inserts the database initialization block after the `arch_forge:providers` line:

```go
    // arch_forge:providers
    // db, err := database.Connect(ctx, database.Config{DSN: os.Getenv("DATABASE_URL")})
    // if err != nil { log.Fatalf("connect to database: %v", err) }
    // defer db.Close()
```

### Rules for Custom Module Anchors

- A module may declare a `patches` entry in `module.yaml` referencing any anchor name. If the anchor is not found in the target file, the patch is skipped when the entry has `optional: true`; otherwise it is an error.
- Anchor names must be lowercase, using only letters, digits, and hyphens (e.g., `arch_forge:event-bus`).
- Multiple patches targeting the same anchor are applied in the order they are declared.
- Anchors are never removed from files after injection; they remain as permanent extension points.

---

## Upgrade Path

When upgrading arch_forge across minor versions:

1. Run `arch_forge update` to install the new binary.
2. Check `CHANGELOG.md` for the `Template API` section of each version between your current and target version.
3. If you maintain custom modules with `.tmpl` files, review the new `TemplateContext` fields — new fields added in a minor version are zero-valued in templates that do not reference them, so existing templates continue to compile and render correctly.
4. Re-run `arch_forge doctor` on existing projects to verify no new validation rules flag issues.

When upgrading across major versions (v1 → v2):

1. Read the migration guide published with the major release.
2. Update all references to renamed or removed `TemplateContext` fields in your custom `.tmpl` files.
3. Update all references to renamed or removed template functions.
4. Re-generate golden test fixtures if you use snapshot testing against arch_forge's output.
5. Submit issues against the arch_forge repository if a breaking change is undocumented.
