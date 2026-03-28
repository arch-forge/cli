package domain

// ArchLayer represents a named architectural layer.
type ArchLayer string

const (
	LayerDomain     ArchLayer = "domain"
	LayerPort       ArchLayer = "port"
	LayerApp        ArchLayer = "app"
	LayerAdapter    ArchLayer = "adapter"
	LayerEntities   ArchLayer = "entities"
	LayerUseCases   ArchLayer = "use cases"
	LayerController ArchLayer = "controller"
	LayerGateway    ArchLayer = "gateway"
	LayerInternal   ArchLayer = "internal"
	LayerPublic     ArchLayer = "public"
	LayerEntrypoint ArchLayer = "entrypoint"
	LayerUnknown    ArchLayer = ""
)

// FileNode represents a single entry in the project file tree.
type FileNode struct {
	Name        string     `json:"name"`
	IsDir       bool       `json:"is_dir"`
	Layer       ArchLayer  `json:"layer,omitempty"`
	ModuleOwner string     `json:"module_owner,omitempty"`
	Children    []FileNode `json:"children,omitempty"`
}

// ProjectStats holds aggregate counts.
type ProjectStats struct {
	TotalFiles       int `json:"total_files"`
	TotalDirectories int `json:"total_directories"`
	ModuleCount      int `json:"module_count"`
}

// ProjectSummary is the complete read-only snapshot of an inspected project.
type ProjectSummary struct {
	Name             string       `json:"name"`
	ModulePath       string       `json:"module_path"`
	Arch             Architecture `json:"arch"`
	Variant          Variant      `json:"variant"`
	GoVersion        string       `json:"go_version"`
	InstalledModules []string     `json:"installed_modules"`
	Tree             FileNode     `json:"tree"`
	Stats            ProjectStats `json:"stats"`
}
