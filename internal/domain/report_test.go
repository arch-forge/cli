package domain_test

import (
	"testing"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestComputeScore(t *testing.T) {
	tests := []struct {
		name       string
		violations []domain.Violation
		totalRules int
		wantScore  float64
	}{
		{
			name:       "no rules yields perfect score",
			violations: nil,
			totalRules: 0,
			wantScore:  10.0,
		},
		{
			name:       "no violations yields perfect score",
			violations: []domain.Violation{},
			totalRules: 10,
			wantScore:  10.0,
		},
		{
			name: "all errors yields zero score",
			violations: []domain.Violation{
				{Severity: domain.SeverityError},
				{Severity: domain.SeverityError},
			},
			totalRules: 2,
			wantScore:  0.0,
		},
		{
			name: "half errors yields 5.0 score",
			violations: []domain.Violation{
				{Severity: domain.SeverityError},
			},
			totalRules: 2,
			wantScore:  5.0,
		},
		{
			name: "only warnings do not reduce score",
			violations: []domain.Violation{
				{Severity: domain.SeverityWarning},
				{Severity: domain.SeverityWarning},
			},
			totalRules: 4,
			wantScore:  10.0,
		},
		{
			name: "mixed errors and warnings",
			violations: []domain.Violation{
				{Severity: domain.SeverityError},
				{Severity: domain.SeverityWarning},
			},
			totalRules: 4,
			wantScore:  7.5, // (4-1)/4*10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &domain.Report{
				Violations: tt.violations,
				TotalRules: tt.totalRules,
			}
			r.ComputeScore()
			assert.InDelta(t, tt.wantScore, r.Score, 0.001)
		})
	}
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name       string
		violations []domain.Violation
		want       bool
	}{
		{
			name:       "no violations returns false",
			violations: nil,
			want:       false,
		},
		{
			name: "only warnings returns false",
			violations: []domain.Violation{
				{Severity: domain.SeverityWarning},
			},
			want: false,
		},
		{
			name: "one error returns true",
			violations: []domain.Violation{
				{Severity: domain.SeverityError},
			},
			want: true,
		},
		{
			name: "mixed returns true",
			violations: []domain.Violation{
				{Severity: domain.SeverityWarning},
				{Severity: domain.SeverityError},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := domain.Report{Violations: tt.violations}
			assert.Equal(t, tt.want, r.HasErrors())
		})
	}
}
