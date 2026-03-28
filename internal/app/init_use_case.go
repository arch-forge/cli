package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/spf13/afero"
)

// InitOptions carries all parameters for initialising a new project.
type InitOptions struct {
	Name       string
	ModulePath string              // Go module path, e.g. "github.com/acme/myapp"
	Arch       domain.Architecture
	Variant    domain.Variant
	Modules    []string            // module names to install
	DryRun     bool
	OutputDir  string              // destination root directory
	GoVersion  string              // e.g. "1.23"; defaults to "1.23" if empty
	Preset     string              // optional named preset (e.g. "starter", "production-api")
	Fs         afero.Fs            // if set, used as the output filesystem (caller-provided, e.g. for dry-run inspection)
}

// InitUseCase orchestrates new project creation.
type InitUseCase struct {
	repo port.TemplateRepository
	gen  port.Generator
	cfg  port.ConfigReader
}

// NewInitUseCase constructs an InitUseCase.
func NewInitUseCase(
	repo port.TemplateRepository,
	gen port.Generator,
	cfg port.ConfigReader,
) *InitUseCase {
	return &InitUseCase{
		repo: repo,
		gen:  gen,
		cfg:  cfg,
	}
}

// Execute runs the full init workflow:
//  1. Apply preset defaults (if --preset is set).
//  2. Validate options (arch valid, variant valid, modules compatible).
//  3. Resolve module dependency graph (topological order).
//  4. Load and render architecture base templates.
//  5. For each module in dependency order, load and render module templates.
//  6. Write archforge.yaml (skipped on DryRun).
func (uc *InitUseCase) Execute(ctx context.Context, opts InitOptions) error {
	// Step 0 — Apply preset defaults, if a preset name was supplied.
	if opts.Preset != "" {
		preset, found := domain.FindPreset(opts.Preset)
		if !found {
			return fmt.Errorf("init: unknown preset %q", opts.Preset)
		}
		// Only fill in values that were not explicitly provided by the caller.
		if opts.Arch == "" {
			opts.Arch = preset.Arch
		}
		if opts.Variant == "" {
			opts.Variant = preset.Variant
		}
		if len(opts.Modules) == 0 {
			opts.Modules = preset.Modules
		}
	}

	// Step 1 — Validate.
	if !opts.Arch.IsValid() {
		return fmt.Errorf("init: %w", domain.ErrArchNotSupported)
	}
	if !opts.Variant.IsValid() {
		return fmt.Errorf("init: %w", domain.ErrVariantNotSupported)
	}
	if opts.GoVersion == "" {
		opts.GoVersion = "1.23"
	}

	// Step 2 — Resolve module dependency order.
	orderedModules, err := uc.resolveDependencyOrder(ctx, opts.Modules, opts.Arch, opts.Variant)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}

	// Step 3 — Build afero.Fs for output.
	var fs afero.Fs
	switch {
	case opts.Fs != nil:
		fs = opts.Fs
	case opts.DryRun:
		fs = afero.NewMemMapFs()
	default:
		fs = afero.NewBasePathFs(afero.NewOsFs(), opts.OutputDir)
	}

	// Step 4 — Resolve paths.
	pathsUC := NewResolvePathsUseCase()
	paths, err := pathsUC.Execute(opts.Arch, opts.Variant, opts.Name)
	if err != nil {
		return fmt.Errorf("init: resolve paths: %w", err)
	}

	// Step 5 — Build base TemplateContext.
	baseTctx := uc.buildBaseContext(opts, paths)

	// Step 6 — Generate arch base templates.
	archTemplates, err := uc.repo.LoadArchTemplates(string(opts.Arch), string(opts.Variant))
	if err != nil {
		return fmt.Errorf("load arch templates: %w", err)
	}
	if err := uc.gen.Generate(ctx, baseTctx, archTemplates, fs); err != nil {
		return fmt.Errorf("generate arch templates: %w", err)
	}

	// Step 7 — Generate each module's templates.
	for _, moduleName := range orderedModules {
		modulePaths, _ := pathsUC.Execute(opts.Arch, opts.Variant, moduleName)
		moduleTctx := baseTctx
		moduleTctx.ModuleName = moduleName
		moduleTctx.Paths = modulePaths

		moduleTemplates, err := uc.repo.LoadModuleTemplates(moduleName, string(opts.Arch), string(opts.Variant))
		if err != nil {
			return fmt.Errorf("load templates for module %s: %w", moduleName, err)
		}
		if err := uc.gen.Generate(ctx, moduleTctx, moduleTemplates, fs); err != nil {
			return fmt.Errorf("generate module %s: %w", moduleName, err)
		}
	}

	// Step 8 — Write archforge.yaml (skip on DryRun).
	if !opts.DryRun {
		cfgPath := filepath.Join(opts.OutputDir, "archforge.yaml")
		projectCfg := &port.ProjectConfig{
			Name:             opts.Name,
			ModulePath:       opts.ModulePath,
			GoVersion:        opts.GoVersion,
			Arch:             opts.Arch,
			Variant:          opts.Variant,
			InstalledModules: orderedModules,
		}
		if err := uc.cfg.Write(cfgPath, projectCfg); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
	}

	return nil
}

