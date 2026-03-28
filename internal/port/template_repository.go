package port

import "github.com/arch-forge/cli/internal/domain"

// TemplateRepository loads template files for a given arch/variant/module
// combination from its backing store.
type TemplateRepository interface {
	// LoadArchTemplates returns all templates for the architecture base.
	// Template root: templates/go/{arch}/{variant}/
	LoadArchTemplates(arch, variant string) ([]TemplateFile, error)

	// LoadModuleTemplates returns template files for a named module filtered
	// for the given arch+variant. Combines base/ and variants/{arch}/{variant}/.
	LoadModuleTemplates(moduleName, arch, variant string) ([]TemplateFile, error)

	// LoadModuleManifest returns the parsed Module manifest for moduleName.
	LoadModuleManifest(moduleName string) (domain.Module, error)

	// ListModules returns the names of all available modules.
	ListModules() ([]string, error)

	// LoadDomainTemplates returns template files for generating a new bounded-context domain.
	// Template root: templates/go/domains/{arch}/{variant}/
	LoadDomainTemplates(arch, variant string) ([]TemplateFile, error)
}
