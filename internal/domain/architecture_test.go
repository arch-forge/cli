package domain_test

import (
	"testing"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestArchitecture_IsValid(t *testing.T) {
	cases := []struct {
		arch  domain.Architecture
		valid bool
	}{
		{domain.ArchHexagonal, true},
		{domain.ArchClean, true},
		{domain.ArchDDD, true},
		{domain.ArchStandard, true},
		{domain.ArchModularMonolith, true},
		{domain.ArchCQRS, true},
		{domain.ArchMicroservice, true},
		{"unknown", false},
		{"", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.arch), func(t *testing.T) {
			assert.Equal(t, tc.valid, tc.arch.IsValid())
		})
	}
}

func TestArchitecture_Info(t *testing.T) {
	t.Run("hexagonal returns correct DisplayName", func(t *testing.T) {
		info, ok := domain.ArchHexagonal.Info()
		assert.True(t, ok)
		assert.Equal(t, "Hexagonal", info.DisplayName)
		assert.NotEmpty(t, info.Description)
	})

	t.Run("unknown returns false", func(t *testing.T) {
		_, ok := domain.Architecture("unknown").Info()
		assert.False(t, ok)
	})
}

func TestAllArchitectures(t *testing.T) {
	all := domain.AllArchitectures()
	assert.Len(t, all, 7, "expected exactly 7 architectures")

	for _, info := range all {
		assert.NotEmpty(t, info.DisplayName, "DisplayName must not be empty for %s", info.Value)
		assert.NotEmpty(t, info.Description, "Description must not be empty for %s", info.Value)
	}
}
