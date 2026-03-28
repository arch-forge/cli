package graph

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/arch-forge/cli/internal/domain"
)

// RenderMermaid writes a Mermaid flowchart diagram to w.
// Nodes are grouped into subgraphs by their architectural layer.
// External nodes are grouped under a special "external" subgraph.
func RenderMermaid(w io.Writer, g domain.DependencyGraph) error {
	if _, err := fmt.Fprintln(w, "graph TD"); err != nil {
		return fmt.Errorf("mermaid: write header: %w", err)
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

	// Write each subgraph.
	for _, layerKey := range layerKeys {
		nodes := layerNodes[layerKey]
		// Sort nodes within each layer for deterministic output.
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Path < nodes[j].Path
		})

		if _, err := fmt.Fprintf(w, "    subgraph %s\n", layerKey); err != nil {
			return fmt.Errorf("mermaid: write subgraph: %w", err)
		}

		for _, n := range nodes {
			nodeID := toMermaidID(n.Path)
			if _, err := fmt.Fprintf(w, "        %s[\"%s\"]\n", nodeID, n.Path); err != nil {
				return fmt.Errorf("mermaid: write node: %w", err)
			}
		}

		if _, err := fmt.Fprintln(w, "    end"); err != nil {
			return fmt.Errorf("mermaid: write subgraph end: %w", err)
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
		fromID := toMermaidID(e.From)
		toID := toMermaidID(e.To)
		if _, err := fmt.Fprintf(w, "    %s --> %s\n", fromID, toID); err != nil {
			return fmt.Errorf("mermaid: write edge: %w", err)
		}
	}

	return nil
}

// toMermaidID converts a package path to a valid Mermaid node identifier
// by replacing '/', '-', and '.' with '_'.
func toMermaidID(path string) string {
	r := strings.NewReplacer("/", "_", "-", "_", ".", "_")
	id := r.Replace(path)
	if id == "" {
		return "root"
	}
	return id
}
