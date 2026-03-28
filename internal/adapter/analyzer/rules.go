package analyzer

import (
	"fmt"

	"github.com/arch-forge/cli/internal/domain"
)

// LayerRule describes a directional import prohibition.
type LayerRule struct {
	Name      string
	FromLayer string // package path segment (e.g., "internal/domain")
	ToLayer   string // forbidden import segment
	Severity  domain.Severity
}

// rulesForArch returns the import direction rules for the given arch/variant/modulePath.
// Uses domain.ResolvePaths to derive the layer paths.
func rulesForArch(arch domain.Architecture, variant domain.Variant, modulePath string) ([]LayerRule, error) {
	paths, err := domain.ResolvePaths(arch, variant, modulePath)
	if err != nil {
		return nil, fmt.Errorf("rulesForArch: %w", err)
	}

	// For ModularMonolith and Standard layout, all violations are warnings.
	errorSeverity := domain.SeverityError
	if arch == domain.ArchModularMonolith || arch == domain.ArchStandard {
		errorSeverity = domain.SeverityWarning
	}

	// Skip rules when domain == port (avoids pointless self-rules for Standard/ModularMonolith).
	var rules []LayerRule

	// Domain must NOT import Port, App, or Adapter layers.
	if paths.Domain != paths.Port {
		rules = append(rules, LayerRule{
			Name:      "domain-no-port",
			FromLayer: paths.Domain,
			ToLayer:   paths.Port,
			Severity:  errorSeverity,
		})
	}
	if paths.Domain != paths.App {
		rules = append(rules, LayerRule{
			Name:      "domain-no-app",
			FromLayer: paths.Domain,
			ToLayer:   paths.App,
			Severity:  errorSeverity,
		})
	}
	if paths.Domain != paths.Adapter {
		rules = append(rules, LayerRule{
			Name:      "domain-no-adapter",
			FromLayer: paths.Domain,
			ToLayer:   paths.Adapter,
			Severity:  errorSeverity,
		})
	}

	// Port must NOT import App or Adapter layers.
	if paths.Port != paths.App {
		rules = append(rules, LayerRule{
			Name:      "port-no-app",
			FromLayer: paths.Port,
			ToLayer:   paths.App,
			Severity:  errorSeverity,
		})
	}
	if paths.Port != paths.Adapter {
		rules = append(rules, LayerRule{
			Name:      "port-no-adapter",
			FromLayer: paths.Port,
			ToLayer:   paths.Adapter,
			Severity:  errorSeverity,
		})
	}

	// App must NOT import Adapter layer.
	if paths.App != paths.Adapter {
		rules = append(rules, LayerRule{
			Name:      "app-no-adapter",
			FromLayer: paths.App,
			ToLayer:   paths.Adapter,
			Severity:  errorSeverity,
		})
	}

	// Adapter packages must NOT cross-import other adapter subtrees (lateral dependency).
	// This rule is always a warning regardless of architecture.
	rules = append(rules, LayerRule{
		Name:      "adapter-no-lateral",
		FromLayer: paths.Adapter,
		ToLayer:   paths.Adapter, // special: lateral check handled in the analyzer
		Severity:  domain.SeverityWarning,
	})

	return rules, nil
}
