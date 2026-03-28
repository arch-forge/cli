package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/arch-forge/cli/internal/app"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Stub implementations ---

type stubConfigReader struct {
	cfg *port.ProjectConfig
	err error
}

func (s *stubConfigReader) Read(_ string) (*port.ProjectConfig, error) {
	return s.cfg, s.err
}

func (s *stubConfigReader) Write(_ string, _ *port.ProjectConfig) error {
	return nil
}

type stubAnalyzer struct {
	report domain.Report
	err    error
}

func (s *stubAnalyzer) Analyze(_ port.AnalysisRequest) (domain.Report, error) {
	return s.report, s.err
}

// --- Tests ---

func TestDoctorUseCase_Execute_Success(t *testing.T) {
	cfg := &stubConfigReader{
		cfg: &port.ProjectConfig{
			Name:       "testapp",
			ModulePath: "github.com/acme/testapp",
			Arch:       domain.ArchHexagonal,
			Variant:    domain.VariantClassic,
		},
	}
	expectedReport := domain.Report{
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
		TotalRules: 10,
		Score:      10.0,
	}
	analyzer := &stubAnalyzer{report: expectedReport}

	uc := app.NewDoctorUseCase(cfg, analyzer)

	// Create a temp dir with a placeholder archforge.yaml so Abs works.
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "archforge.yaml"), []byte("name: testapp\n"), 0o644)

	report, err := uc.Execute(app.DoctorOptions{
		ProjectDir:     dir,
		ScoreThreshold: 7.0,
	})

	require.NoError(t, err)
	assert.Equal(t, expectedReport.Arch, report.Arch)
	assert.Equal(t, expectedReport.Variant, report.Variant)
	assert.InDelta(t, 10.0, report.Score, 0.001)
}

func TestDoctorUseCase_Execute_NoConfig(t *testing.T) {
	configErr := errors.New("archforge.yaml not found")
	cfg := &stubConfigReader{err: configErr}
	analyzer := &stubAnalyzer{}

	uc := app.NewDoctorUseCase(cfg, analyzer)

	dir := t.TempDir()
	_, err := uc.Execute(app.DoctorOptions{ProjectDir: dir})

	require.Error(t, err)
	assert.ErrorContains(t, err, "read config")
}

func TestDoctorUseCase_Execute_DefaultDir(t *testing.T) {
	cfg := &stubConfigReader{
		cfg: &port.ProjectConfig{
			Name:       "myapp",
			ModulePath: "github.com/acme/myapp",
			Arch:       domain.ArchHexagonal,
			Variant:    domain.VariantClassic,
		},
	}
	analyzer := &stubAnalyzer{
		report: domain.Report{
			Arch:    domain.ArchHexagonal,
			Variant: domain.VariantClassic,
			Score:   10.0,
		},
	}

	uc := app.NewDoctorUseCase(cfg, analyzer)

	// Empty ProjectDir should default to "." without error.
	report, err := uc.Execute(app.DoctorOptions{
		ProjectDir:     "",
		ScoreThreshold: 7.0,
	})

	require.NoError(t, err)
	assert.Equal(t, domain.ArchHexagonal, report.Arch)
}

func TestDoctorUseCase_Execute_AnalyzerError(t *testing.T) {
	cfg := &stubConfigReader{
		cfg: &port.ProjectConfig{
			Arch:    domain.ArchHexagonal,
			Variant: domain.VariantClassic,
		},
	}
	analyzerErr := errors.New("walk failed")
	analyzer := &stubAnalyzer{err: analyzerErr}

	uc := app.NewDoctorUseCase(cfg, analyzer)

	dir := t.TempDir()
	_, err := uc.Execute(app.DoctorOptions{ProjectDir: dir})

	require.Error(t, err)
	assert.ErrorContains(t, err, "analyze")
}
