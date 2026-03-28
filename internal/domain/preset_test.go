package domain_test

import (
	"testing"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllPresets_NotEmpty(t *testing.T) {
	presets := domain.AllPresets()
	assert.GreaterOrEqual(t, len(presets), 3, "expected at least 3 presets")
}

func TestFindPreset_Found(t *testing.T) {
	names := []string{"starter", "production-api", "microservice"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			preset, found := domain.FindPreset(name)
			require.True(t, found, "expected preset %q to be found", name)
			assert.Equal(t, name, preset.Name)
		})
	}
}

func TestFindPreset_NotFound(t *testing.T) {
	_, found := domain.FindPreset("does-not-exist")
	assert.False(t, found, "expected unknown preset to return false")
}

func TestPreset_ValidArch(t *testing.T) {
	for _, p := range domain.AllPresets() {
		t.Run(p.Name, func(t *testing.T) {
			assert.True(t, p.Arch.IsValid(), "preset %q has invalid architecture %q", p.Name, p.Arch)
		})
	}
}

func TestPreset_ValidVariant(t *testing.T) {
	for _, p := range domain.AllPresets() {
		t.Run(p.Name, func(t *testing.T) {
			assert.True(t, p.Variant.IsValid(), "preset %q has invalid variant %q", p.Name, p.Variant)
		})
	}
}

func TestPreset_HasModules(t *testing.T) {
	for _, p := range domain.AllPresets() {
		t.Run(p.Name, func(t *testing.T) {
			assert.NotEmpty(t, p.Modules, "preset %q should have at least one module", p.Name)
		})
	}
}
