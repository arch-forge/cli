package domain

// FixKind classifies the type of auto-fix that can be applied.
type FixKind string

const (
	// FixMoveFile suggests moving a file to the correct layer directory.
	FixMoveFile FixKind = "move_file"
	// FixRemoveImport suggests removing a forbidden import statement.
	FixRemoveImport FixKind = "remove_import"
	// FixManual indicates the violation requires manual resolution.
	FixManual FixKind = "manual"
)

// FixSuggestion describes how to resolve a Violation.
type FixSuggestion struct {
	Violation Violation
	Kind      FixKind
	// Description is a human-readable explanation of the fix.
	Description string
	// SourcePath and DestPath are used for FixMoveFile suggestions.
	SourcePath string
	DestPath   string
	// ImportPath is the import path to remove for FixRemoveImport suggestions.
	ImportPath string
	// AutoApplicable reports whether this fix can be applied without user review.
	AutoApplicable bool
}
