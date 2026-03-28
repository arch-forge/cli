// Package port defines the interfaces used by the application layer.
package port

import "github.com/arch-forge/cli/internal/domain"

// GraphBuildRequest carries parameters for building a dependency graph.
type GraphBuildRequest struct {
	ProjectDir string
	ModulePath string
	Arch       domain.Architecture
	Variant    domain.Variant
	// IncludeExternal controls whether third-party packages appear as nodes.
	IncludeExternal bool
}

// GraphBuilder analyzes a Go project and returns its package dependency graph.
type GraphBuilder interface {
	Build(req GraphBuildRequest) (domain.DependencyGraph, error)
}
