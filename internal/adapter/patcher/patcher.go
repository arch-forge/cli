package patcher

import (
	"context"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/arch-forge/cli/internal/port"
	"github.com/spf13/afero"
)

// FilePatcher implements port.Patcher.
type FilePatcher struct{}

// NewFilePatcher constructs a ready-to-use FilePatcher.
func NewFilePatcher() *FilePatcher {
	return &FilePatcher{}
}

// Apply implements port.Patcher.
// For each PatchRequest, it finds all matching files and applies the patch to each.
func (p *FilePatcher) Apply(
	ctx context.Context,
	rootDir string,
	patches []port.PatchRequest,
	fs afero.Fs,
) error {
	for _, req := range patches {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		matched, err := matchFiles(fs, rootDir, req.TargetGlob)
		if err != nil {
			return fmt.Errorf("match files for glob %q: %w", req.TargetGlob, err)
		}

		if len(matched) == 0 {
			if req.Optional {
				continue
			}
			return fmt.Errorf("no files matched glob %q", req.TargetGlob)
		}

		for _, filePath := range matched {
			content, err := afero.ReadFile(fs, filePath)
			if err != nil {
				return fmt.Errorf("read file %s: %w", filePath, err)
			}

			patched, err := applyPatch(content, req)
			if err != nil {
				return fmt.Errorf("apply patch to %s: %w", filePath, err)
			}

			formatted, err := formatGoSource(filePath, patched)
			if err != nil {
				// formatGoSource returns unformatted content on failure; treat as warning only.
				formatted = patched
			}

			if err := afero.WriteFile(fs, filePath, formatted, 0o644); err != nil {
				return fmt.Errorf("write file %s: %w", filePath, err)
			}
		}
	}
	return nil
}

// matchFiles walks the filesystem rooted at rootDir and returns all paths
// matching the glob pattern (relative to rootDir).
// Uses afero.Walk + filepath.Match.
func matchFiles(fs afero.Fs, rootDir, glob string) ([]string, error) {
	var matched []string

	err := afero.Walk(fs, rootDir, func(fullPath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, fullPath)
		if err != nil {
			return fmt.Errorf("rel path for %s: %w", fullPath, err)
		}

		ok, err := filepath.Match(glob, relPath)
		if err != nil {
			return fmt.Errorf("match glob %q against %q: %w", glob, relPath, err)
		}
		if ok {
			matched = append(matched, fullPath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return matched, nil
}

// applyPatch applies a single PatchRequest to file content.
// The anchor tag format is: "// arch_forge:<anchor>"
// Supported actions:
//
//	"inject_after"  — insert content on the line after the anchor line
//	"inject_before" — insert content on the line before the anchor line
//	"replace"       — replace the anchor line with content
//
// Returns the modified bytes, or an error if anchor not found (unless req.Optional).
func applyPatch(content []byte, req port.PatchRequest) ([]byte, error) {
	lines := strings.Split(string(content), "\n")
	anchorTag := "// arch_forge:" + req.Anchor

	anchorIndex := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == anchorTag {
			anchorIndex = i
			break
		}
	}

	if anchorIndex == -1 {
		if req.Optional {
			return content, nil
		}
		return nil, fmt.Errorf("anchor %q not found in file", req.Anchor)
	}

	var result []string
	switch req.Action {
	case "inject_after":
		result = make([]string, 0, len(lines)+1)
		result = append(result, lines[:anchorIndex+1]...)
		result = append(result, req.Content)
		result = append(result, lines[anchorIndex+1:]...)
	case "inject_before":
		result = make([]string, 0, len(lines)+1)
		result = append(result, lines[:anchorIndex]...)
		result = append(result, req.Content)
		result = append(result, lines[anchorIndex:]...)
	case "replace":
		result = make([]string, len(lines))
		copy(result, lines)
		result[anchorIndex] = req.Content
	default:
		return nil, fmt.Errorf("unknown patch action %q", req.Action)
	}

	return []byte(strings.Join(result, "\n")), nil
}

// formatGoSource runs go/format on .go file content.
// Non-Go content (path doesn't end in ".go") is returned unchanged.
// If formatting fails (e.g. patched file has syntax error), returns the
// unformatted content — formatting is best-effort and does not fail hard.
func formatGoSource(filePath string, content []byte) ([]byte, error) {
	if !strings.HasSuffix(filePath, ".go") {
		return content, nil
	}

	formatted, err := format.Source(content)
	if err != nil {
		// Return unformatted content and a descriptive warning.
		return content, fmt.Errorf("go/format warning for %s: %w", filePath, err)
	}
	return formatted, nil
}

