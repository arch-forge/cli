package app

import (
	"fmt"
	"path/filepath"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
)

// InspectOptions carries parameters for the inspect workflow.
type InspectOptions struct {
	ProjectDir string
	MaxDepth   int // 0 = default (3)
}

// InspectUseCase reads archforge.yaml and the filesystem, and returns a ProjectSummary.
type InspectUseCase struct {
	cfg     port.ConfigReader
	scanner port.FileSystemScanner
}

// NewInspectUseCase constructs an InspectUseCase.
func NewInspectUseCase(cfg port.ConfigReader, scanner port.FileSystemScanner) *InspectUseCase {
	return &InspectUseCase{cfg: cfg, scanner: scanner}
}

// Execute runs the full inspect workflow:
//  1. Resolve the absolute project directory.
//  2. Read archforge.yaml.
//  3. Build layer and module-owner maps.
//  4. Scan the filesystem.
//  5. Count nodes and return the ProjectSummary.
func (uc *InspectUseCase) Execute(opts InspectOptions) (domain.ProjectSummary, error) {
	projectDir := opts.ProjectDir
	if projectDir == "" {
		projectDir = "."
	}

	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return domain.ProjectSummary{}, fmt.Errorf("inspect: resolve project dir: %w", err)
	}

	cfg, err := uc.cfg.Read(filepath.Join(absDir, "archforge.yaml"))
	if err != nil {
		return domain.ProjectSummary{}, fmt.Errorf("inspect: read config: %w", err)
	}

	layerMap := buildLayerMap(cfg.Arch, cfg.Variant)
	moduleOwners := buildModuleOwners(cfg.Arch, cfg.Variant, cfg.InstalledModules)

	maxDepth := opts.MaxDepth
	if maxDepth == 0 {
		maxDepth = 3
	}

	tree, err := uc.scanner.Scan(port.ScanOptions{
		RootDir:      absDir,
		MaxDepth:     maxDepth,
		LayerMap:     layerMap,
		ModuleOwners: moduleOwners,
	})
	if err != nil {
		return domain.ProjectSummary{}, fmt.Errorf("inspect: scan filesystem: %w", err)
	}

	stats := countNodes(tree)
	stats.ModuleCount = len(cfg.InstalledModules)

	return domain.ProjectSummary{
		Name:             cfg.Name,
		ModulePath:       cfg.ModulePath,
		Arch:             cfg.Arch,
		Variant:          cfg.Variant,
		GoVersion:        cfg.GoVersion,
		InstalledModules: cfg.InstalledModules,
		Tree:             tree,
		Stats:            stats,
	}, nil
}

// buildLayerMap returns a map of relative directory path → ArchLayer for the given arch/variant.
func buildLayerMap(arch domain.Architecture, variant domain.Variant) map[string]domain.ArchLayer {
	paths, err := domain.ResolvePaths(arch, variant, "")
	if err != nil {
		return map[string]domain.ArchLayer{}
	}

	m := make(map[string]domain.ArchLayer)

	switch arch {
	case domain.ArchHexagonal, domain.ArchMicroservice:
		setLayer(m, paths.Domain, domain.LayerDomain)
		setLayer(m, paths.Port, domain.LayerPort)
		setLayer(m, paths.App, domain.LayerApp)
		setLayer(m, paths.Adapter, domain.LayerAdapter)

	case domain.ArchClean:
		setLayer(m, paths.Domain, domain.LayerEntities)
		setLayer(m, paths.Port, domain.LayerPort)
		setLayer(m, paths.App, domain.LayerUseCases)
		setLayer(m, paths.Adapter, domain.LayerController)

	case domain.ArchStandard:
		setLayer(m, "cmd", domain.LayerEntrypoint)
		setLayer(m, "internal", domain.LayerInternal)
		setLayer(m, "pkg", domain.LayerPublic)

	case domain.ArchDDD:
		setLayer(m, paths.Domain, domain.LayerDomain)
		setLayer(m, paths.Port, domain.LayerPort)
		setLayer(m, paths.App, domain.LayerApp)
		setLayer(m, paths.Adapter, domain.LayerAdapter)

	case domain.ArchCQRS:
		setLayer(m, paths.Domain, domain.LayerDomain)
		setLayer(m, paths.Port, domain.LayerPort)
		setLayer(m, paths.App, domain.LayerApp)
		setLayer(m, paths.Adapter, domain.LayerAdapter)

	case domain.ArchModularMonolith:
		setLayer(m, "internal", domain.LayerInternal)
	}

	return m
}

// setLayer adds path → layer to m, skipping empty paths.
func setLayer(m map[string]domain.ArchLayer, path string, layer domain.ArchLayer) {
	if path != "" {
		m[path] = layer
	}
}

// buildModuleOwners returns a map of directory path prefix → module name.
func buildModuleOwners(arch domain.Architecture, variant domain.Variant, modules []string) map[string]string {
	owners := make(map[string]string, len(modules))
	for _, mod := range modules {
		paths, err := domain.ResolvePaths(arch, variant, mod)
		if err != nil {
			continue
		}
		if paths.Domain != "" {
			owners[paths.Domain] = mod
		}
	}
	return owners
}

// countNodes walks the FileNode tree and returns aggregate statistics.
// The module count is set by the caller because it comes from config, not the tree.
func countNodes(node domain.FileNode) domain.ProjectStats {
	stats := domain.ProjectStats{}
	walkNode(node, &stats)
	return stats
}

func walkNode(node domain.FileNode, stats *domain.ProjectStats) {
	if node.IsDir {
		stats.TotalDirectories++
	} else {
		stats.TotalFiles++
	}
	for _, child := range node.Children {
		walkNode(child, stats)
	}
}
