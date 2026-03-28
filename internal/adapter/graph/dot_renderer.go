package graph

import (
	"fmt"
	"io"
	"sort"

	"github.com/arch-forge/cli/internal/domain"
)

// RenderDOT writes a Graphviz DOT format diagram to w.
// Nodes are grouped into clusters by their architectural layer.
// External nodes are grouped under a special "external" cluster.
func RenderDOT(w io.Writer, g domain.DependencyGraph) error {
	if _, err := fmt.Fprintln(w, "digraph dependencies {"); err != nil {
		return fmt.Errorf("dot: write header: %w", err)
	}
	if _, err := fmt.Fprintln(w, "    rankdir=TB;"); err != nil {
		return fmt.Errorf("dot: write rankdir: %w", err)
	}
	if _, err := fmt.Fprintln(w, "    node [shape=box];"); err != nil {
		return fmt.Errorf("dot: write node attr: %w", err)
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("dot: write blank line: %w", err)
	}

	// Group nodes by layer.
	layerNodes := make(map[string][]domain.PackageNode)
	for _, n := range g.Nodes {
		groupKey := string(n.Layer)
		if n.IsExternal {
			groupKey = "external"
		}
		if groupKey == "" {
			groupKey = "unknown"
		}
		layerNodes[groupKey] = append(layerNodes[groupKey], n)
	}

	// Sort layer keys for deterministic output.
	layerKeys := make([]string, 0, len(layerNodes))
	for k := range layerNodes {
		layerKeys = append(layerKeys, k)
	}
	sort.Strings(layerKeys)

	// Write each cluster subgraph.
	for _, layerKey := range layerKeys {
		nodes := layerNodes[layerKey]
		// Sort nodes within each cluster for deterministic output.
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Path < nodes[j].Path
		})

		if _, err := fmt.Fprintf(w, "    subgraph cluster_%s {\n", layerKey); err != nil {
			return fmt.Errorf("dot: write cluster: %w", err)
		}
		if _, err := fmt.Fprintf(w, "        label=\"%s\";\n", layerKey); err != nil {
			return fmt.Errorf("dot: write cluster label: %w", err)
		}

		for _, n := range nodes {
			if _, err := fmt.Fprintf(w, "        \"%s\";\n", n.Path); err != nil {
				return fmt.Errorf("dot: write node: %w", err)
			}
		}

		if _, err := fmt.Fprintln(w, "    }"); err != nil {
			return fmt.Errorf("dot: write cluster end: %w", err)
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("dot: write blank line: %w", err)
		}
	}

	// Write edges, sorted for deterministic output.
	edges := make([]domain.PackageEdge, len(g.Edges))
	copy(edges, g.Edges)
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})

	for _, e := range edges {
		if _, err := fmt.Fprintf(w, "    \"%s\" -> \"%s\";\n", e.From, e.To); err != nil {
			return fmt.Errorf("dot: write edge: %w", err)
		}
	}

	if _, err := fmt.Fprintln(w, "}"); err != nil {
		return fmt.Errorf("dot: write footer: %w", err)
	}

	return nil
}
