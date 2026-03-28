package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arch-forge/cli/internal/adapter/scanner"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mkdirAll creates directories under root for convenience.
func mkdirAll(t *testing.T, root string, dirs ...string) {
	t.Helper()
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(filepath.Join(root, d), 0o755))
	}
}

// touchFile creates an empty file at path.
func touchFile(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, nil, 0o644))
}

// findNode returns the first FileNode with the given name, or nil.
func findNode(nodes []domain.FileNode, name string) *domain.FileNode {
	for i := range nodes {
		if nodes[i].Name == name {
			return &nodes[i]
		}
	}
	return nil
}

func TestScan_HappyPath(t *testing.T) {
	root := t.TempDir()

	mkdirAll(t, root, "internal/domain", "internal/port", "cmd")
	touchFile(t, filepath.Join(root, "go.mod"))
	touchFile(t, filepath.Join(root, "internal/domain/model.go"))

	s := scanner.NewOsScanner()
	tree, err := s.Scan(port.ScanOptions{
		RootDir:  root,
		MaxDepth: 0, // unlimited
	})

	require.NoError(t, err)
	assert.True(t, tree.IsDir)
	assert.NotEmpty(t, tree.Children)

	// "internal" should appear before files.
	internalNode := findNode(tree.Children, "internal")
	require.NotNil(t, internalNode, "expected 'internal' directory node")
	assert.True(t, internalNode.IsDir)

	goModNode := findNode(tree.Children, "go.mod")
	require.NotNil(t, goModNode, "expected 'go.mod' file node")
	assert.False(t, goModNode.IsDir)
}

func TestScan_DirsBeforeFiles(t *testing.T) {
	root := t.TempDir()

	mkdirAll(t, root, "aaa")
	touchFile(t, filepath.Join(root, "zzz.go"))
	touchFile(t, filepath.Join(root, "aaa.go"))

	s := scanner.NewOsScanner()
	tree, err := s.Scan(port.ScanOptions{
		RootDir:  root,
		MaxDepth: 1,
	})

	require.NoError(t, err)
	require.Len(t, tree.Children, 3) // aaa (dir), aaa.go, zzz.go

	assert.True(t, tree.Children[0].IsDir, "first child should be directory 'aaa'")
	assert.Equal(t, "aaa", tree.Children[0].Name)
	assert.False(t, tree.Children[1].IsDir)
	assert.Equal(t, "aaa.go", tree.Children[1].Name)
	assert.Equal(t, "zzz.go", tree.Children[2].Name)
}

func TestScan_DepthLimit(t *testing.T) {
	root := t.TempDir()

	// Create three levels deep: a/b/c/deep.go
	mkdirAll(t, root, "a/b/c")
	touchFile(t, filepath.Join(root, "a/b/c/deep.go"))

	s := scanner.NewOsScanner()

	// With maxDepth=2 the scanner should stop before descending into "b/c".
	tree, err := s.Scan(port.ScanOptions{
		RootDir:  root,
		MaxDepth: 2,
	})

	require.NoError(t, err)

	// Navigate: root → a → b
	aNode := findNode(tree.Children, "a")
	require.NotNil(t, aNode)

	bNode := findNode(aNode.Children, "b")
	require.NotNil(t, bNode)

	// "b" is at depth 2, so it should NOT have children resolved.
	assert.Empty(t, bNode.Children, "b should have no children at maxDepth=2")
}

func TestScan_SkipDotGit(t *testing.T) {
	root := t.TempDir()

	mkdirAll(t, root, ".git/objects", "internal")
	touchFile(t, filepath.Join(root, ".git/HEAD"))
	touchFile(t, filepath.Join(root, "internal/app.go"))

	s := scanner.NewOsScanner()
	tree, err := s.Scan(port.ScanOptions{
		RootDir:  root,
		MaxDepth: 0,
	})

	require.NoError(t, err)

	gitNode := findNode(tree.Children, ".git")
	assert.Nil(t, gitNode, ".git should be skipped by default")

	internalNode := findNode(tree.Children, "internal")
	assert.NotNil(t, internalNode)
}

func TestScan_LayerAnnotation(t *testing.T) {
	root := t.TempDir()

	mkdirAll(t, root, "internal/domain", "internal/port")

	layerMap := map[string]domain.ArchLayer{
		"internal/domain": domain.LayerDomain,
		"internal/port":   domain.LayerPort,
	}

	s := scanner.NewOsScanner()
	tree, err := s.Scan(port.ScanOptions{
		RootDir:  root,
		MaxDepth: 0,
		LayerMap: layerMap,
	})

	require.NoError(t, err)

	internalNode := findNode(tree.Children, "internal")
	require.NotNil(t, internalNode)

	domainNode := findNode(internalNode.Children, "domain")
	require.NotNil(t, domainNode)
	assert.Equal(t, domain.LayerDomain, domainNode.Layer)

	portNode := findNode(internalNode.Children, "port")
	require.NotNil(t, portNode)
	assert.Equal(t, domain.LayerPort, portNode.Layer)
}

func TestScan_CustomSkipDirs(t *testing.T) {
	root := t.TempDir()

	mkdirAll(t, root, "vendor/pkg", "internal")
	touchFile(t, filepath.Join(root, "internal/main.go"))

	s := scanner.NewOsScanner()
	tree, err := s.Scan(port.ScanOptions{
		RootDir:  root,
		MaxDepth: 0,
		SkipDirs: []string{"vendor"},
	})

	require.NoError(t, err)

	vendorNode := findNode(tree.Children, "vendor")
	assert.Nil(t, vendorNode, "vendor should be skipped")
}
