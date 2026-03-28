package app_test

import (
	"errors"
	"testing"

	"github.com/arch-forge/cli/internal/app"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- stub ConfigReader for inspect tests ---

type inspectStubConfigReader struct {
	cfg *port.ProjectConfig
	err error
}

func (s *inspectStubConfigReader) Read(_ string) (*port.ProjectConfig, error) {
	return s.cfg, s.err
}

func (s *inspectStubConfigReader) Write(_ string, _ *port.ProjectConfig) error { return nil }

// --- stub FileSystemScanner ---

type stubScanner struct {
	tree     domain.FileNode
	err      error
	lastOpts port.ScanOptions
}

func (s *stubScanner) Scan(opts port.ScanOptions) (domain.FileNode, error) {
	s.lastOpts = opts
	return s.tree, s.err
}

func TestInspectUseCase_Execute_HappyPath(t *testing.T) {
	cfg := &port.ProjectConfig{
		Name:             "myapp",
		ModulePath:       "github.com/acme/myapp",
		GoVersion:        "1.23",
		Arch:             domain.ArchHexagonal,
		Variant:          domain.VariantClassic,
		InstalledModules: []string{"api", "database"},
	}

	tree := domain.FileNode{
		Name:  "myapp",
		IsDir: true,
		Children: []domain.FileNode{
			{Name: "internal", IsDir: true, Children: []domain.FileNode{
				{Name: "domain", IsDir: true},
				{Name: "app.go", IsDir: false},
			}},
			{Name: "go.mod", IsDir: false},
		},
	}

	sc := &stubScanner{tree: tree}
	uc := app.NewInspectUseCase(&inspectStubConfigReader{cfg: cfg}, sc)

	summary, err := uc.Execute(app.InspectOptions{ProjectDir: t.TempDir(), MaxDepth: 5})

	require.NoError(t, err)
	assert.Equal(t, "myapp", summary.Name)
	assert.Equal(t, "github.com/acme/myapp", summary.ModulePath)
	assert.Equal(t, domain.ArchHexagonal, summary.Arch)
	assert.Equal(t, domain.VariantClassic, summary.Variant)
	assert.Equal(t, "1.23", summary.GoVersion)
	assert.Equal(t, []string{"api", "database"}, summary.InstalledModules)
	assert.Equal(t, 2, summary.Stats.ModuleCount)
	assert.Equal(t, 5, sc.lastOpts.MaxDepth)
}

func TestInspectUseCase_Execute_MissingConfig(t *testing.T) {
	sc := &stubScanner{}
	uc := app.NewInspectUseCase(&inspectStubConfigReader{err: errors.New("file not found")}, sc)

	_, err := uc.Execute(app.InspectOptions{ProjectDir: t.TempDir()})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "inspect:")
}

func TestInspectUseCase_Execute_DefaultDepth(t *testing.T) {
	cfg := &port.ProjectConfig{
		Name:    "proj",
		Arch:    domain.ArchHexagonal,
		Variant: domain.VariantClassic,
	}

	sc := &stubScanner{tree: domain.FileNode{Name: "proj", IsDir: true}}
	uc := app.NewInspectUseCase(&inspectStubConfigReader{cfg: cfg}, sc)

	_, err := uc.Execute(app.InspectOptions{ProjectDir: t.TempDir(), MaxDepth: 0})

	require.NoError(t, err)
	// MaxDepth 0 should default to 3.
	assert.Equal(t, 3, sc.lastOpts.MaxDepth)
}

func TestInspectUseCase_Execute_LayerMapPassedToScanner(t *testing.T) {
	cfg := &port.ProjectConfig{
		Name:    "proj",
		Arch:    domain.ArchHexagonal,
		Variant: domain.VariantClassic,
	}

	sc := &stubScanner{tree: domain.FileNode{Name: "proj", IsDir: true}}
	uc := app.NewInspectUseCase(&inspectStubConfigReader{cfg: cfg}, sc)

	_, err := uc.Execute(app.InspectOptions{ProjectDir: t.TempDir(), MaxDepth: 2})
	require.NoError(t, err)

	// Hexagonal classic should include domain and port in layer map.
	assert.Equal(t, domain.LayerDomain, sc.lastOpts.LayerMap["internal/domain"])
	assert.Equal(t, domain.LayerPort, sc.lastOpts.LayerMap["internal/ports"])
	assert.Equal(t, domain.LayerApp, sc.lastOpts.LayerMap["internal/application"])
	assert.Equal(t, domain.LayerAdapter, sc.lastOpts.LayerMap["internal/adapters"])
}

func TestInspectUseCase_Execute_StatsCountedCorrectly(t *testing.T) {
	cfg := &port.ProjectConfig{
		Name:             "proj",
		Arch:             domain.ArchHexagonal,
		Variant:          domain.VariantClassic,
		InstalledModules: []string{"x"},
	}

	tree := domain.FileNode{
		Name:  "proj",
		IsDir: true,
		Children: []domain.FileNode{
			{Name: "a", IsDir: true, Children: []domain.FileNode{
				{Name: "b.go", IsDir: false},
				{Name: "c.go", IsDir: false},
			}},
			{Name: "main.go", IsDir: false},
		},
	}

	sc := &stubScanner{tree: tree}
	uc := app.NewInspectUseCase(&inspectStubConfigReader{cfg: cfg}, sc)

	summary, err := uc.Execute(app.InspectOptions{ProjectDir: t.TempDir()})
	require.NoError(t, err)

	// Root dir + "a" dir = 2 dirs; b.go, c.go, main.go = 3 files.
	assert.Equal(t, 3, summary.Stats.TotalFiles)
	assert.Equal(t, 2, summary.Stats.TotalDirectories)
	assert.Equal(t, 1, summary.Stats.ModuleCount)
}
