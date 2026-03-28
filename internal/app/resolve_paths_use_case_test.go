package app_test

import (
	"errors"
	"testing"

	"github.com/arch-forge/cli/internal/app"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePathsUseCase_Hexagonal_Classic(t *testing.T) {
	uc := app.NewResolvePathsUseCase()
	paths, err := uc.Execute(domain.ArchHexagonal, domain.VariantClassic, "")
	require.NoError(t, err)
	assert.Equal(t, "internal/domain", paths.Domain)
	assert.Equal(t, "internal/ports", paths.Port)
	assert.Equal(t, "internal/application", paths.App)
	assert.Equal(t, "internal/adapters", paths.Adapter)
	assert.Equal(t, "migrations", paths.Migration)
}

func TestResolvePathsUseCase_InvalidArch(t *testing.T) {
	uc := app.NewResolvePathsUseCase()
	_, err := uc.Execute("not-an-arch", domain.VariantClassic, "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrArchNotSupported))
}

func TestResolvePathsUseCase_InvalidVariant(t *testing.T) {
	uc := app.NewResolvePathsUseCase()
	_, err := uc.Execute(domain.ArchHexagonal, "not-a-variant", "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrVariantNotSupported))
}

func TestResolvePathsUseCase_Modular(t *testing.T) {
	uc := app.NewResolvePathsUseCase()
	paths, err := uc.Execute(domain.ArchHexagonal, domain.VariantModular, "auth")
	require.NoError(t, err)
	assert.Equal(t, "internal/auth/domain", paths.Domain)
	assert.Equal(t, "internal/auth/ports", paths.Port)
	assert.Equal(t, "internal/auth/application", paths.App)
	assert.Equal(t, "internal/auth/migrations", paths.Migration)
}

func TestResolvePathsUseCase_Clean_Classic(t *testing.T) {
	uc := app.NewResolvePathsUseCase()
	paths, err := uc.Execute(domain.ArchClean, domain.VariantClassic, "")
	require.NoError(t, err)
	assert.Equal(t, "internal/domain", paths.Domain)
	assert.Equal(t, "internal/ports", paths.Port)
}

func TestResolvePathsUseCase_AllArchitectures(t *testing.T) {
	uc := app.NewResolvePathsUseCase()
	for _, info := range domain.AllArchitectures() {
		info := info
		t.Run(string(info.Value), func(t *testing.T) {
			_, err := uc.Execute(info.Value, domain.VariantClassic, "")
			assert.NoError(t, err)
		})
	}
}
