package domain

// ModuleCategory classifies a module by its functional area.
type ModuleCategory string

const (
	CategoryCore            ModuleCategory = "core"
	CategoryInfrastructure  ModuleCategory = "infrastructure"
	CategoryObservability   ModuleCategory = "observability"
	CategoryDevOps          ModuleCategory = "devops"
	CategoryTesting         ModuleCategory = "testing"
	CategorySecurity        ModuleCategory = "security"
)

// IsValid reports whether c is a recognized module category.
func (c ModuleCategory) IsValid() bool {
	switch c {
	case CategoryCore, CategoryInfrastructure, CategoryObservability,
		CategoryDevOps, CategoryTesting, CategorySecurity:
		return true
	}
	return false
}
