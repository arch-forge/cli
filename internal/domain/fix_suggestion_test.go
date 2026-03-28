package domain_test

import (
	"testing"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
)

// TestFixKind_Constants verifies the string values of FixKind constants.
func TestFixKind_Constants(t *testing.T) {
	assert.Equal(t, domain.FixKind("move_file"), domain.FixMoveFile)
	assert.Equal(t, domain.FixKind("remove_import"), domain.FixRemoveImport)
	assert.Equal(t, domain.FixKind("manual"), domain.FixManual)
}

// TestFixSuggestion_AutoApplicable verifies the default AutoApplicable value is false.
func TestFixSuggestion_AutoApplicable(t *testing.T) {
	s := domain.FixSuggestion{}
	assert.False(t, s.AutoApplicable, "AutoApplicable should default to false")
}
