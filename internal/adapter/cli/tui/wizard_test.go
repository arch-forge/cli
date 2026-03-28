package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWizardResult_Cancelled verifies that a cancelled WizardResult has the
// expected zero/false values for all fields except Cancelled.
func TestWizardResult_Cancelled(t *testing.T) {
	r := WizardResult{Cancelled: true}

	assert.True(t, r.Cancelled)
	assert.False(t, r.Confirmed)
	assert.Empty(t, r.ProjectName)
	assert.Empty(t, r.Architecture)
	assert.Empty(t, r.Variant)
	assert.Empty(t, r.Modules)
}

// TestModuleSelectModel_Toggle verifies that toggling adds and removes a module
// from the selected map.
func TestModuleSelectModel_Toggle(t *testing.T) {
	m := newModuleSelectModel()

	// Initially nothing is selected.
	assert.False(t, m.selected[0])

	// First toggle: select index 0.
	m.toggle(0)
	assert.True(t, m.selected[0])

	// Second toggle: deselect index 0.
	m.toggle(0)
	assert.False(t, m.selected[0])
}

// TestModuleSelectModel_SelectedModules verifies that selectedModules returns
// only the names of modules that have been toggled on.
func TestModuleSelectModel_SelectedModules(t *testing.T) {
	m := newModuleSelectModel()

	// Select indices 0 ("api") and 2 ("logging").
	m.toggle(0)
	m.toggle(2)

	selected := m.selectedModules()
	assert.Len(t, selected, 2)
	assert.Contains(t, selected, "api")
	assert.Contains(t, selected, "logging")
	assert.NotContains(t, selected, "database")
}

// TestWizardInput_Defaults verifies that an empty WizardInput has sensible zero values.
func TestWizardInput_Defaults(t *testing.T) {
	var input WizardInput

	assert.Empty(t, input.ProjectName)
	assert.Empty(t, input.ModulePath)
	assert.Empty(t, input.GoVersion)
}
