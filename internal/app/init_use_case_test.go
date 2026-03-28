package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/arch-forge/cli/internal/adapter/generator"
	"github.com/arch-forge/cli/internal/app"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRepository is a test double for port.TemplateRepository.
type mockRepository struct {
	archTemplates   []port.TemplateFile
	moduleTemplates map[string][]port.TemplateFile
	manifests       map[string]domain.Module
}

func (m *mockRepository) LoadArchTemplates(arch, variant string) ([]port.TemplateFile, error) {
	return m.archTemplates, nil
}

func (m *mockRepository) LoadModuleTemplates(moduleName, arch, variant string) ([]port.TemplateFile, error) {
	templates, ok := m.moduleTemplates[moduleName]
	if !ok {
		return nil, nil
	}
	return templates, nil
}

func (m *mockRepository) LoadModuleManifest(name string) (domain.Module, error) {
	mod, ok := m.manifests[name]
	if !ok {
		return domain.Module{}, fmt.Errorf("%w: %s", domain.ErrModuleNotFound, name)
	}
	return mod, nil
}

func (m *mockRepository) ListModules() ([]string, error) {
	names := make([]string, 0, len(m.manifests))
	for k := range m.manifests {
		names = append(names, k)
	}
	return names, nil
}

func (m *mockRepository) LoadDomainTemplates(arch, variant string) ([]port.TemplateFile, error) {
	return nil, nil
}

// mockConfigWriter is a test double for port.ConfigReader that records writes.
type mockConfigWriter struct {
	written *port.ProjectConfig
	writErr error
}

func (c *mockConfigWriter) Read(path string) (*port.ProjectConfig, error) {
	return nil, domain.ErrProjectNotFound
}

func (c *mockConfigWriter) Write(path string, cfg *port.ProjectConfig) error {
	if c.writErr != nil {
		return c.writErr
	}
	c.written = cfg
	return nil
}

func newTestInitUseCase(repo port.TemplateRepository, cfgWriter port.ConfigReader) *app.InitUseCase {
	eng := generator.NewEngine()
	return app.NewInitUseCase(repo, eng, cfgWriter)
}

func TestInitUseCase_Execute_DryRun(t *testing.T) {
	repo := &mockRepository{
		archTemplates:   []port.TemplateFile{},
		moduleTemplates: map[string][]port.TemplateFile{},
		manifests:       map[string]domain.Module{},
	}
	cfgWriter := &mockConfigWriter{}

	uc := newTestInitUseCase(repo, cfgWriter)

	opts := app.InitOptions{
		Name:       "myapp",
		ModulePath: "github.com/test/myapp",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantModular,
		Modules:    nil,
		DryRun:     true,
		OutputDir:  "/tmp/unused",
		GoVersion:  "1.23",
	}

	err := uc.Execute(context.Background(), opts)
	require.NoError(t, err)

	// DryRun=true — config should NOT have been written.
	assert.Nil(t, cfgWriter.written, "config should not be written during dry run")
}

func TestInitUseCase_Execute_GeneratesFiles(t *testing.T) {
	repo := &mockRepository{
		archTemplates: []port.TemplateFile{
			{
				RelPath: "go.mod.tmpl",
				Content: []byte("module {{ .Module }}\n\ngo {{ .GoVersion }}"),
			},
			{
				RelPath: "cmd/main.go.tmpl",
				Content: []byte("package main\n\nfunc main() {}"),
			},
		},
		moduleTemplates: map[string][]port.TemplateFile{},
		manifests:       map[string]domain.Module{},
	}
	cfgWriter := &mockConfigWriter{}

	uc := newTestInitUseCase(repo, cfgWriter)

	tmpDir := t.TempDir()
	opts := app.InitOptions{
		Name:       "testapp",
		ModulePath: "github.com/test/testapp",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
		Modules:    nil,
		DryRun:     false,
		OutputDir:  tmpDir,
		GoVersion:  "1.23",
	}

	err := uc.Execute(context.Background(), opts)
	require.NoError(t, err)

	// Config should have been written.
	require.NotNil(t, cfgWriter.written)
	assert.Equal(t, "testapp", cfgWriter.written.Name)
	assert.Equal(t, domain.ArchHexagonal, cfgWriter.written.Arch)
}

