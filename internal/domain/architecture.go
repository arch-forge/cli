package domain

// Architecture represents a supported project architecture pattern.
type Architecture string

const (
	ArchHexagonal      Architecture = "hexagonal"
	ArchClean          Architecture = "clean"
	ArchDDD            Architecture = "ddd"
	ArchStandard       Architecture = "standard"
	ArchModularMonolith Architecture = "modular_monolith"
	ArchCQRS           Architecture = "cqrs"
	ArchMicroservice   Architecture = "microservice"
)

// ArchInfo holds human-readable metadata for an architecture.
type ArchInfo struct {
	Value       Architecture
	DisplayName string
	Description string
}

// archInfos is the ordered list of all supported architectures.
var archInfos = []ArchInfo{
	{
		Value:       ArchHexagonal,
		DisplayName: "Hexagonal",
		Description: "Ports & Adapters — isolated domain with interchangeable adapters",
	},
	{
		Value:       ArchClean,
		DisplayName: "Clean Architecture",
		Description: "Concentric layers — entities, use cases, interface adapters, frameworks",
	},
	{
		Value:       ArchDDD,
		DisplayName: "Domain-Driven Design",
		Description: "Domain-Driven Design — bounded contexts, aggregates, domain events",
	},
	{
		Value:       ArchStandard,
		DisplayName: "Standard Layout",
		Description: "Go Standard Layout — community-standard cmd/internal/pkg structure",
	},
	{
		Value:       ArchModularMonolith,
		DisplayName: "Modular Monolith",
		Description: "Modular Monolith — explicit module boundaries prepared for extraction",
	},
	{
		Value:       ArchCQRS,
		DisplayName: "CQRS + Event Sourcing",
		Description: "CQRS + Event Sourcing — separate command/query sides with event store",
	},
	{
		Value:       ArchMicroservice,
		DisplayName: "Microservice",
		Description: "Single Microservice — production-ready service with observability",
	},
}

// IsValid reports whether a is a recognized architecture.
func (a Architecture) IsValid() bool {
	for _, info := range archInfos {
		if info.Value == a {
			return true
		}
	}
	return false
}

// Info returns the ArchInfo for a, and whether it was found.
func (a Architecture) Info() (ArchInfo, bool) {
	for _, info := range archInfos {
		if info.Value == a {
			return info, true
		}
	}
	return ArchInfo{}, false
}

// AllArchitectures returns a copy of the ordered slice of all architectures.
func AllArchitectures() []ArchInfo {
	result := make([]ArchInfo, len(archInfos))
	copy(result, archInfos)
	return result
}
