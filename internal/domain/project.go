package domain

// ProjectInfo holds the minimal identifying information for a project.
type ProjectInfo struct {
	Name       string
	ModulePath string
}

// Project represents a fully initialized arch_forge project configuration.
type Project struct {
	Name             string
	ModulePath       string
	Version          string
	Arch             Architecture
	Variant          Variant
	InstalledModules []string
	GoVersion        string
}

// HasModule reports whether moduleName is already installed in the project.
func (p Project) HasModule(moduleName string) bool {
	for _, m := range p.InstalledModules {
		if m == moduleName {
			return true
		}
	}
	return false
}

// AddModule returns a new Project value with moduleName appended to InstalledModules.
// The original Project is not modified.
func (p Project) AddModule(moduleName string) Project {
	modules := make([]string, len(p.InstalledModules), len(p.InstalledModules)+1)
	copy(modules, p.InstalledModules)
	modules = append(modules, moduleName)
	p.InstalledModules = modules
	return p
}
