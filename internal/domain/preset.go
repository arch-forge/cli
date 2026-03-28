package domain

// Preset represents a named combination of architecture + variant + modules.
type Preset struct {
	Name        string
	DisplayName string
	Description string
	Arch        Architecture
	Variant     Variant
	Modules     []string
}

// presets is the registry of all built-in presets.
var presets = []Preset{
	{
		Name:        "starter",
		DisplayName: "Starter",
		Description: "Standard Layout with API server, structured logging, Docker, and Makefile",
		Arch:        ArchStandard,
		Variant:     VariantModular,
		Modules:     []string{"api", "logging", "docker", "makefile"},
	},
	{
		Name:        "production-api",
		DisplayName: "Production API",
		Description: "Production-ready Hexagonal API with auth, database, logging, Docker, and Makefile",
		Arch:        ArchHexagonal,
		Variant:     VariantClassic,
		Modules:     []string{"api", "database", "auth", "logging", "docker", "makefile"},
	},
	{
		Name:        "microservice",
		DisplayName: "Microservice",
		Description: "Single microservice with gRPC, database, structured logging, Docker, and Makefile",
		Arch:        ArchMicroservice,
		Variant:     VariantClassic,
		Modules:     []string{"grpc", "database", "logging", "docker", "makefile"},
	},
}

// AllPresets returns a copy of the ordered slice of all built-in presets.
func AllPresets() []Preset {
	result := make([]Preset, len(presets))
	copy(result, presets)
	return result
}

// FindPreset looks up a preset by name. Returns the preset and true if found,
// or a zero-value Preset and false if no preset with the given name exists.
func FindPreset(name string) (Preset, bool) {
	for _, p := range presets {
		if p.Name == name {
			return p, true
		}
	}
	return Preset{}, false
}
