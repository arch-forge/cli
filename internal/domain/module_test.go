package domain_test

import (
	"testing"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestModule_SupportsArch_EmptyList(t *testing.T) {
	m := domain.Module{
		Name:                   "test-module",
		SupportedArchitectures: nil, // empty means "supports all"
	}
	assert.True(t, m.SupportsArch(domain.ArchHexagonal))
	assert.True(t, m.SupportsArch(domain.ArchClean))
	assert.True(t, m.SupportsArch(domain.ArchDDD))
}

func TestModule_SupportsArch_WithList(t *testing.T) {
	m := domain.Module{
		Name:                   "test-module",
		SupportedArchitectures: []domain.Architecture{domain.ArchHexagonal, domain.ArchClean},
	}
	assert.True(t, m.SupportsArch(domain.ArchHexagonal))
	assert.True(t, m.SupportsArch(domain.ArchClean))
	assert.False(t, m.SupportsArch(domain.ArchDDD))
	assert.False(t, m.SupportsArch(domain.ArchStandard))
}

func TestModule_RequiredDeps(t *testing.T) {
	m := domain.Module{
		Name: "test-module",
		Dependencies: []domain.ModuleDep{
			{Name: "logger", Required: true},
			{Name: "metrics", Required: false},
			{Name: "database", Required: true},
		},
	}
	required := m.RequiredDeps()
	assert.Len(t, required, 2)
	assert.Equal(t, "logger", required[0].Name)
	assert.Equal(t, "database", required[1].Name)
}

func TestModule_RequiredDeps_None(t *testing.T) {
	m := domain.Module{
		Name: "test-module",
		Dependencies: []domain.ModuleDep{
			{Name: "optional-dep", Required: false},
		},
	}
	required := m.RequiredDeps()
	assert.Empty(t, required)
}

func TestModule_SupportsVariant_EmptyList(t *testing.T) {
	m := domain.Module{
		Name:              "test-module",
		SupportedVariants: nil, // empty means "supports all"
	}
	assert.True(t, m.SupportsVariant(domain.VariantClassic))
	assert.True(t, m.SupportsVariant(domain.VariantModular))
}

func TestModule_SupportsVariant_WithList(t *testing.T) {
	m := domain.Module{
		Name:              "test-module",
		SupportedVariants: []domain.Variant{domain.VariantClassic},
	}
	assert.True(t, m.SupportsVariant(domain.VariantClassic))
	assert.False(t, m.SupportsVariant(domain.VariantModular))
}
