package analyzer_test

import (
	"testing"

	"github.com/arch-forge/cli/internal/adapter/analyzer"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSuggestFixes_EmptyReport verifies that an empty report produces an empty slice.
func TestSuggestFixes_EmptyReport(t *testing.T) {
	report := domain.Report{}
	suggestions := analyzer.SuggestFixes(report)
	assert.Empty(t, suggestions)
}

// TestSuggestFixes_DomainViolation verifies that a domain-no-port violation produces a FixMoveFile suggestion.
func TestSuggestFixes_DomainViolation(t *testing.T) {
	violation := domain.Violation{
		File:     "internal/domain/order.go",
		Line:     5,
		Rule:     "domain-no-port",
		Message:  `package in "internal/domain" must not import "internal/port"`,
		Severity: domain.SeverityError,
	}
	report := domain.Report{Violations: []domain.Violation{violation}}

	suggestions := analyzer.SuggestFixes(report)
	require.Len(t, suggestions, 1)

	s := suggestions[0]
	assert.Equal(t, domain.FixMoveFile, s.Kind)
	assert.Equal(t, violation, s.Violation)
	assert.Contains(t, s.Description, "domain")
	assert.Contains(t, s.Description, "port")
	assert.Equal(t, "internal/domain/order.go", s.SourcePath)
	assert.Equal(t, "internal/app/order.go", s.DestPath)
	assert.False(t, s.AutoApplicable)
}

// TestSuggestFixes_LateralAdapter verifies that an adapter-no-lateral violation produces a FixManual suggestion.
func TestSuggestFixes_LateralAdapter(t *testing.T) {
	violation := domain.Violation{
		File:     "internal/adapter/cli/root.go",
		Line:     5,
		Rule:     "adapter-no-lateral",
		Message:  `adapter subtree "cli" must not import adapter subtree "generator"`,
		Severity: domain.SeverityWarning,
	}
	report := domain.Report{Violations: []domain.Violation{violation}}

	suggestions := analyzer.SuggestFixes(report)
	require.Len(t, suggestions, 1)

	s := suggestions[0]
	assert.Equal(t, domain.FixManual, s.Kind)
	assert.Equal(t, violation, s.Violation)
	assert.Contains(t, s.Description, "Lateral adapter dependency detected")
	assert.False(t, s.AutoApplicable)
}

// TestSuggestFixes_PortViolation verifies that a port-no-app violation produces a FixRemoveImport suggestion.
func TestSuggestFixes_PortViolation(t *testing.T) {
	violation := domain.Violation{
		File:     "internal/port/user_service.go",
		Line:     8,
		Rule:     "port-no-app",
		Message:  `package in "internal/port" must not import "internal/app"`,
		Severity: domain.SeverityError,
	}
	report := domain.Report{Violations: []domain.Violation{violation}}

	suggestions := analyzer.SuggestFixes(report)
	require.Len(t, suggestions, 1)

	s := suggestions[0]
	assert.Equal(t, domain.FixRemoveImport, s.Kind)
	assert.Equal(t, violation, s.Violation)
	assert.Contains(t, s.Description, "port layer must not import app")
	assert.Equal(t, "internal/app", s.ImportPath)
	assert.False(t, s.AutoApplicable)
}

// TestSuggestFixes_DefaultRule verifies that an unrecognized rule produces a FixManual suggestion.
func TestSuggestFixes_DefaultRule(t *testing.T) {
	violation := domain.Violation{
		File:     "internal/app/some_file.go",
		Line:     3,
		Rule:     "app-no-adapter",
		Message:  `package in "internal/app" must not import "internal/adapter"`,
		Severity: domain.SeverityError,
	}
	report := domain.Report{Violations: []domain.Violation{violation}}

	suggestions := analyzer.SuggestFixes(report)
	require.Len(t, suggestions, 1)

	s := suggestions[0]
	assert.Equal(t, domain.FixManual, s.Kind)
	assert.Contains(t, s.Description, "Manual fix required")
	assert.False(t, s.AutoApplicable)
}

// TestExtractTargetLayer verifies extractTargetLayer via SuggestFixes outputs.
func TestExtractTargetLayer(t *testing.T) {
	tests := []struct {
		rule            string
		targetLayer     string
		descContains    string
	}{
		{"domain-no-port", "port", "port"},
		{"domain-no-app", "app", "app"},
		{"domain-no-adapter", "adapter", "adapter"},
		{"port-no-app", "app", "app"},
		{"port-no-adapter", "adapter", "adapter"},
		// adapter-no-lateral maps to the FixManual case; description mentions "Lateral".
		{"adapter-no-lateral", "lateral", "Lateral adapter dependency"},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			violation := domain.Violation{
				File:    "internal/domain/file.go",
				Rule:    tt.rule,
				Message: `package in "internal/domain" must not import "internal/` + tt.targetLayer + `"`,
			}
			report := domain.Report{Violations: []domain.Violation{violation}}
			suggestions := analyzer.SuggestFixes(report)
			require.Len(t, suggestions, 1)
			assert.Contains(t, suggestions[0].Description, tt.descContains)
		})
	}
}
