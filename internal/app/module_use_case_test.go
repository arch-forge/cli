package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arch-forge/cli/internal/app"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate_NewModule(t *testing.T) {
	tmp := t.TempDir()
	uc := app.NewModuleUseCase()

	err := uc.Create(app.ModuleCreateOptions{
		Name:     "mymodule",
		Category: "custom",
		Dir:      tmp,
	})
	require.NoError(t, err)

	moduleDir := filepath.Join(tmp, "mymodule")

	// Check directories exist.
	assert.DirExists(t, moduleDir)
	assert.DirExists(t, filepath.Join(moduleDir, "base"))
	assert.DirExists(t, filepath.Join(moduleDir, "hooks"))

	// Check files exist.
	assert.FileExists(t, filepath.Join(moduleDir, "module.yaml"))
	assert.FileExists(t, filepath.Join(moduleDir, "prompts.yaml"))
	assert.FileExists(t, filepath.Join(moduleDir, "base", "README.md"))
	assert.FileExists(t, filepath.Join(moduleDir, "hooks", ".gitkeep"))

	// Check module.yaml content.
	moduleYAML, err := os.ReadFile(filepath.Join(moduleDir, "module.yaml"))
	require.NoError(t, err)
	content := string(moduleYAML)
	assert.Contains(t, content, "name: mymodule")
	assert.Contains(t, content, "category: custom")
	assert.Contains(t, content, "version: 1.0.0")
}

func TestCreate_AlreadyExists(t *testing.T) {
	tmp := t.TempDir()
	uc := app.NewModuleUseCase()

	// Create the directory first.
	moduleDir := filepath.Join(tmp, "existing")
	require.NoError(t, os.MkdirAll(moduleDir, 0o755))

	err := uc.Create(app.ModuleCreateOptions{
		Name: "existing",
		Dir:  tmp,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestValidate_ValidModule(t *testing.T) {
	tmp := t.TempDir()
	uc := app.NewModuleUseCase()

	// Scaffold the module.
	err := uc.Create(app.ModuleCreateOptions{
		Name:     "validmod",
		Category: "custom",
		Dir:      tmp,
	})
	require.NoError(t, err)

	moduleDir := filepath.Join(tmp, "validmod")

	// Fix up module.yaml so name matches dir and description is not placeholder.
	moduleYAML := `name: validmod
version: 1.0.0
description: "A valid test module"
category: custom

architectures:
  - hexagonal
  - clean

variants:
  - classic
  - modular

dependencies:
  required: []
  optional: []

options: []
patches: []
`
	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "module.yaml"), []byte(moduleYAML), 0o644))

	result, err := uc.Validate(app.ModuleValidateOptions{ModuleDir: moduleDir})
	require.NoError(t, err)
	assert.Equal(t, domain.LocalModuleStatusValid, result.Status)
	assert.True(t, result.IsValid())
}

func TestValidate_MissingManifest(t *testing.T) {
	tmp := t.TempDir()
	uc := app.NewModuleUseCase()

	// Create directory without module.yaml.
	moduleDir := filepath.Join(tmp, "nomodule")
	require.NoError(t, os.MkdirAll(moduleDir, 0o755))

	result, err := uc.Validate(app.ModuleValidateOptions{ModuleDir: moduleDir})
	require.NoError(t, err)
	assert.Equal(t, domain.LocalModuleStatusInvalid, result.Status)
	assert.False(t, result.IsValid())

	// Should have an error issue about missing module.yaml.
	hasManifestError := false
	for _, issue := range result.Issues {
		if issue.Kind == "error" && issue.Field == "module.yaml" {
			hasManifestError = true
			break
		}
	}
	assert.True(t, hasManifestError, "expected error issue for missing module.yaml")
}

func TestValidate_InvalidTemplate(t *testing.T) {
	tmp := t.TempDir()
	uc := app.NewModuleUseCase()

	// Scaffold a module first.
	err := uc.Create(app.ModuleCreateOptions{
		Name:     "badtmpl",
		Category: "custom",
		Dir:      tmp,
	})
	require.NoError(t, err)

	moduleDir := filepath.Join(tmp, "badtmpl")

	// Fix module.yaml so it has no other issues.
	moduleYAML := `name: badtmpl
version: 1.0.0
description: "A module with a broken template"
category: custom

architectures:
  - hexagonal

variants:
  - classic

dependencies:
  required: []
  optional: []

options: []
patches: []
`
	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "module.yaml"), []byte(moduleYAML), 0o644))

	// Write a broken template.
	brokenTmpl := `package main

// {{ .Invalid } this is broken template syntax
`
	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "base", "broken.go.tmpl"), []byte(brokenTmpl), 0o644))

	result, err := uc.Validate(app.ModuleValidateOptions{ModuleDir: moduleDir})
	require.NoError(t, err)
	assert.Equal(t, domain.LocalModuleStatusInvalid, result.Status)
	assert.False(t, result.IsValid())

	hasTemplateError := false
	for _, issue := range result.Issues {
		if issue.Kind == "error" && issue.Field != "" {
			hasTemplateError = true
			break
		}
	}
	assert.True(t, hasTemplateError, "expected error issue for broken template")
}

func TestValidate_NonexistentDir(t *testing.T) {
	uc := app.NewModuleUseCase()

	_, err := uc.Validate(app.ModuleValidateOptions{ModuleDir: "/nonexistent/path/to/module"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}
