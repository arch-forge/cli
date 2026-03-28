package port

import "github.com/arch-forge/cli/internal/domain"

// ScanOptions controls what the FileSystemScanner traverses.
type ScanOptions struct {
	RootDir      string
	MaxDepth     int
	LayerMap     map[string]domain.ArchLayer // relative dir path → layer label
	ModuleOwners map[string]string           // path prefix → module name
	SkipDirs     []string                    // dir names to skip (e.g., ".git", "vendor")
}

// FileSystemScanner walks a real filesystem and returns a FileNode tree.
type FileSystemScanner interface {
	Scan(opts ScanOptions) (domain.FileNode, error)
}
