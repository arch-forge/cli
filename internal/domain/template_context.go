package domain

// TemplateContext holds all values made available to template rendering.
type TemplateContext struct {
	Project    ProjectInfo
	Arch       Architecture
	Variant    Variant
	ModuleName string
	Options    map[string]any
	Modules    []string
	GoVersion  string
	Entity     *EntityInfo
	Module     string // Go module path (= Project.ModulePath)
	Paths      ResolvedPaths
	DomainName string // bounded-context domain name (set by domain add command)
}
