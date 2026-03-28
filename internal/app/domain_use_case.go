package app

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/spf13/afero"
)

// DomainAddOptions carries parameters for adding a bounded-context domain to an existing project.
type DomainAddOptions struct {
	ProjectDir string
	Name       string
	DryRun     bool
}

// DomainAddUseCase scaffolds a new bounded-context domain following the project's architecture.
type DomainAddUseCase struct {
	repo port.TemplateRepository
	gen  port.Generator
	cfg  port.ConfigReader
}

// NewDomainAddUseCase constructs a DomainAddUseCase with the given dependencies.
func NewDomainAddUseCase(repo port.TemplateRepository, gen port.Generator, cfg port.ConfigReader) *DomainAddUseCase {
	return &DomainAddUseCase{repo: repo, gen: gen, cfg: cfg}
}

// Execute runs the full domain-add workflow:
//  1. Read archforge.yaml from ProjectDir.
//  2. Validate opts.Name is not empty.
//  3. For non-dry-run modular variants: check the domain directory does not already exist.
//  4. Load domain templates for the project's arch+variant.
//  5. Compute module root and rewrite template rel-paths.
//  6. Build TemplateContext and call the generator.
func (uc *DomainAddUseCase) Execute(ctx context.Context, opts DomainAddOptions) error {
	// Step 1 — Read config.
	cfgPath := filepath.Join(opts.ProjectDir, "archforge.yaml")
	projectCfg, err := uc.cfg.Read(cfgPath)
	if err != nil {
		return fmt.Errorf("read project config: %w", err)
	}

	// Step 2 — Validate name.
	if strings.TrimSpace(opts.Name) == "" {
		return fmt.Errorf("domain add: name must not be empty")
	}

	domainSnake := toSnakeCaseDomain(opts.Name)

	// Step 3 — Check domain does not already exist (real run only, modular variants).
	if !opts.DryRun && projectCfg.Variant == domain.VariantModular {
		expectedDir := filepath.Join(opts.ProjectDir, "internal", domainSnake)
		exists, checkErr := afero.DirExists(afero.NewOsFs(), expectedDir)
		if checkErr != nil {
			return fmt.Errorf("check domain directory: %w", checkErr)
		}
		if exists {
			return fmt.Errorf("domain %q: %w", opts.Name, domain.ErrDomainAlreadyExists)
		}
	}

	// Step 4 — Load domain templates.
	templates, err := uc.repo.LoadDomainTemplates(string(projectCfg.Arch), string(projectCfg.Variant))
	if err != nil {
		return fmt.Errorf("load domain templates: %w", err)
	}

	// Step 5 — Compute module root and rewrite rel-paths.
	moduleRoot := domainModuleRoot(projectCfg.Arch, projectCfg.Variant, domainSnake)
	for i := range templates {
		rel := strings.ReplaceAll(templates[i].RelPath, "__domain__", domainSnake)
		templates[i].RelPath = moduleRoot + rel
	}

	// Step 6 — Resolve paths and build template context.
	resolvedPaths, err := domain.ResolvePaths(projectCfg.Arch, projectCfg.Variant, domainSnake)
	if err != nil {
		return fmt.Errorf("resolve paths: %w", err)
	}

	tctx := domain.TemplateContext{
		DomainName: opts.Name,
		Module:     projectCfg.ModulePath,
		Arch:       projectCfg.Arch,
		Variant:    projectCfg.Variant,
		Project: domain.ProjectInfo{
			Name:       projectCfg.Name,
			ModulePath: projectCfg.ModulePath,
		},
		GoVersion: projectCfg.GoVersion,
		Paths:     resolvedPaths,
	}

	// Step 7 — Choose filesystem.
	var fs afero.Fs
	if opts.DryRun {
		fs = afero.NewMemMapFs()
	} else {
		fs = afero.NewBasePathFs(afero.NewOsFs(), opts.ProjectDir)
	}

	// Step 8 — Generate files.
	if err := uc.gen.Generate(ctx, tctx, templates, fs); err != nil {
		return fmt.Errorf("generate domain %q: %w", opts.Name, err)
	}

	return nil
}

// domainModuleRoot returns the prefix to prepend to each template RelPath.
// For modular variants (except DDD and ModularMonolith which encode domain name
// via __domain__ placeholder in their template paths), the prefix is
// "internal/{domainName}/". Classic variants rely entirely on __domain__
// placeholders in filenames.
func domainModuleRoot(arch domain.Architecture, variant domain.Variant, domainName string) string {
	if variant == domain.VariantModular {
		switch arch {
		case domain.ArchDDD, domain.ArchModularMonolith:
			// These archs encode the domain name in the template path via __domain__.
			return ""
		default:
			return path.Join("internal", domainName) + "/"
		}
	}
	// Classic variants: __domain__ placeholder in filenames handles placement.
	return ""
}

// toSnakeCaseDomain converts a domain name (e.g. "myDomain", "MyDomain", "my-domain")
// to snake_case (e.g. "my_domain"). This mirrors the snakeCase function in the
// generator package without creating a cross-package dependency on unexported symbols.
func toSnakeCaseDomain(s string) string {
	var words []string
	var current strings.Builder

	runes := []rune(s)
	for i, r := range runes {
		if r == '_' || r == '-' {
			if current.Len() > 0 {
				words = append(words, strings.ToLower(current.String()))
				current.Reset()
			}
			continue
		}

		// Detect lowercase→uppercase transition (e.g. "myField": 'y'→'F').
		if i > 0 && unicode.IsUpper(r) && unicode.IsLower(runes[i-1]) {
			if current.Len() > 0 {
				words = append(words, strings.ToLower(current.String()))
				current.Reset()
			}
		}

		// Detect start of an acronym → word boundary (e.g. "XMLParser": 'L'→'P').
		if i > 0 && unicode.IsUpper(r) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
			if current.Len() > 1 {
				words = append(words, strings.ToLower(current.String()))
				current.Reset()
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, strings.ToLower(current.String()))
	}

	if len(words) == 0 {
		return strings.ToLower(s)
	}
	return strings.Join(words, "_")
}
