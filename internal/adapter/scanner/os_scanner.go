// Package scanner provides a filesystem scanner adapter for arch_forge.
package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
)

// defaultSkipDirs is the set of directory names skipped when SkipDirs is empty.
var defaultSkipDirs = []string{".git", "vendor", ".idea", "node_modules", ".DS_Store"}

// OsScanner implements port.FileSystemScanner using the real OS filesystem.
type OsScanner struct{}

// NewOsScanner returns a new OsScanner.
func NewOsScanner() *OsScanner { return &OsScanner{} }

// Scan walks the filesystem starting at opts.RootDir and returns a FileNode tree.
func (s *OsScanner) Scan(opts port.ScanOptions) (domain.FileNode, error) {
	skipSet := opts.SkipDirs
	if len(skipSet) == 0 {
		skipSet = defaultSkipDirs
	}

	root := domain.FileNode{
		Name:  filepath.Base(opts.RootDir),
		IsDir: true,
	}

	children, err := scanDir(opts.RootDir, "", 0, opts.MaxDepth, opts, skipSet)
	if err != nil {
		return domain.FileNode{}, err
	}
	root.Children = children

	// Annotate the root node itself if the empty path maps to a layer.
	if layer, ok := opts.LayerMap[""]; ok {
		root.Layer = layer
	}

	return root, nil
}

// scanDir recursively reads a directory and returns its FileNode children.
func scanDir(absPath, relPath string, depth, maxDepth int, opts port.ScanOptions, skipSet []string) ([]domain.FileNode, error) {
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	// Separate directories and files, then sort each group alphabetically.
	var dirs, files []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e)
		} else {
			files = append(files, e)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	// Merge: directories first, then files.
	ordered := make([]os.DirEntry, 0, len(dirs)+len(files))
	ordered = append(ordered, dirs...)
	ordered = append(ordered, files...)

	var nodes []domain.FileNode
	for _, e := range ordered {
		name := e.Name()

		// Build relative path for this entry.
		var childRel string
		if relPath == "" {
			childRel = name
		} else {
			childRel = relPath + "/" + name
		}

		if e.IsDir() {
			// Skip unwanted directories by name.
			if shouldSkip(name, skipSet) {
				continue
			}

			node := domain.FileNode{
				Name:  name,
				IsDir: true,
			}

			// Annotate with layer.
			if layer, ok := opts.LayerMap[childRel]; ok {
				node.Layer = layer
			}

			// Annotate with module owner.
			node.ModuleOwner = resolveModuleOwner(childRel, opts.ModuleOwners)

			// Recurse only if we have not reached maxDepth (0 means unlimited).
			if maxDepth == 0 || depth+1 < maxDepth {
				childAbsPath := filepath.Join(absPath, name)
				children, err := scanDir(childAbsPath, childRel, depth+1, maxDepth, opts, skipSet)
				if err != nil {
					return nil, err
				}
				node.Children = children
			}

			nodes = append(nodes, node)
		} else {
			node := domain.FileNode{
				Name:  name,
				IsDir: false,
			}

			// Annotate file's module owner by its parent directory path.
			node.ModuleOwner = resolveModuleOwner(relPath, opts.ModuleOwners)

			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// shouldSkip reports whether a directory name is in the skip set.
func shouldSkip(name string, skipSet []string) bool {
	for _, s := range skipSet {
		if name == s {
			return true
		}
	}
	return false
}

// resolveModuleOwner returns the module name whose path prefix matches relPath.
func resolveModuleOwner(relPath string, owners map[string]string) string {
	best := ""
	bestOwner := ""
	for prefix, owner := range owners {
		if strings.HasPrefix(relPath, prefix) && len(prefix) > len(best) {
			best = prefix
			bestOwner = owner
		}
	}
	return bestOwner
}
