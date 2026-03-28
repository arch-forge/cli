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

// stubGraphBuilder is a test double for port.GraphBuilder.
type stubGraphBuilder struct {
	graph domain.DependencyGraph
	err   error
}

func (s *stubGraphBuilder) Build(_ port.GraphBuildRequest) (domain.DependencyGraph, error) {
	return s.graph, s.err
}

// stubGraphConfigReader is a test double for port.ConfigReader.
type stubGraphConfigReader struct {
	cfg *port.ProjectConfig
	err error
}

func (s *stubGraphConfigReader) Read(_ string) (*port.ProjectConfig, error) {
	return s.cfg, s.err
}

func (s *stubGraphConfigReader) Write(_ string, _ *port.ProjectConfig) error {
	return nil
}

// TestGraphUseCase_Execute_Success verifies that Execute returns the graph on success.
func TestGraphUseCase_Execute_Success(t *testing.T) {
	cfg := &port.ProjectConfig{
		Name:       "testproject",
		ModulePath: "example.com/testproject",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
	}
	expectedGraph := domain.DependencyGraph{
		ModulePath: "example.com/testproject",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
		Nodes: []domain.PackageNode{
			{Path: "internal/domain", Layer: domain.LayerDomain},
		},
		Edges: []domain.PackageEdge{},
	}

	uc := app.NewGraphUseCase(
		&stubGraphConfigReader{cfg: cfg},
		&stubGraphBuilder{graph: expectedGraph},
	)

	result, err := uc.Execute(app.GraphOptions{
		ProjectDir: ".",
		Format:     "mermaid",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedGraph.ModulePath, result.ModulePath)
	assert.Equal(t, expectedGraph.Arch, result.Arch)
	assert.Len(t, result.Nodes, 1)
}

// TestGraphUseCase_Execute_MissingConfig verifies that Execute returns an error
// when the configuration file cannot be read.
func TestGraphUseCase_Execute_MissingConfig(t *testing.T) {
	configErr := errors.New("archforge.yaml not found")

	uc := app.NewGraphUseCase(
		&stubGraphConfigReader{err: configErr},
		&stubGraphBuilder{},
	)

	_, err := uc.Execute(app.GraphOptions{ProjectDir: "."})

	require.Error(t, err)
	assert.ErrorContains(t, err, "archforge.yaml not found")
}
