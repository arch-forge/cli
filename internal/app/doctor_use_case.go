package app

import (
	"fmt"
	"path/filepath"

	"github.com/arch-forge/cli/internal/port"
	"github.com/arch-forge/cli/internal/domain"
)

// DoctorOptions carries all parameters for running an architecture compliance check.
type DoctorOptions struct {
	ProjectDir     string
	ScoreThreshold float64 // default 7.0
	Fix            bool    // when true, the CLI layer will generate and display fix suggestions
}

// DoctorUseCase orchestrates the architecture compliance analysis.
type DoctorUseCase struct {
	cfg      port.ConfigReader
	analyzer port.Analyzer
}

// NewDoctorUseCase constructs a DoctorUseCase.
func NewDoctorUseCase(cfg port.ConfigReader, analyzer port.Analyzer) *DoctorUseCase {
	return &DoctorUseCase{
		cfg:      cfg,
		analyzer: analyzer,
	}
}

// Execute runs the full doctor workflow:
//  1. Resolve the project directory (defaults to ".").
//  2. Read archforge.yaml for arch/variant/module config.
//  3. Analyze the project for architecture violations.
//  4. Return the report.
func (uc *DoctorUseCase) Execute(opts DoctorOptions) (domain.Report, error) {
	// Step 1 — Resolve absolute project directory.
	projectDir := opts.ProjectDir
	if projectDir == "" {
		projectDir = "."
	}
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return domain.Report{}, fmt.Errorf("doctor: resolve project dir: %w", err)
	}

	// Step 2 — Read archforge.yaml.
	cfgPath := filepath.Join(absDir, "archforge.yaml")
	cfg, err := uc.cfg.Read(cfgPath)
	if err != nil {
		return domain.Report{}, fmt.Errorf("doctor: read config: %w", err)
	}

	// Step 3 — Build analysis request and run analyzer.
	req := port.AnalysisRequest{
		ProjectDir: absDir,
		ModulePath: cfg.ModulePath,
		Arch:       cfg.Arch,
		Variant:    cfg.Variant,
	}

	report, err := uc.analyzer.Analyze(req)
	if err != nil {
		return domain.Report{}, fmt.Errorf("doctor: analyze: %w", err)
	}

	return report, nil
}
