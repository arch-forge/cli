package repository

import (
	"embed"
	"fmt"
	"io/fs"
	"path"

	archforge "github.com/arch-forge/cli"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"gopkg.in/yaml.v3"
)

// templateFS is the embedded filesystem sourced from the root archforge package.
var templateFS embed.FS = archforge.TemplateFS

// EmbeddedRepository implements port.TemplateRepository using the
// binary-embedded templates/go/ directory tree.
type EmbeddedRepository struct{}

// NewEmbeddedRepository constructs a new EmbeddedRepository.
func NewEmbeddedRepository() *EmbeddedRepository {
	return &EmbeddedRepository{}
}

// LoadArchTemplates returns all templates for the given architecture and variant.
// Template root: templates/go/{arch}/{variant}/
// Returns nil, nil if the directory does not exist.
func (r *EmbeddedRepository) LoadArchTemplates(arch, variant string) ([]port.TemplateFile, error) {
	root := path.Join("templates/go", arch, variant)
	return walkEmbedDir(templateFS, root)
}

// LoadModuleTemplates returns template files for a named module filtered for
// the given arch+variant. Combines base/ and variants/{arch}/{variant}/.
// Returns nil, nil if directories do not exist.
func (r *EmbeddedRepository) LoadModuleTemplates(moduleName, arch, variant string) ([]port.TemplateFile, error) {
	baseRoot := path.Join("templates/go/modules", moduleName, "base")
	base, err := walkEmbedDir(templateFS, baseRoot)
	if err != nil {
		return nil, fmt.Errorf("load module base templates: %w", err)
	}

	variantRoot := path.Join("templates/go/modules", moduleName, "variants", arch, variant)
	additional, err := walkEmbedDir(templateFS, variantRoot)
	if err != nil {
		return nil, fmt.Errorf("load module variant templates: %w", err)
	}

	return mergeTemplates(base, additional), nil
}

// LoadModuleManifest returns the parsed Module manifest for moduleName.
// Returns domain.ErrModuleNotFound if the manifest file does not exist.
func (r *EmbeddedRepository) LoadModuleManifest(moduleName string) (domain.Module, error) {
	manifestPath := path.Join("templates/go/modules", moduleName, "module.yaml")
	data, err := templateFS.ReadFile(manifestPath)
	if err != nil {
		return domain.Module{}, domain.ErrModuleNotFound
	}

	var manifest moduleManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return domain.Module{}, fmt.Errorf("parse module manifest %q: %w", moduleName, err)
	}

	return manifestToModule(manifest), nil
}

