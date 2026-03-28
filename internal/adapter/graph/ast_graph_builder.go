// Package graph provides adapters for building and rendering dependency graphs.
package graph

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
)

// ASTGraphBuilder implements port.GraphBuilder using go/parser and go/ast.
type ASTGraphBuilder struct{}

// NewASTGraphBuilder constructs a new ASTGraphBuilder.
func NewASTGraphBuilder() *ASTGraphBuilder {
	return &ASTGraphBuilder{}
}

// Build walks the project directory, parses Go imports, and returns a DependencyGraph.
func (b *ASTGraphBuilder) Build(req port.GraphBuildRequest) (domain.DependencyGraph, error) {
	modulePrefix := req.ModulePath + "/"

	// edgeSet deduplicates edges by "from|to" key.
	edgeSet := make(map[string]domain.PackageEdge)
	// nodeSet deduplicates nodes by path.
	nodeSet := make(map[string]domain.PackageNode)

	layerMap := buildLayerMapForGraph(req.Arch, req.Variant)

	err := filepath.WalkDir(req.ProjectDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip non-relevant directories.
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process Go source files, skip test files.
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Determine the package path relative to project root.
		dir := filepath.Dir(path)
		relDir, relErr := filepath.Rel(req.ProjectDir, dir)
		if relErr != nil {
			return fmt.Errorf("graph: resolve relative path: %w", relErr)
		}
		// Normalize path separator to forward slashes.
		relDir = filepath.ToSlash(relDir)

		var pkgPath string
		if relDir == "." {
			pkgPath = req.ModulePath
		} else {
			pkgPath = req.ModulePath + "/" + relDir
		}

		// Convert the full module path back to a relative project path for display.
		var pkgRelPath string
		if strings.HasPrefix(pkgPath, modulePrefix) {
			pkgRelPath = strings.TrimPrefix(pkgPath, modulePrefix)
		} else {
			pkgRelPath = relDir
		}

		// Register this package as an internal node.
		if _, exists := nodeSet[pkgRelPath]; !exists {
			layer := resolveLayer(pkgRelPath, layerMap)
			nodeSet[pkgRelPath] = domain.PackageNode{
				Path:       pkgRelPath,
				Layer:      layer,
				IsExternal: false,
			}
		}

		// Parse imports only.
		fset := token.NewFileSet()
		f, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			// Skip files that cannot be parsed.
			return nil
		}

		for _, imp := range f.Imports {
			// Strip surrounding quotes from import path.
			importPath := strings.Trim(imp.Path.Value, `"`)

			var toPath string
			isExternal := false

			if strings.HasPrefix(importPath, modulePrefix) {
				// Internal package — strip module prefix to get relative path.
				toPath = strings.TrimPrefix(importPath, modulePrefix)
			} else if importPath == req.ModulePath {
				// Root package import.
				toPath = ""
			} else {
				// External (third-party or stdlib) package.
				isExternal = true
				toPath = importPath
			}

			if !req.IncludeExternal && isExternal {
				continue
			}

			// Register target node.
			if _, exists := nodeSet[toPath]; !exists {
				var layer domain.ArchLayer
				if isExternal {
					layer = domain.LayerUnknown
				} else {
					layer = resolveLayer(toPath, layerMap)
				}
				nodeSet[toPath] = domain.PackageNode{
					Path:       toPath,
					Layer:      layer,
					IsExternal: isExternal,
				}
			}

			// Register edge.
			edgeKey := pkgRelPath + "|" + toPath
			if _, exists := edgeSet[edgeKey]; !exists {
				edgeSet[edgeKey] = domain.PackageEdge{
					From: pkgRelPath,
					To:   toPath,
				}
			}
		}

		return nil
	})
	if err != nil {
		return domain.DependencyGraph{}, fmt.Errorf("graph: walk project dir: %w", err)
	}

	// Collect nodes and edges from sets.
	nodes := make([]domain.PackageNode, 0, len(nodeSet))
	for _, n := range nodeSet {
		nodes = append(nodes, n)
	}

	edges := make([]domain.PackageEdge, 0, len(edgeSet))
	for _, e := range edgeSet {
		edges = append(edges, e)
	}

	return domain.DependencyGraph{
		ModulePath: req.ModulePath,
		Arch:       req.Arch,
		Variant:    req.Variant,
		Nodes:      nodes,
		Edges:      edges,
	}, nil
}

// resolveLayer looks up the architectural layer for a given package path using the layer map.
// It checks for exact matches and prefix matches (longest prefix wins).
func resolveLayer(pkgPath string, layerMap map[string]domain.ArchLayer) domain.ArchLayer {
	if layer, ok := layerMap[pkgPath]; ok {
		return layer
	}
	// Try prefix matching — longest prefix wins.
	bestLen := -1
	var bestLayer domain.ArchLayer = domain.LayerUnknown
	for prefix, layer := range layerMap {
		if strings.HasPrefix(pkgPath, prefix+"/") && len(prefix) > bestLen {
			bestLen = len(prefix)
			bestLayer = layer
		}
	}
	return bestLayer
}

// buildLayerMapForGraph returns a map of relative directory path → ArchLayer for the given arch/variant.
// This mirrors the buildLayerMap logic in the inspect use case but lives in the graph adapter package.
func buildLayerMapForGraph(arch domain.Architecture, variant domain.Variant) map[string]domain.ArchLayer {
	paths, err := domain.ResolvePaths(arch, variant, "")
	if err != nil {
		return map[string]domain.ArchLayer{}
	}

	m := make(map[string]domain.ArchLayer)

	switch arch {
	case domain.ArchHexagonal, domain.ArchMicroservice:
		setLayerGraph(m, paths.Domain, domain.LayerDomain)
		setLayerGraph(m, paths.Port, domain.LayerPort)
		setLayerGraph(m, paths.App, domain.LayerApp)
		setLayerGraph(m, paths.Adapter, domain.LayerAdapter)

	case domain.ArchClean:
		setLayerGraph(m, paths.Domain, domain.LayerEntities)
		setLayerGraph(m, paths.Port, domain.LayerPort)
		setLayerGraph(m, paths.App, domain.LayerUseCases)
		setLayerGraph(m, paths.Adapter, domain.LayerController)

	case domain.ArchStandard:
		setLayerGraph(m, "cmd", domain.LayerEntrypoint)
		setLayerGraph(m, "internal", domain.LayerInternal)
		setLayerGraph(m, "pkg", domain.LayerPublic)

	case domain.ArchDDD:
		setLayerGraph(m, paths.Domain, domain.LayerDomain)
		setLayerGraph(m, paths.Port, domain.LayerPort)
		setLayerGraph(m, paths.App, domain.LayerApp)
		setLayerGraph(m, paths.Adapter, domain.LayerAdapter)

	case domain.ArchCQRS:
		setLayerGraph(m, paths.Domain, domain.LayerDomain)
		setLayerGraph(m, paths.Port, domain.LayerPort)
		setLayerGraph(m, paths.App, domain.LayerApp)
		setLayerGraph(m, paths.Adapter, domain.LayerAdapter)

	case domain.ArchModularMonolith:
		setLayerGraph(m, "internal", domain.LayerInternal)
	}

	return m
}

// setLayerGraph adds path → layer to m, skipping empty paths.
func setLayerGraph(m map[string]domain.ArchLayer, path string, layer domain.ArchLayer) {
	if path != "" {
		m[path] = layer
	}
}
