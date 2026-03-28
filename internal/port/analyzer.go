package port

import "github.com/arch-forge/cli/internal/domain"

// AnalysisRequest bundles the inputs for an architecture compliance analysis.
type AnalysisRequest struct {
	ProjectDir string
	ModulePath string
	Arch       domain.Architecture
	Variant    domain.Variant
}

// Analyzer checks a Go project for architecture compliance violations.
type Analyzer interface {
	Analyze(req AnalysisRequest) (domain.Report, error)
}
