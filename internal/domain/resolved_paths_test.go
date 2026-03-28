package domain_test

import (
	"errors"
	"testing"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePaths_Hexagonal_Classic(t *testing.T) {
	paths, err := domain.ResolvePaths(domain.ArchHexagonal, domain.VariantClassic, "")
	require.NoError(t, err)
	assert.Equal(t, "internal/domain", paths.Domain)
	assert.Equal(t, "internal/ports", paths.Port)
	assert.Equal(t, "migrations", paths.Migration)
}

func TestResolvePaths_Clean_Classic(t *testing.T) {
	paths, err := domain.ResolvePaths(domain.ArchClean, domain.VariantClassic, "")
	require.NoError(t, err)
	assert.Equal(t, "internal/domain", paths.Domain)
}

func TestResolvePaths_Standard_Classic(t *testing.T) {
	paths, err := domain.ResolvePaths(domain.ArchStandard, domain.VariantClassic, "")
	require.NoError(t, err)
	assert.Equal(t, "internal/handler", paths.Handler)
}

func TestResolvePaths_Modular(t *testing.T) {
	paths, err := domain.ResolvePaths(domain.ArchHexagonal, domain.VariantModular, "auth")
	require.NoError(t, err)
	assert.Equal(t, "internal/auth/domain", paths.Domain)
	assert.Equal(t, "internal/auth/ports", paths.Port)
	assert.Equal(t, "internal/auth/migrations", paths.Migration)
}

func TestResolvePaths_InvalidArch(t *testing.T) {
	_, err := domain.ResolvePaths("invalid", domain.VariantClassic, "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrArchNotSupported))
}

func TestResolvePaths_InvalidVariant(t *testing.T) {
	_, err := domain.ResolvePaths(domain.ArchHexagonal, "invalid", "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrVariantNotSupported))
}
