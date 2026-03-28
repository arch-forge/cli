package graph_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arch-forge/cli/internal/adapter/graph"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testModulePath = "example.com/testproject"

func makeRequest(dir string, includeExternal bool) port.GraphBuildRequest {
	return port.GraphBuildRequest{
		ProjectDir:      dir,
		ModulePath:      testModulePath,
		Arch:            domain.ArchHexagonal,
		Variant:         domain.VariantClassic,
		IncludeExternal: includeExternal,
	}
}

func writeFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(dir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
}

// TestBuild_EmptyProject verifies that an empty directory produces an empty graph.
func TestBuild_EmptyProject(t *testing.T) {
	dir := t.TempDir()
	builder := graph.NewASTGraphBuilder()

	g, err := builder.Build(makeRequest(dir, false))

	require.NoError(t, err)
	assert.Empty(t, g.Nodes)
	assert.Empty(t, g.Edges)
	assert.Equal(t, testModulePath, g.ModulePath)
}

// TestBuild_SinglePackage verifies that a single Go file with no imports
// produces exactly one node and no edges.
func TestBuild_SinglePackage(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "internal/domain/entity.go", `package domain
`)

	builder := graph.NewASTGraphBuilder()
	g, err := builder.Build(makeRequest(dir, false))

	require.NoError(t, err)
	assert.Len(t, g.Nodes, 1)
	assert.Equal(t, "internal/domain", g.Nodes[0].Path)
	assert.Empty(t, g.Edges)
}

// TestBuild_InternalImport verifies that an import between two internal packages
// produces the correct edge.
func TestBuild_InternalImport(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "internal/domain/entity.go", `package domain
`)
	writeFile(t, dir, "internal/port/service.go", `package port

import _ "example.com/testproject/internal/domain"
`)

	builder := graph.NewASTGraphBuilder()
	g, err := builder.Build(makeRequest(dir, false))

	require.NoError(t, err)

	// Find nodes by path.
	nodesByPath := make(map[string]domain.PackageNode)
	for _, n := range g.Nodes {
		nodesByPath[n.Path] = n
	}
	assert.Contains(t, nodesByPath, "internal/domain")
	assert.Contains(t, nodesByPath, "internal/port")
	assert.False(t, nodesByPath["internal/domain"].IsExternal)
	assert.False(t, nodesByPath["internal/port"].IsExternal)

	// Find the edge from port -> domain.
	edgeFound := false
	for _, e := range g.Edges {
		if e.From == "internal/port" && e.To == "internal/domain" {
			edgeFound = true
			break
		}
	}
	assert.True(t, edgeFound, "expected edge from internal/port to internal/domain")
}

// TestBuild_ExternalImport verifies that external imports are excluded by default.
func TestBuild_ExternalImport(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "internal/domain/entity.go", `package domain

import _ "fmt"
`)

	builder := graph.NewASTGraphBuilder()
	g, err := builder.Build(makeRequest(dir, false))

	require.NoError(t, err)

	// Only the internal package node should appear; "fmt" should be excluded.
	for _, n := range g.Nodes {
		assert.False(t, n.IsExternal, "no external nodes expected when IncludeExternal=false")
	}
	assert.Empty(t, g.Edges, "no edges expected since external target is filtered")
}

// TestBuild_ExternalImport_Included verifies that external imports appear as nodes
// when IncludeExternal is true.
func TestBuild_ExternalImport_Included(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "internal/domain/entity.go", `package domain

import _ "fmt"
`)

	builder := graph.NewASTGraphBuilder()
	g, err := builder.Build(makeRequest(dir, true))

	require.NoError(t, err)

	externalFound := false
	for _, n := range g.Nodes {
		if n.Path == "fmt" && n.IsExternal {
			externalFound = true
		}
	}
	assert.True(t, externalFound, "expected external node for 'fmt'")

	// The edge from internal/domain -> fmt should exist.
	edgeFound := false
	for _, e := range g.Edges {
		if e.From == "internal/domain" && e.To == "fmt" {
			edgeFound = true
			break
		}
	}
	assert.True(t, edgeFound, "expected edge from internal/domain to fmt")
}