// ListModules returns the names of all available modules.
// Returns an empty slice (not an error) if the modules directory does not exist.
func (r *EmbeddedRepository) ListModules() ([]string, error) {
	modulesDir := "templates/go/modules"
	entries, err := templateFS.ReadDir(modulesDir)
	if err != nil {
		// Directory does not exist yet — not an error.
		return []string{}, nil
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// LoadDomainTemplates returns template files for generating a new bounded-context domain.
// Template root: templates/go/domains/{arch}/{variant}/
// Returns nil, nil if the directory does not exist.
func (r *EmbeddedRepository) LoadDomainTemplates(arch, variant string) ([]port.TemplateFile, error) {
	root := path.Join("templates/go/domains", arch, variant)
	return walkEmbedDir(templateFS, root)
}

// walkEmbedDir walks an embed.FS subtree rooted at embedRoot,
// returning TemplateFiles with RelPath relative to embedRoot.
// Returns nil, nil if the root directory does not exist.
func walkEmbedDir(fsys embed.FS, embedRoot string) ([]port.TemplateFile, error) {
	var files []port.TemplateFile

	err := fs.WalkDir(fsys, embedRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			// If the root itself does not exist, skip gracefully.
			if p == embedRoot {
				return fs.SkipAll
			}
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, readErr := fsys.ReadFile(p)
		if readErr != nil {
			return fmt.Errorf("read embedded file %q: %w", p, readErr)
		}

		relPath, _ := cutPrefix(p, embedRoot+"/")

		files = append(files, port.TemplateFile{
			RelPath: relPath,
			Content: data,
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk embed dir %q: %w", embedRoot, err)
	}

	return files, nil
}

// cutPrefix removes prefix from s if present, returning the remainder and true.
// If prefix is not present, s and false are returned.
func cutPrefix(s, prefix string) (string, bool) {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):], true
	}
	return s, false
}

// mergeTemplates merges additional into base, with additional overriding
// on matching RelPath.
func mergeTemplates(base, additional []port.TemplateFile) []port.TemplateFile {
	if len(additional) == 0 {
		return base
	}

	// Build an index of additional files by RelPath.
	overrides := make(map[string]port.TemplateFile, len(additional))
	for _, f := range additional {
		overrides[f.RelPath] = f
	}

	// Start with base files, replacing any that are overridden.
	result := make([]port.TemplateFile, 0, len(base)+len(additional))
	for _, f := range base {
		if override, ok := overrides[f.RelPath]; ok {
			result = append(result, override)
			delete(overrides, f.RelPath)
		} else {
			result = append(result, f)
		}
	}

	// Append any additional files not present in base.
	for _, f := range additional {
		if _, stillPresent := overrides[f.RelPath]; stillPresent {
			result = append(result, f)
		}
	}

	return result
}

// moduleManifest is the unexported struct used to parse a module.yaml file.
type moduleManifest struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Category    string   `yaml:"category"`
	Architectures []string `yaml:"architectures"`
	Variants    []string `yaml:"variants"`
	Dependencies []struct {
		Name     string `yaml:"name"`
		Required bool   `yaml:"required"`
	} `yaml:"dependencies"`
	GoDependencies []struct {
		Package string `yaml:"package"`
		Version string `yaml:"version"`
	} `yaml:"go_dependencies"`
	Options []struct {
		Name        string   `yaml:"name"`
		Type        string   `yaml:"type"`
		Values      []string `yaml:"values"`
		Default     any      `yaml:"default"`
		Description string   `yaml:"description"`
		Condition   string   `yaml:"condition"`
	} `yaml:"options"`
	Patches []struct {
		TargetGlob   string `yaml:"target"`
		Action       string `yaml:"action"`
		Anchor       string `yaml:"anchor"`
		TemplatePath string `yaml:"template"`
		Optional     bool   `yaml:"optional"`
	} `yaml:"patches"`
}

// manifestToModule converts a moduleManifest to a domain.Module.
func manifestToModule(m moduleManifest) domain.Module {
	archs := make([]domain.Architecture, 0, len(m.Architectures))
	for _, a := range m.Architectures {
		archs = append(archs, domain.Architecture(a))
	}

	variants := make([]domain.Variant, 0, len(m.Variants))
	for _, v := range m.Variants {
		variants = append(variants, domain.Variant(v))
	}

	deps := make([]domain.ModuleDep, 0, len(m.Dependencies))
	for _, d := range m.Dependencies {
		deps = append(deps, domain.ModuleDep{
			Name:     d.Name,
			Required: d.Required,
		})
	}

	goDeps := make([]domain.GoDep, 0, len(m.GoDependencies))
	for _, g := range m.GoDependencies {
		goDeps = append(goDeps, domain.GoDep{
			Package: g.Package,
			Version: g.Version,
		})
	}

	options := make([]domain.ModuleOption, 0, len(m.Options))
	for _, o := range m.Options {
		options = append(options, domain.ModuleOption{
			Name:        o.Name,
			Type:        o.Type,
			Values:      o.Values,
			Default:     o.Default,
			Description: o.Description,
			Condition:   o.Condition,
		})
	}

	patches := make([]domain.Patch, 0, len(m.Patches))
	for _, p := range m.Patches {
		patches = append(patches, domain.Patch{
			TargetGlob:   p.TargetGlob,
			Action:       p.Action,
			Anchor:       p.Anchor,
			TemplatePath: p.TemplatePath,
			Optional:     p.Optional,
		})
	}

	return domain.Module{
		Name:                   m.Name,
		Version:                m.Version,
		Description:            m.Description,
		Category:               domain.ModuleCategory(m.Category),
		SupportedArchitectures: archs,
		SupportedVariants:      variants,
		Dependencies:           deps,
		GoDeps:                 goDeps,
		Options:                options,
		Patches:                patches,
	}
}
