package domain

import "fmt"

// ResolvedPaths holds the canonical directory paths for each layer of a project.
type ResolvedPaths struct {
	Domain     string
	Port       string
	App        string
	Adapter    string
	Handler    string
	Repository string
	Migration  string
	Test       string
}

// ResolvePaths calculates the canonical directory paths for the given architecture and variant.
//
// For VariantClassic, flat paths specific to each architecture are returned.
// For VariantModular, all paths are scoped under internal/{moduleName}/.
//
// Returns ErrArchNotSupported if arch is not recognized, or ErrVariantNotSupported
// if variant is not recognized.
func ResolvePaths(arch Architecture, variant Variant, moduleName string) (ResolvedPaths, error) {
	if !arch.IsValid() {
		return ResolvedPaths{}, fmt.Errorf("resolve paths: %w", ErrArchNotSupported)
	}
	if !variant.IsValid() {
		return ResolvedPaths{}, fmt.Errorf("resolve paths: %w", ErrVariantNotSupported)
	}

	if variant == VariantModular {
		// DDD modular uses a fixed internal/order/ structure, not a generic module name base.
		if arch == ArchDDD {
			return ResolvedPaths{
				Domain:     "internal/order/domain",
				Port:       "internal/order/domain",
				App:        "internal/order/application",
				Adapter:    "internal/order/infrastructure",
				Handler:    "internal/order/infrastructure/http",
				Repository: "internal/order/infrastructure/persistence/postgres",
				Migration:  "migrations",
				Test:       "internal/order/domain",
			}, nil
		}
		base := fmt.Sprintf("internal/%s", moduleName)
		return ResolvedPaths{
			Domain:     base + "/domain",
			Port:       base + "/ports",
			App:        base + "/application",
			Adapter:    base + "/adapters",
			Handler:    base + "/adapters/inbound/http",
			Repository: base + "/adapters/outbound/postgres",
			Migration:  base + "/migrations",
			Test:       base + "/domain",
		}, nil
	}

	// VariantClassic — arch-specific paths.
	var paths ResolvedPaths

	switch arch {
	case ArchHexagonal:
		paths = ResolvedPaths{
			Domain:     "internal/domain",
			Port:       "internal/ports",
			App:        "internal/application",
			Adapter:    "internal/adapters",
			Handler:    "internal/adapters/inbound/http",
			Repository: "internal/adapters/outbound/postgres",
		}
	case ArchClean:
		paths = ResolvedPaths{
			Domain:     "internal/domain",
			Port:       "internal/ports",
			App:        "internal/usecase",
			Adapter:    "internal/adapters",
			Handler:    "internal/adapters/http",
			Repository: "internal/adapters/postgres",
		}
	case ArchStandard:
		paths = ResolvedPaths{
			Domain:     "internal/model",
			Port:       "internal/service",
			App:        "internal/service",
			Adapter:    "internal/handler",
			Handler:    "internal/handler",
			Repository: "internal/repository",
		}
	case ArchDDD:
		paths = ResolvedPaths{
			Domain:     "internal/order/domain",
			Port:       "internal/order/domain",
			App:        "internal/order/application",
			Adapter:    "internal/order/infrastructure",
			Handler:    "internal/order/infrastructure/http",
			Repository: "internal/order/infrastructure/persistence/postgres",
		}
	case ArchCQRS:
		paths = ResolvedPaths{
			Domain:     "internal/domain",
			Port:       "internal/event",
			App:        "internal/command",
			Adapter:    "internal/query",
			Handler:    "internal/infrastructure/http",
			Repository: "internal/infrastructure/postgres",
		}
	case ArchModularMonolith:
		paths = ResolvedPaths{
			Domain:     "internal/module",
			Port:       "internal/module",
			App:        "internal/module",
			Adapter:    "internal/module",
			Handler:    "internal/platform",
			Repository: "internal/platform",
		}
	case ArchMicroservice:
		paths = ResolvedPaths{
			Domain:     "internal/domain",
			Port:       "internal/port",
			App:        "internal/app",
			Adapter:    "internal/adapter",
			Handler:    "internal/adapter/http",
			Repository: "internal/adapter/postgres",
		}
	}

	paths.Migration = "migrations"
	paths.Test = paths.Domain

	return paths, nil
}
