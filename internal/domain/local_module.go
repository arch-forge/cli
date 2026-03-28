// Package domain contains the core business entities for arch_forge.
package domain

// LocalModuleStatus represents the validation state of a local module.
type LocalModuleStatus string

const (
	// LocalModuleStatusValid means the module passed all checks with no errors or warnings.
	LocalModuleStatusValid LocalModuleStatus = "valid"
	// LocalModuleStatusInvalid means the module has one or more error-level issues.
	LocalModuleStatusInvalid LocalModuleStatus = "invalid"
	// LocalModuleStatusWarning means the module has no errors but has one or more warnings.
	LocalModuleStatusWarning LocalModuleStatus = "warning"
)

// LocalModuleIssue describes a single validation problem in a local module.
type LocalModuleIssue struct {
	Kind    string // "error" or "warning"
	Message string
	Field   string // optional: the yaml field or file path that caused the issue
}

// LocalModuleValidation is the result of validating a local module directory.
type LocalModuleValidation struct {
	ModuleName string
	ModuleDir  string
	Status     LocalModuleStatus
	Issues     []LocalModuleIssue
}

// IsValid reports whether there are no error-level issues.
func (v LocalModuleValidation) IsValid() bool {
	for _, issue := range v.Issues {
		if issue.Kind == "error" {
			return false
		}
	}
	return true
}
