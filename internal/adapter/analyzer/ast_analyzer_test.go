package analyzer

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testdataDir returns the absolute path to a testdata subdirectory.
func testdataDir(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func TestAnalyze_HexagonalViolation(t *testing.T) {
	a := NewASTAnalyzer()
	req := port.AnalysisRequest{
		ProjectDir: testdataDir("hexagonal_violation"),
		ModulePath: "example.com/testapp",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
	}

	report, err := a.Analyze(req)
	require.NoError(t, err)

	// Expect at least one error violation (domain importing port).
	assert.True(t, report.HasErrors(), "expected at least one error violation")

	// Find the domain-no-port violation.
	found := false
	for _, v := range report.Violations {
		if v.Rule == "domain-no-port" && v.Severity == domain.SeverityError {
			found = true
			break
		}
	}
	assert.True(t, found, "expected a domain-no-port error violation")

	// Score should be below 10.
	assert.Less(t, report.Score, 10.0)
}

func TestAnalyze_HexagonalClean(t *testing.T) {
	a := NewASTAnalyzer()
	req := port.AnalysisRequest{
		ProjectDir: testdataDir("hexagonal_clean"),
		ModulePath: "example.com/cleanapp",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
	}

	report, err := a.Analyze(req)
	require.NoError(t, err)

	// Clean project: no error violations expected.
	assert.False(t, report.HasErrors(), "expected no error violations in a clean project")

	// Score should be perfect.
	assert.InDelta(t, 10.0, report.Score, 0.001)
}

func TestAnalyze_InvalidDir(t *testing.T) {
	a := NewASTAnalyzer()
	req := port.AnalysisRequest{
		ProjectDir: "/nonexistent/path/that/does/not/exist",
		ModulePath: "example.com/nope",
		Arch:       domain.ArchHexagonal,
		Variant:    domain.VariantClassic,
	}

	_, err := a.Analyze(req)
	assert.Error(t, err)
}
