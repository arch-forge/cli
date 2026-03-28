package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/arch-forge/cli/internal/adapter/config"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViperConfig_Read_NotFound(t *testing.T) {
	vc := config.NewViperConfig()
	_, err := vc.Read("/nonexistent/path/archforge.yaml")
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrProjectNotFound))
}

func TestViperConfig_Write_Read_RoundTrip(t *testing.T) {
	vc := config.NewViperConfig()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "archforge.yaml")

	original := &port.ProjectConfig{
		Name:             "myapp",
		ModulePath:       "github.com/acme/myapp",
		GoVersion:        "1.23",
		Arch:             domain.ArchHexagonal,
		Variant:          domain.VariantClassic,
		InstalledModules: []string{"logger", "database"},
	}

	err := vc.Write(cfgPath, original)
	require.NoError(t, err)

	// Verify file was written.
	_, err = os.Stat(cfgPath)
	require.NoError(t, err)

	// Read it back and assert fields match.
	read, err := vc.Read(cfgPath)
	require.NoError(t, err)
	require.NotNil(t, read)

	assert.Equal(t, original.Name, read.Name)
	assert.Equal(t, original.ModulePath, read.ModulePath)
	assert.Equal(t, original.GoVersion, read.GoVersion)
	assert.Equal(t, original.Arch, read.Arch)
	assert.Equal(t, original.Variant, read.Variant)
	assert.Equal(t, original.InstalledModules, read.InstalledModules)
}

func TestViperConfig_Read_InvalidYAML(t *testing.T) {
	vc := config.NewViperConfig()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "archforge.yaml")

	// Write invalid YAML content.
	err := os.WriteFile(cfgPath, []byte("name: [\ninvalid yaml here: {{}"), 0o644)
	require.NoError(t, err)

	_, err = vc.Read(cfgPath)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidConfig))
}

func TestViperConfig_Write_CreatesDirectories(t *testing.T) {
	vc := config.NewViperConfig()

	tmpDir := t.TempDir()
	// Use a nested path that doesn't exist yet.
	cfgPath := filepath.Join(tmpDir, "subdir", "nested", "archforge.yaml")

	cfg := &port.ProjectConfig{
		Name:       "testapp",
		ModulePath: "github.com/test/testapp",
		GoVersion:  "1.23",
		Arch:       domain.ArchClean,
		Variant:    domain.VariantModular,
	}

	err := vc.Write(cfgPath, cfg)
	require.NoError(t, err)

	_, err = os.Stat(cfgPath)
	assert.NoError(t, err, "config file should be created along with intermediate directories")
}