// buildBaseContext constructs the base TemplateContext from the init options and resolved paths.
func (uc *InitUseCase) buildBaseContext(opts InitOptions, paths domain.ResolvedPaths) domain.TemplateContext {
	return domain.TemplateContext{
		Project: domain.ProjectInfo{
			Name:       opts.Name,
			ModulePath: opts.ModulePath,
		},
		Arch:      opts.Arch,
		Variant:   opts.Variant,
		GoVersion: opts.GoVersion,
		Module:    opts.ModulePath,
		Paths:     paths,
		Options:   make(map[string]any),
		Modules:   opts.Modules,
	}
}

// resolveDependencyOrder returns module names in topological order
// (dependencies before dependents). It validates that each module
// exists in the repository and is compatible with the arch+variant.
func (uc *InitUseCase) resolveDependencyOrder(
	ctx context.Context,
	modules []string,
	arch domain.Architecture,
	variant domain.Variant,
) ([]string, error) {
	// requested tracks names the user explicitly requested.
	requested := make(map[string]bool, len(modules))
	for _, name := range modules {
		requested[name] = true
	}

	// ordered collects the final topological result.
	var ordered []string
	// visited tracks modules already appended to ordered.
	visited := make(map[string]bool)
	// inStack detects cycles.
	inStack := make(map[string]bool)

	var visit func(name string, explicitlyRequested bool) error
	visit = func(name string, explicitlyRequested bool) error {
		if visited[name] {
			return nil
		}
		if inStack[name] {
			// Cycle detected — treat as already resolved to avoid infinite loop.
			return nil
		}

		mod, err := uc.repo.LoadModuleManifest(name)
		if err != nil {
			// If it wraps ErrModuleNotFound or is that error itself.
			if isNotFound(err) {
				if explicitlyRequested {
					return fmt.Errorf("module %q: %w", name, domain.ErrModuleNotFound)
				}
				// Optional dependency not found — skip silently.
				return nil
			}
			return fmt.Errorf("load manifest for module %q: %w", name, err)
		}

		// Validate compatibility.
		if !mod.SupportsArch(arch) {
			return fmt.Errorf("module %q: %w", name, domain.ErrIncompatibleModule)
		}
		if !mod.SupportsVariant(variant) {
			return fmt.Errorf("module %q: %w", name, domain.ErrIncompatibleModule)
		}

		inStack[name] = true

		// Recurse into dependencies before appending this module.
		for _, dep := range mod.Dependencies {
			if err := visit(dep.Name, dep.Required); err != nil {
				return err
			}
		}

		inStack[name] = false
		if !visited[name] {
			visited[name] = true
			ordered = append(ordered, name)
		}
		return nil
	}

	for _, name := range modules {
		if err := visit(name, true); err != nil {
			return nil, err
		}
	}

	return ordered, nil
}

// isNotFound reports whether err wraps or equals domain.ErrModuleNotFound.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	unwrapped := err
	for unwrapped != nil {
		if unwrapped == domain.ErrModuleNotFound {
			return true
		}
		next, ok := unwrapped.(interface{ Unwrap() error })
		if !ok {
			break
		}
		unwrapped = next.Unwrap()
	}
	return false
}
