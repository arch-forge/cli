package port

import "github.com/arch-forge/cli/internal/domain"

// ProjectConfig is the serialisable representation of archforge.yaml.
type ProjectConfig struct {
	Name             string              `yaml:"name"`
	ModulePath       string              `yaml:"module_path"`
	GoVersion        string              `yaml:"go_version"`
	Arch             domain.Architecture `yaml:"arch"`
	Variant          domain.Variant      `yaml:"variant"`
	InstalledModules []string            `yaml:"modules"`
}

// ConfigReader reads and writes the archforge.yaml project configuration.
type ConfigReader interface {
	// Read parses archforge.yaml at the given file path.
	Read(path string) (*ProjectConfig, error)

	// Write serialises cfg and writes it to path, creating the file
	// if it does not exist.
	Write(path string, cfg *ProjectConfig) error
}
