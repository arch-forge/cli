package app

import (
	"fmt"
	"path/filepath"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
)

// GraphOptions carries all parameters for the graph workflow.
type GraphOptions struct {
	ProjectDir      string
	Format          string // "mermaid" (default), "dot"
	IncludeExternal bool
}

// GraphUseCase orchestrates building the import dependency graph for a project.
type GraphUseCase struct {
	cfg     port.ConfigReader
	builder port.GraphBuilder
}

// NewGraphUseCase constructs a GraphUseCase.
func NewGraphUseCase(cfg port.ConfigReader, builder port.GraphBuilder) *GraphUseCase {
	return &GraphUseCase{
		cfg:     cfg,
		builder: builder,
	}
}

// Execute runs the full graph workflow:
//  1. Resolve the absolute project directory (defaults to ".").
//  2. Read archforge.yaml for arch/variant/module config.
//  3. Build the dependency graph using the GraphBuilder.
//  4. Return the graph.
func (uc *GraphUseCase) Execute(opts GraphOptions) (domain.DependencyGraph, error) {
	// Step 1 — Resolve absolute project directory.
	projectDir := opts.ProjectDir
	if projectDir == "" {
		projectDir = "."
	}

	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return domain.DependencyGraph{}, fmt.Errorf("graph: resolve project dir: %w", err)
	}

	// Step 2 — Read archforge.yaml.
	cfgPath := filepath.Join(absDir, "archforge.yaml")
	cfg, err := uc.cfg.Read(cfgPath)
	if err != nil {
		return domain.DependencyGraph{}, fmt.Errorf("graph: read config: %w", err)
	}

	// Step 3 — Build the dependency graph.
	req := port.GraphBuildRequest{
		ProjectDir:      absDir,
		ModulePath:      cfg.ModulePath,
		Arch:            cfg.Arch,
		Variant:         cfg.Variant,
		IncludeExternal: opts.IncludeExternal,
	}

	graph, err := uc.builder.Build(req)
	if err != nil {
		return domain.DependencyGraph{}, fmt.Errorf("graph: build graph: %w", err)
	}

	return graph, nil
}
