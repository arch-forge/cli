package domain_test

import (
	"testing"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestVariant_IsValid(t *testing.T) {
	cases := []struct {
		variant domain.Variant
		valid   bool
	}{
		{domain.VariantClassic, true},
		{domain.VariantModular, true},
		{"unknown", false},
		{"", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.variant), func(t *testing.T) {
			assert.Equal(t, tc.valid, tc.variant.IsValid())
		})
	}
}

func TestAllVariants(t *testing.T) {
	all := domain.AllVariants()
	assert.Len(t, all, 2, "expected exactly 2 variants")

	for _, info := range all {
		assert.NotEmpty(t, info.DisplayName, "DisplayName must not be empty for %s", info.Value)
		assert.NotEmpty(t, info.Description, "Description must not be empty for %s", info.Value)
	}
}
