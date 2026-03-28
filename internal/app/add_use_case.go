package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/spf13/afero"
)

// AddOptions carries parameters for adding modules to an existing project.
type AddOptions struct {
	ProjectDir    string
	Modules       []string
	ModuleOptions map[string]map[string]any // module name → option key → value
	DryRun        bool
}

// AddUseCase adds one or more modules to an existing arch_forge project.
type AddUseCase struct {
	repo    port.TemplateRepository
	gen     port.Generator
	cfg     port.ConfigReader
	patcher port.Patcher
}

// NewAddUseCase constructs an AddUseCase with the given dependencies.
func NewAddUseCase(
	repo port.TemplateRepository,
	gen port.Generator,
	cfg port.ConfigReader,
	patcher port.Patcher,
) *AddUseCase {
	return &AddUseCase{
		repo:    repo,
		gen:     gen,
		cfg:     cfg,
		patcher: patcher,
	}
}

// Execute runs the full add workflow:
//  1. Read archforge.yaml from ProjectDir.
//  2. Validate requested modules are not already installed.
//  3. Validate module compatibility with project arch+variant.
//  4. Resolve missing transitive dependencies in topological order.
//  5. Generate new files for each module into the project filesystem.
//  6. Apply module patches (hooks) to existing files.
//  7. Update archforge.yaml with new module list (skipped on DryRun).
func (uc *AddUseCase) Execute(ctx context.Context, opts AddOptions) error {
	// Step 1 — Read config.
	cfgPath := configPath(opts.ProjectDir)
	projectCfg, err := uc.cfg.Read(cfgPath)
	if err != nil {
		return fmt.Errorf("read project config: %w", err)
	}

	// Step 2 — Check not already installed.
	for _, name := range opts.Modules {
		for _, installed := range projectCfg.InstalledModules {
			if name == installed {
				return fmt.Errorf("module %q: %w", name, domain.ErrModuleAlreadyInstalled)
			}
		}
	}

	// Step 3 & 4 — Resolve dependency order, skipping already-installed modules.
	toInstall, err := uc.resolveAddOrder(ctx, opts.Modules, projectCfg.Arch, projectCfg.Variant, projectCfg.InstalledModules)
	if err != nil {
		return fmt.Errorf("add: %w", err)
	}

	// Step 4 — Set up filesystem.
	var fs afero.Fs
	if opts.DryRun {
		fs = afero.NewMemMapFs()
	} else {
		fs = afero.NewBasePathFs(afero.NewOsFs(), opts.ProjectDir)
	}

	// Step 5 — Generate files for each new module.
	pathsUC := NewResolvePathsUseCase()
	allModules := append(projectCfg.InstalledModules, toInstall...)

	// tctx is declared here so it remains accessible in the patch loop below.
	var tctx domain.TemplateContext

	for _, moduleName := range toInstall {
		modulePaths, err := pathsUC.Execute(projectCfg.Arch, projectCfg.Variant, moduleName)
		if err != nil {
			return fmt.Errorf("resolve paths for %s: %w", moduleName, err)
		}

		moduleOpts := opts.ModuleOptions[moduleName]
		if moduleOpts == nil {
			moduleOpts = make(map[string]any)
		}

		tctx = domain.TemplateContext{
			Project: domain.ProjectInfo{
				Name:       projectCfg.Name,
				ModulePath: projectCfg.ModulePath,
			},
			Arch:       projectCfg.Arch,
			Variant:    projectCfg.Variant,
			ModuleName: moduleName,
			Options:    moduleOpts,
			Modules:    allModules,
			GoVersion:  projectCfg.GoVersion,
			Module:     projectCfg.ModulePath,
			Paths:      modulePaths,
		}

		templates, err := uc.repo.LoadModuleTemplates(moduleName, string(projectCfg.Arch), string(projectCfg.Variant))
		if err != nil {
			return fmt.Errorf("load templates for module %s: %w", moduleName, err)
		}
		if err := uc.gen.Generate(ctx, tctx, templates, fs); err != nil {
			return fmt.Errorf("generate module %s: %w", moduleName, err)
		}
	}

	// Step 6 — Apply patches for each new module.
	for _, moduleName := range toInstall {
		manifest, err := uc.repo.LoadModuleManifest(moduleName)
		if err != nil {
			return fmt.Errorf("load manifest for %s: %w", moduleName, err)
		}

		patchRequests, err := uc.buildPatchRequests(ctx, manifest, tctx)
		if err != nil {
			return fmt.Errorf("build patch requests for %s: %w", moduleName, err)
		}

		if len(patchRequests) > 0 {
			if err := uc.patcher.Apply(ctx, opts.ProjectDir, patchRequests, fs); err != nil {
				return fmt.Errorf("apply patches for %s: %w", moduleName, err)
			}
		}
	}

	// Step 7 — Update archforge.yaml (skip on DryRun).
	if !opts.DryRun {
		projectCfg.InstalledModules = append(projectCfg.InstalledModules, toInstall...)
		if err := uc.cfg.Write(cfgPath, projectCfg); err != nil {
			return fmt.Errorf("update config: %w", err)
		}
	}

	return nil
}

// configPath returns the path to archforge.yaml in the given directory.
func configPath(projectDir string) string {
	return filepath.Join(projectDir, "archforge.yaml")
}

// resolveAddOrder returns modules to install in topological order,
// excluding any already in alreadyInstalled.
func (uc *AddUseCase) resolveAddOrder(
	ctx context.Context,
	modules []string,
	arch domain.Architecture,
	variant domain.Variant,
	alreadyInstalled []string,
) ([]string, error) {
	// Build a fast-lookup set for already-installed modules.
	installed := make(map[string]bool, len(alreadyInstalled))
	for _, name := range alreadyInstalled {
		installed[name] = true
	}

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
		// If already installed in the project, skip without error.
		if installed[name] {
			return nil
		}
		if visited[name] {
			return nil
		}
		if inStack[name] {
			// Cycle detected — treat as already resolved to avoid infinite loop.
			return nil
		}

		mod, err := uc.repo.LoadModuleManifest(name)
		if err != nil {
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

// buildPatchRequests renders each module patch's hook template and returns
// a []port.PatchRequest ready for the Patcher.
// For M1 simplicity, Content is left empty; full hook template rendering
// is deferred to M1.7+.
func (uc *AddUseCase) buildPatchRequests(
	ctx context.Context,
	manifest domain.Module,
	tctx domain.TemplateContext,
) ([]port.PatchRequest, error) {
	if len(manifest.Patches) == 0 {
		return nil, nil
	}
	requests := make([]port.PatchRequest, 0, len(manifest.Patches))
	for _, p := range manifest.Patches {
		requests = append(requests, port.PatchRequest{
			TargetGlob: p.TargetGlob,
			Action:     p.Action,
			Anchor:     p.Anchor,
			Content:    "", // hook template rendering done in M1.7+
			Optional:   p.Optional,
		})
	}
	return requests, nil
}
