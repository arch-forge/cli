// Package tui provides the interactive terminal wizard for arch_forge init.
package tui

import (
	"fmt"
	"strings"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// step represents a wizard step index.
type step int

const (
	stepArch    step = iota // architecture selection list
	stepVariant             // variant selection list
	stepModules             // module multi-select
	stepConfirm             // confirmation screen
	stepDone                // final state (program exits)
)

// WizardInput holds the pre-filled values passed from the CLI flags.
type WizardInput struct {
	ProjectName string // may be empty if user needs to provide
	ModulePath  string // may be empty
	GoVersion   string // may be empty
}

// WizardResult holds the values collected by the wizard.
type WizardResult struct {
	ProjectName  string
	ModulePath   string
	GoVersion    string
	Architecture domain.Architecture
	Variant      domain.Variant
	Modules      []string
	Confirmed    bool
	Cancelled    bool
}

// moduleEntry holds a module name and its description.
type moduleEntry struct {
	name string
	desc string
}

// availableModules is the fixed list of modules shown in the multi-select step.
var availableModules = []moduleEntry{
	{"api", "HTTP API server with chi router"},
	{"database", "PostgreSQL connection and pool"},
	{"logging", "Structured logging with slog"},
	{"docker", "Multi-stage Dockerfile and docker-compose"},
	{"makefile", "Standard Makefile targets"},
	{"auth", "JWT authentication middleware"},
	{"cache", "Redis cache client"},
	{"grpc", "gRPC server with interceptors"},
	{"crud", "CRUD scaffold for an entity"},
}

// ──────────────────────────────────────────────────────────────────────────────
// list.Item implementations
// ──────────────────────────────────────────────────────────────────────────────

// archListItem wraps domain.ArchInfo to implement list.Item.
type archListItem struct {
	value domain.ArchInfo
}

func (a archListItem) Title() string       { return a.value.DisplayName }
func (a archListItem) Description() string { return a.value.Description }
func (a archListItem) FilterValue() string { return string(a.value.Value) }

// variantListItem wraps domain.VariantInfo to implement list.Item.
type variantListItem struct {
	value domain.VariantInfo
}

func (v variantListItem) Title() string       { return v.value.DisplayName }
func (v variantListItem) Description() string { return v.value.Description }
func (v variantListItem) FilterValue() string { return string(v.value.Value) }

// ──────────────────────────────────────────────────────────────────────────────
// Custom module multi-select model
// ──────────────────────────────────────────────────────────────────────────────

// moduleSelectModel is a simple cursor+toggle list for module selection.
type moduleSelectModel struct {
	choices  []string     // module names
	descs    []string     // descriptions
	cursor   int
	selected map[int]bool
}

// newModuleSelectModel builds a moduleSelectModel from availableModules.
func newModuleSelectModel() moduleSelectModel {
	choices := make([]string, len(availableModules))
	descs := make([]string, len(availableModules))
	for i, m := range availableModules {
		choices[i] = m.name
		descs[i] = m.desc
	}
	return moduleSelectModel{
		choices:  choices,
		descs:    descs,
		cursor:   0,
		selected: make(map[int]bool),
	}
}

// toggle flips the selection state of the item at the given index.
func (m *moduleSelectModel) toggle(idx int) {
	if idx < 0 || idx >= len(m.choices) {
		return
	}
	m.selected[idx] = !m.selected[idx]
}

// selectedModules returns the names of all selected modules.
func (m moduleSelectModel) selectedModules() []string {
	var names []string
	for i, name := range m.choices {
		if m.selected[i] {
			names = append(names, name)
		}
	}
	return names
}

// view renders the module multi-select UI.
func (m moduleSelectModel) view() string {
	var sb strings.Builder
	sb.WriteString("Select modules (space to toggle, enter to confirm):\n\n")
	for i, name := range m.choices {
		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = "[x]"
		}
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		sb.WriteString(fmt.Sprintf("%s%s %-10s %s\n", cursor, checkbox, name, m.descs[i]))
	}
	sb.WriteString("\n(↑↓ to move, space to select, enter to confirm, esc to go back)\n")
	return sb.String()
}

// ──────────────────────────────────────────────────────────────────────────────
// Main wizard model
// ──────────────────────────────────────────────────────────────────────────────

// wizardModel is the top-level Bubble Tea model for the init wizard.
type wizardModel struct {
	step        step
	input       WizardInput
	archList    list.Model
	variantList list.Model
	modules     moduleSelectModel
	result      WizardResult
	width       int
	height      int
}

