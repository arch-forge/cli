package port

import (
	"context"

	"github.com/spf13/afero"
)

// PatchRequest bundles all information needed to apply one patch
// to a set of matched files.
type PatchRequest struct {
	// TargetGlob is a glob pattern relative to the project root.
	TargetGlob string
	// Action is the patch strategy: "inject_after" | "inject_before" | "replace".
	Action string
	// Anchor is the arch_forge comment tag name, e.g. "imports" for "// arch_forge:imports".
	Anchor string
	// Content is the pre-rendered text to inject.
	Content string
	// Optional means the patcher silently skips if no files match or anchor not found.
	Optional bool
}

// Patcher applies anchor-based text patches to source files in a filesystem.
type Patcher interface {
	// Apply executes all PatchRequests against files inside fs.
	// Each request's TargetGlob is resolved relative to rootDir.
	Apply(ctx context.Context, rootDir string, patches []PatchRequest, fs afero.Fs) error
}
