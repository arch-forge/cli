package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/arch-forge/cli/internal/domain"
)

// ANSI color escape codes — no external dependencies.
const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorRed    = "\033[31m" //nolint:deadcode,varcheck // reserved for future use
)

// isTerminal reports whether f is connected to a terminal.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// RenderTree writes an annotated directory tree to w.
func RenderTree(w io.Writer, summary domain.ProjectSummary, useColor bool) error {
	// Header.
	modules := "none"
	if len(summary.InstalledModules) > 0 {
		modules = strings.Join(summary.InstalledModules, ", ")
	}

	separator := strings.Repeat("─", 45)

	fmt.Fprintf(w, "Project:      %s\n", summary.Name)
	fmt.Fprintf(w, "Module path:  %s\n", summary.ModulePath)
	fmt.Fprintf(w, "Architecture: %s (%s)\n", summary.Arch, summary.Variant)
	fmt.Fprintf(w, "Go version:   %s\n", summary.GoVersion)
	fmt.Fprintf(w, "Modules:      %s\n", modules)
	fmt.Fprintln(w, separator)

	// Tree root.
	fmt.Fprintln(w, summary.Tree.Name)

	// Render children of root.
	for i, child := range summary.Tree.Children {
		isLast := i == len(summary.Tree.Children)-1
		renderNode(w, child, "", isLast, useColor)
	}

	// Footer.
	fmt.Fprintln(w, separator)
	fmt.Fprintf(w, "%d files  |  %d directories  |  %d modules\n",
		summary.Stats.TotalFiles,
		summary.Stats.TotalDirectories,
		summary.Stats.ModuleCount,
	)

	return nil
}

// renderNode recursively writes a single FileNode with box-drawing characters.
func renderNode(w io.Writer, node domain.FileNode, prefix string, isLast bool, useColor bool) {
	// Choose connector.
	connector := "├── "
	childPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		childPrefix = prefix + "    "
	}

	// Build display name with optional color.
	name := node.Name
	if useColor && node.IsDir {
		name = colorBlue + name + colorReset
	}

	// Build annotation string.
	annotation := buildAnnotation(node, useColor)

	fmt.Fprintf(w, "%s%s%s%s\n", prefix, connector, name, annotation)

	// Recurse into children.
	for i, child := range node.Children {
		isChildLast := i == len(node.Children)-1
		renderNode(w, child, childPrefix, isChildLast, useColor)
	}
}

// buildAnnotation constructs the layer and module annotation suffix for a node.
func buildAnnotation(node domain.FileNode, useColor bool) string {
	var sb strings.Builder

	if node.Layer != domain.LayerUnknown {
		layerStr := string(node.Layer)
		if useColor {
			layerStr = layerColor(node.Layer) + layerStr + colorReset
		}
		sb.WriteString("  [")
		sb.WriteString(layerStr)
		sb.WriteString("]")
	}

	if node.ModuleOwner != "" {
		sb.WriteString("  {mod:")
		sb.WriteString(node.ModuleOwner)
		sb.WriteString("}")
	}

	return sb.String()
}

// layerColor returns the ANSI color for the given layer.
func layerColor(layer domain.ArchLayer) string {
	switch layer {
	case domain.LayerDomain, domain.LayerEntities:
		return colorCyan
	case domain.LayerPort:
		return colorBlue
	case domain.LayerApp, domain.LayerUseCases:
		return colorGreen
	case domain.LayerAdapter, domain.LayerController:
		return colorYellow
	default:
		return colorReset
	}
}

// RenderJSON writes the ProjectSummary as indented JSON to w.
func RenderJSON(w io.Writer, summary domain.ProjectSummary) error {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("render json: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
