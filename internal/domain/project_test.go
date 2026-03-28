package domain_test

import (
	"testing"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProject_HasModule(t *testing.T) {
	p := domain.Project{
		Name:             "myapp",
		InstalledModules: []string{"logger", "database", "api"},
	}

	assert.True(t, p.HasModule("logger"))
	assert.True(t, p.HasModule("database"))
	assert.True(t, p.HasModule("api"))
	assert.False(t, p.HasModule("metrics"))
	assert.False(t, p.HasModule(""))
}

func TestProject_AddModule(t *testing.T) {
	original := domain.Project{
		Name:             "myapp",
		InstalledModules: []string{"logger"},
	}

	updated := original.AddModule("database")

	// New project has the module appended.
	require.Len(t, updated.InstalledModules, 2)
	assert.Equal(t, "logger", updated.InstalledModules[0])
	assert.Equal(t, "database", updated.InstalledModules[1])

	// Original project is unchanged (value semantics).
	require.Len(t, original.InstalledModules, 1)
	assert.Equal(t, "logger", original.InstalledModules[0])
}

func TestProject_AddModule_EmptyInitial(t *testing.T) {
	p := domain.Project{Name: "myapp"}
	updated := p.AddModule("logger")
	require.Len(t, updated.InstalledModules, 1)
	assert.Equal(t, "logger", updated.InstalledModules[0])
}
