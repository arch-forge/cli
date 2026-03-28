// Package domain contains the core entities and business rules for arch_forge.
package domain

// PackageNode represents a Go package in the dependency graph.
type PackageNode struct {
	// Path is the project-relative package path (e.g., "internal/domain").
	Path string
	// Layer is the architectural layer this package belongs to.
	Layer ArchLayer
	// IsExternal is true if the package is outside the project's module (third-party).
	IsExternal bool
}

// PackageEdge represents a directed import dependency between two packages.
type PackageEdge struct {
	// From is the source package path.
	From string
	// To is the target package path.
	To string
}

// DependencyGraph holds the complete import graph for a project.
type DependencyGraph struct {
	ModulePath string
	Arch       Architecture
	Variant    Variant
	Nodes      []PackageNode
	Edges      []PackageEdge
}