func TestInitUseCase_Execute_InvalidArch(t *testing.T) {
	repo := &mockRepository{
		archTemplates:   []port.TemplateFile{},
		moduleTemplates: map[string][]port.TemplateFile{},
		manifests:       map[string]domain.Module{},
	}
	cfgWriter := &mockConfigWriter{}

	uc := newTestInitUseCase(repo, cfgWriter)

	opts := app.InitOptions{
		Name:       "myapp",
		ModulePath: "github.com/test/myapp",
		Arch:       "not-valid-arch",
		Variant:    domain.VariantClassic,
		DryRun:     true,
	}

	err := uc.Execute(context.Background(), opts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrArchNotSupported))
}

func TestInitUseCase_Execute_InvalidVariant(t *testing.T) {
	repo := &mockRepository{
		archTemplates:   []port.TemplateFile{},
		moduleTemplates: map[string][]port.TemplateFile{},
		manifests:       map[string]domain.Module{},
	}
	cfgWriter := &mockConfigWriter{}

	uc := newTestInitUseCase(repo, cfgWriter)

	opts := app.InitOptions{
		Name:       "myapp",
		ModulePath: "github.com/test/myapp",
		Arch:       domain.ArchHexagonal,
		Variant:    "not-valid-variant",
		DryRun:     true,
	}

	err := uc.Execute(context.Background(), opts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrVariantNotSupported))
}

func TestInitUseCase_Execute_WithModule(t *testing.T) {
	loggerModule := domain.Module{
		Name:                   "logger",
		SupportedArchitectures: nil,
		SupportedVariants:      nil,
		Dependencies:           nil,
	}

	repo := &mockRepository{
		archTemplates: []port.TemplateFile{},
		moduleTemplates: map[string][]port.TemplateFile{
			"logger": {
				{
					RelPath: "internal/logger/logger.go.tmpl",
					Content: []byte("package logger"),
				},
			},
		},
		manifests: map[string]domain.Module{
			"logger": loggerModule,
		},
	}
	cfgWriter := &mockConfigWriter{}

	uc := newTestInitUseCase(repo, cfgWriter)

	opts := app.InitOptions{
		Name:       "myapp",
		ModulePath: "github.com/test/myapp",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
		Modules:    []string{"logger"},
		DryRun:     false,
		OutputDir:  t.TempDir(),
		GoVersion:  "1.23",
	}

	err := uc.Execute(context.Background(), opts)
	require.NoError(t, err)

	require.NotNil(t, cfgWriter.written)
	assert.Contains(t, cfgWriter.written.InstalledModules, "logger")
}

func TestInitUseCase_Execute_ModuleNotFound(t *testing.T) {
	repo := &mockRepository{
		archTemplates:   []port.TemplateFile{},
		moduleTemplates: map[string][]port.TemplateFile{},
		manifests:       map[string]domain.Module{},
	}
	cfgWriter := &mockConfigWriter{}

	uc := newTestInitUseCase(repo, cfgWriter)

	opts := app.InitOptions{
		Name:       "myapp",
		ModulePath: "github.com/test/myapp",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
		Modules:    []string{"nonexistent-module"},
		DryRun:     true,
	}

	err := uc.Execute(context.Background(), opts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrModuleNotFound))
}

