package graph_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arch-forge/cli/internal/adapter/graph"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRenderMermaid_EmptyGraph verifies that an empty graph produces a valid Mermaid header.
func TestRenderMermaid_EmptyGraph(t *testing.T) {
	var buf bytes.Buffer
	g := domain.DependencyGraph{
		ModulePath: "example.com/test",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
	}

	err := graph.RenderMermaid(&buf, g)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "graph TD")
}

// TestRenderMermaid_WithNodes verifies that nodes appear with correct Mermaid IDs and labels.
func TestRenderMermaid_WithNodes(t *testing.T) {
	var buf bytes.Buffer
	g := domain.DependencyGraph{
		ModulePath: "example.com/test",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
		Nodes: []domain.PackageNode{
			{Path: "internal/domain", Layer: domain.LayerDomain, IsExternal: false},
			{Path: "internal/port", Layer: domain.LayerPort, IsExternal: false},
		},
	}

	err := graph.RenderMermaid(&buf, g)

	require.NoError(t, err)
	output := buf.String()

	// Header must be present.
	assert.Contains(t, output, "graph TD")

	// Node IDs must use underscores instead of slashes.
	assert.Contains(t, output, "internal_domain")
	assert.Contains(t, output, "internal_port")

	// Node labels must use the full path.
	assert.Contains(t, output, `"internal/domain"`)
	assert.Contains(t, output, `"internal/port"`)

	// Subgraph blocks must be present.
	assert.Contains(t, output, "subgraph domain")
	assert.Contains(t, output, "subgraph port")
}

// TestRenderMermaid_WithEdges verifies that edges appear as --> lines.
func TestRenderMermaid_WithEdges(t *testing.T) {
	var buf bytes.Buffer
	g := domain.DependencyGraph{
		ModulePath: "example.com/test",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
		Nodes: []domain.PackageNode{
			{Path: "internal/domain", Layer: domain.LayerDomain},
			{Path: "internal/port", Layer: domain.LayerPort},
		},
		Edges: []domain.PackageEdge{
			{From: "internal/port", To: "internal/domain"},
		},
	}

	err := graph.RenderMermaid(&buf, g)

	require.NoError(t, err)
	output := buf.String()

	// Edge must appear as a --> line.
	assert.True(t,
		strings.Contains(output, "internal_port --> internal_domain"),
		"expected edge line 'internal_port --> internal_domain' in:\n%s", output,
	)
}
