package domain_test

import (
	"testing"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestLocalModuleValidation_IsValid_NoErrors(t *testing.T) {
	v := domain.LocalModuleValidation{
		ModuleName: "mymodule",
		ModuleDir:  "modules/mymodule",
		Status:     domain.LocalModuleStatusValid,
		Issues:     []domain.LocalModuleIssue{},
	}
	assert.True(t, v.IsValid())
}

func TestLocalModuleValidation_IsValid_WithError(t *testing.T) {
	v := domain.LocalModuleValidation{
		ModuleName: "mymodule",
		ModuleDir:  "modules/mymodule",
		Status:     domain.LocalModuleStatusInvalid,
		Issues: []domain.LocalModuleIssue{
			{Kind: "error", Message: "module.yaml not found", Field: "module.yaml"},
		},
	}
	assert.False(t, v.IsValid())
}

func TestLocalModuleValidation_IsValid_WarningOnly(t *testing.T) {
	v := domain.LocalModuleValidation{
		ModuleName: "mymodule",
		ModuleDir:  "modules/mymodule",
		Status:     domain.LocalModuleStatusWarning,
		Issues: []domain.LocalModuleIssue{
			{Kind: "warning", Message: "description contains placeholder text", Field: "description"},
		},
	}
	assert.True(t, v.IsValid())
}
