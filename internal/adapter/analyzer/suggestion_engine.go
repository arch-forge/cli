package analyzer

import (
	"fmt"
	"strings"

	"github.com/arch-forge/cli/internal/domain"
)

// SuggestFixes generates FixSuggestions for each violation in the report.
// It returns one suggestion per violation.
func SuggestFixes(report domain.Report) []domain.FixSuggestion {
	suggestions := make([]domain.FixSuggestion, 0, len(report.Violations))

	for _, v := range report.Violations {
		suggestions = append(suggestions, suggestFix(v))
	}

	return suggestions
}

// suggestFix maps a single Violation to the most appropriate FixSuggestion.
func suggestFix(v domain.Violation) domain.FixSuggestion {
	switch {
	case strings.Contains(v.Rule, "no-lateral"):
		return domain.FixSuggestion{
			Violation:      v,
			Kind:           domain.FixManual,
			Description:    "Lateral adapter dependency detected. Refactor to use a shared port interface instead of importing between adapter packages.",
			AutoApplicable: false,
		}

	case strings.Contains(v.Rule, "domain-no-"):
		targetLayer := extractTargetLayer(v.Rule)
		dest := strings.Replace(v.File, "domain", "app", 1)
		return domain.FixSuggestion{
			Violation:      v,
			Kind:           domain.FixMoveFile,
			Description:    fmt.Sprintf("Move %q out of domain layer — domain must not depend on %s", v.File, targetLayer),
			SourcePath:     v.File,
			DestPath:       dest,
			AutoApplicable: false,
		}

	case v.Rule == "port-no-app" || v.Rule == "port-no-adapter":
		targetLayer := extractTargetLayer(v.Rule)
		importPath := extractImportFromMessage(v.Message)
		return domain.FixSuggestion{
			Violation:      v,
			Kind:           domain.FixRemoveImport,
			Description:    fmt.Sprintf("Remove forbidden import in %s — port layer must not import %s", v.File, targetLayer),
			ImportPath:     importPath,
			AutoApplicable: false,
		}

	default:
		return domain.FixSuggestion{
			Violation:      v,
			Kind:           domain.FixManual,
			Description:    fmt.Sprintf("Manual fix required: %s", v.Message),
			AutoApplicable: false,
		}
	}
}

// extractTargetLayer extracts the target layer name from a rule name.
// Rule format: "{from_layer}-no-{to_layer}" or "adapter-no-lateral".
// Example: "domain-no-port" → "port".
func extractTargetLayer(rule string) string {
	parts := strings.SplitN(rule, "-no-", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return rule
}

// extractImportFromMessage attempts to extract the imported package path from a violation message.
// Violation messages follow the format: `package in "from" must not import "to"`.
func extractImportFromMessage(message string) string {
	// Message format: `package in "internal/port" must not import "internal/app"`
	const marker = `must not import "`
	idx := strings.Index(message, marker)
	if idx == -1 {
		return ""
	}
	rest := message[idx+len(marker):]
	end := strings.Index(rest, `"`)
	if end == -1 {
		return rest
	}
	return rest[:end]
}