// newWizardModel constructs the initial wizard model from the given input.
func newWizardModel(input WizardInput) wizardModel {
	// Build architecture list.
	archInfos := domain.AllArchitectures()
	archItems := make([]list.Item, len(archInfos))
	for i, a := range archInfos {
		archItems[i] = archListItem{value: a}
	}
	archL := list.New(archItems, list.NewDefaultDelegate(), 0, 0)
	archL.Title = "Choose an architecture"
	archL.SetShowStatusBar(false)
	archL.SetFilteringEnabled(false)
	archL.SetShowHelp(false)

	// Build variant list.
	variantInfos := domain.AllVariants()
	variantItems := make([]list.Item, len(variantInfos))
	for i, v := range variantInfos {
		variantItems[i] = variantListItem{value: v}
	}
	variantL := list.New(variantItems, list.NewDefaultDelegate(), 0, 0)
	variantL.Title = "Choose a variant"
	variantL.SetShowStatusBar(false)
	variantL.SetFilteringEnabled(false)
	variantL.SetShowHelp(false)

	return wizardModel{
		step:        stepArch,
		input:       input,
		archList:    archL,
		variantList: variantL,
		modules:     newModuleSelectModel(),
		result: WizardResult{
			ProjectName: input.ProjectName,
			ModulePath:  input.ModulePath,
			GoVersion:   input.GoVersion,
		},
	}
}

// Init implements tea.Model. No async commands needed.
func (m wizardModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. Routes messages to the active step handler.
func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle global ctrl+c / window resize for all steps.
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h := msg.Height - 4
		if h < 1 {
			h = 1
		}
		m.archList.SetSize(msg.Width, h)
		m.variantList.SetSize(msg.Width, h)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.result.Cancelled = true
			m.step = stepDone
			return m, tea.Quit
		}
	}

	switch m.step {
	case stepArch:
		return m.updateArch(msg)
	case stepVariant:
		return m.updateVariant(msg)
	case stepModules:
		return m.updateModules(msg)
	case stepConfirm:
		return m.updateConfirm(msg)
	}
	return m, nil
}

// updateArch handles input on the architecture selection step.
func (m wizardModel) updateArch(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			if item, ok := m.archList.SelectedItem().(archListItem); ok {
				m.result.Architecture = item.value.Value
				m.step = stepVariant
				return m, nil
			}
		}
	}
	var cmd tea.Cmd
	m.archList, cmd = m.archList.Update(msg)
	return m, cmd
}

// updateVariant handles input on the variant selection step.
func (m wizardModel) updateVariant(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			if item, ok := m.variantList.SelectedItem().(variantListItem); ok {
				m.result.Variant = item.value.Value
				m.step = stepModules
				return m, nil
			}
		case "esc":
			m.step = stepArch
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.variantList, cmd = m.variantList.Update(msg)
	return m, cmd
}

// updateModules handles input on the module multi-select step.
func (m wizardModel) updateModules(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.modules.cursor > 0 {
				m.modules.cursor--
			}
		case "down", "j":
			if m.modules.cursor < len(m.modules.choices)-1 {
				m.modules.cursor++
			}
		case " ":
			m.modules.toggle(m.modules.cursor)
		case "enter":
			m.result.Modules = m.modules.selectedModules()
			m.step = stepConfirm
		case "esc":
			m.step = stepVariant
		}
	}
	return m, nil
}

// updateConfirm handles input on the confirmation step.
func (m wizardModel) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter", "y":
			m.result.Confirmed = true
			m.step = stepDone
			return m, tea.Quit
		case "esc", "n":
			m.step = stepModules
		}
	}
	return m, nil
}

// View implements tea.Model. Renders the current step.
func (m wizardModel) View() string {
	switch m.step {
	case stepArch:
		return m.archList.View()
	case stepVariant:
		return m.variantList.View()
	case stepModules:
		return m.modules.view()
	case stepConfirm:
		return m.viewConfirm()
	case stepDone:
		return ""
	}
	return ""
}

// viewConfirm renders the confirmation screen.
func (m wizardModel) viewConfirm() string {
	modList := "(none)"
	if len(m.result.Modules) > 0 {
		modList = strings.Join(m.result.Modules, ", ")
	}
	return fmt.Sprintf(
		"Ready to generate your project:\n\n"+
			"  Name:         %s\n"+
			"  Architecture: %s\n"+
			"  Variant:      %s\n"+
			"  Modules:      %s\n\n"+
			"Press enter or y to confirm, esc to go back.\n",
		m.result.ProjectName,
		string(m.result.Architecture),
		string(m.result.Variant),
		modList,
	)
}

// ──────────────────────────────────────────────────────────────────────────────
// Public entry point
// ──────────────────────────────────────────────────────────────────────────────

// Run launches the interactive wizard and returns the collected result.
// input provides any pre-filled values from CLI flags.
func Run(input WizardInput) (WizardResult, error) {
	m := newWizardModel(input)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return WizardResult{Cancelled: true}, fmt.Errorf("run wizard: %w", err)
	}
	wm, ok := finalModel.(wizardModel)
	if !ok {
		return WizardResult{Cancelled: true}, fmt.Errorf("unexpected final model type")
	}
	return wm.result, nil
}
