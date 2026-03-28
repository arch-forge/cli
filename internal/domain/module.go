package domain

// ModuleDep describes a dependency on another module.
type ModuleDep struct {
	Name     string
	Required bool
}

// GoDep describes a Go package dependency required by a module.
type GoDep struct {
	Package string
	Version string
}

// ModuleOption describes a configurable option exposed by a module.
type ModuleOption struct {
	Name        string
	Type        string
	Values      []string
	Default     any
	Description string
	Condition   string
}

// Patch describes a file modification to be applied when a module is installed.
type Patch struct {
	TargetGlob   string
	Action       string
	Anchor       string
	TemplatePath string
	Optional     bool
}

// Module represents an installable arch_forge module with its metadata and constraints.
type Module struct {
	Name        string
	Version     string
	Description string
	Category    ModuleCategory

	SupportedArchitectures []Architecture
	SupportedVariants      []Variant
	Dependencies           []ModuleDep
	GoDeps                 []GoDep
	Options                []ModuleOption
	Patches                []Patch
}

// SupportsArch reports whether the module supports the given architecture.
// An empty SupportedArchitectures list means all architectures are supported.
func (m Module) SupportsArch(arch Architecture) bool {
	if len(m.SupportedArchitectures) == 0 {
		return true
	}
	for _, a := range m.SupportedArchitectures {
		if a == arch {
			return true
		}
	}
	return false
}

// SupportsVariant reports whether the module supports the given variant.
// An empty SupportedVariants list means all variants are supported.
func (m Module) SupportsVariant(v Variant) bool {
	if len(m.SupportedVariants) == 0 {
		return true
	}
	for _, sv := range m.SupportedVariants {
		if sv == v {
			return true
		}
	}
	return false
}

// RequiredDeps returns only the dependencies that are marked as required.
func (m Module) RequiredDeps() []ModuleDep {
	var required []ModuleDep
	for _, dep := range m.Dependencies {
		if dep.Required {
			required = append(required, dep)
		}
	}
	return required
}