func TestInitUseCase_WithPreset_SetsArchAndModules(t *testing.T) {
	// The "starter" preset resolves to ArchStandard + VariantModular + modules [api, logging, docker, makefile].
	// We provide no arch, variant, or modules so the preset values must be applied.
	repo := &mockRepository{
		archTemplates:   []port.TemplateFile{},
		moduleTemplates: map[string][]port.TemplateFile{},
		// Register all modules that the starter preset includes.
		manifests: map[string]domain.Module{
			"api":      {Name: "api"},
			"logging":  {Name: "logging"},
			"docker":   {Name: "docker"},
			"makefile": {Name: "makefile"},
		},
	}
	cfgWriter := &mockConfigWriter{}

	uc := newTestInitUseCase(repo, cfgWriter)

	opts := app.InitOptions{
		Name:      "myapp",
		ModulePath: "github.com/test/myapp",
		// Arch, Variant, and Modules intentionally left empty — preset fills them.
		Preset:    "starter",
		DryRun:    false,
		OutputDir: t.TempDir(),
		GoVersion: "1.23",
	}

	err := uc.Execute(context.Background(), opts)
	require.NoError(t, err)

	require.NotNil(t, cfgWriter.written)
	assert.Equal(t, domain.ArchStandard, cfgWriter.written.Arch)
	assert.Equal(t, domain.VariantModular, cfgWriter.written.Variant)
	assert.ElementsMatch(t, []string{"api", "logging", "docker", "makefile"}, cfgWriter.written.InstalledModules)
}

func TestInitUseCase_WithPreset_OverriddenByFlags(t *testing.T) {
	// The "starter" preset uses ArchStandard + VariantModular + starter modules.
	// We explicitly set Arch and Variant to override those preset values.
	// Modules are left empty so the preset's modules apply (they must exist in the repo).
	repo := &mockRepository{
		archTemplates:   []port.TemplateFile{},
		moduleTemplates: map[string][]port.TemplateFile{},
		// Register the starter preset modules so dependency resolution succeeds.
		manifests: map[string]domain.Module{
			"api":      {Name: "api"},
			"logging":  {Name: "logging"},
			"docker":   {Name: "docker"},
			"makefile": {Name: "makefile"},
		},
	}
	cfgWriter := &mockConfigWriter{}

	uc := newTestInitUseCase(repo, cfgWriter)

	opts := app.InitOptions{
		Name:       "myapp",
		ModulePath: "github.com/test/myapp",
		Arch:       domain.ArchHexagonal, // explicit override of preset's ArchStandard
		Variant:    domain.VariantClassic, // explicit override of preset's VariantModular
		Preset:     "starter",
		DryRun:     false,
		OutputDir:  t.TempDir(),
		GoVersion:  "1.23",
	}

	err := uc.Execute(context.Background(), opts)
	require.NoError(t, err)

	require.NotNil(t, cfgWriter.written)
	// Explicit arch and variant should win over the preset's values.
	assert.Equal(t, domain.ArchHexagonal, cfgWriter.written.Arch)
	assert.Equal(t, domain.VariantClassic, cfgWriter.written.Variant)
}

func TestInitUseCase_UnknownPreset_ReturnsError(t *testing.T) {
	repo := &mockRepository{
		archTemplates:   []port.TemplateFile{},
		moduleTemplates: map[string][]port.TemplateFile{},
		manifests:       map[string]domain.Module{},
	}
	cfgWriter := &mockConfigWriter{}

	uc := newTestInitUseCase(repo, cfgWriter)

	opts := app.InitOptions{
		Name:      "myapp",
		ModulePath: "github.com/test/myapp",
		Preset:    "nonexistent-preset",
		DryRun:    true,
	}

	err := uc.Execute(context.Background(), opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-preset")
}

func TestInitUseCase_Execute_DefaultGoVersion(t *testing.T) {
	repo := &mockRepository{
		archTemplates:   []port.TemplateFile{},
		moduleTemplates: map[string][]port.TemplateFile{},
		manifests:       map[string]domain.Module{},
	}
	cfgWriter := &mockConfigWriter{}

	uc := newTestInitUseCase(repo, cfgWriter)

	opts := app.InitOptions{
		Name:       "myapp",
		ModulePath: "github.com/test/myapp",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
		DryRun:     false,
		OutputDir:  t.TempDir(),
		GoVersion:  "", // should default to "1.23"
	}

	err := uc.Execute(context.Background(), opts)
	require.NoError(t, err)

	require.NotNil(t, cfgWriter.written)
	assert.Equal(t, "1.23", cfgWriter.written.GoVersion)
}
