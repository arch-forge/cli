package analyzer

import (
	"testing"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRulesForArch_HexagonalClassic(t *testing.T) {
	rules, err := rulesForArch(domain.ArchHexagonal, domain.VariantClassic, "")
	require.NoError(t, err)

	// Expect rules for domain, port, app, and adapter lateral.
	assert.NotEmpty(t, rules)

	// All directional rules for hexagonal should be errors.
	ruleMap := make(map[string]LayerRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	domainNoPort, ok := ruleMap["domain-no-port"]
	require.True(t, ok, "expected domain-no-port rule")
	assert.Equal(t, "internal/domain", domainNoPort.FromLayer)
	assert.Equal(t, "internal/ports", domainNoPort.ToLayer)
	assert.Equal(t, domain.SeverityError, domainNoPort.Severity)

	domainNoApp, ok := ruleMap["domain-no-app"]
	require.True(t, ok, "expected domain-no-app rule")
	assert.Equal(t, domain.SeverityError, domainNoApp.Severity)

	domainNoAdapter, ok := ruleMap["domain-no-adapter"]
	require.True(t, ok, "expected domain-no-adapter rule")
	assert.Equal(t, domain.SeverityError, domainNoAdapter.Severity)

	portNoApp, ok := ruleMap["port-no-app"]
	require.True(t, ok, "expected port-no-app rule")
	assert.Equal(t, domain.SeverityError, portNoApp.Severity)

	portNoAdapter, ok := ruleMap["port-no-adapter"]
	require.True(t, ok, "expected port-no-adapter rule")
	assert.Equal(t, domain.SeverityError, portNoAdapter.Severity)

	appNoAdapter, ok := ruleMap["app-no-adapter"]
	require.True(t, ok, "expected app-no-adapter rule")
	assert.Equal(t, domain.SeverityError, appNoAdapter.Severity)

	// Lateral adapter rule should always be warning.
	lateralRule, ok := ruleMap["adapter-no-lateral"]
	require.True(t, ok, "expected adapter-no-lateral rule")
	assert.Equal(t, domain.SeverityWarning, lateralRule.Severity)
}

func TestRulesForArch_CleanClassic(t *testing.T) {
	rules, err := rulesForArch(domain.ArchClean, domain.VariantClassic, "")
	require.NoError(t, err)
	assert.NotEmpty(t, rules)

	ruleMap := make(map[string]LayerRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// Clean arch uses different layer paths.
	domainNoPort, ok := ruleMap["domain-no-port"]
	require.True(t, ok, "expected domain-no-port rule")
	assert.Equal(t, "internal/domain", domainNoPort.FromLayer)
	assert.Equal(t, "internal/ports", domainNoPort.ToLayer)
	assert.Equal(t, domain.SeverityError, domainNoPort.Severity)
}

func TestRulesForArch_ModularMonolith_AllWarnings(t *testing.T) {
	rules, err := rulesForArch(domain.ArchModularMonolith, domain.VariantClassic, "")
	require.NoError(t, err)
	assert.NotEmpty(t, rules)

	for _, r := range rules {
		// ModularMonolith must have warnings even for directional rules.
		assert.Equal(t, domain.SeverityWarning, r.Severity,
			"rule %q in ModularMonolith should be a warning", r.Name)
	}
}

func TestRulesForArch_Standard_AllWarnings(t *testing.T) {
	rules, err := rulesForArch(domain.ArchStandard, domain.VariantClassic, "")
	require.NoError(t, err)
	assert.NotEmpty(t, rules)

	for _, r := range rules {
		assert.Equal(t, domain.SeverityWarning, r.Severity,
			"rule %q in Standard layout should be a warning", r.Name)
	}
}

func TestRulesForArch_InvalidArch(t *testing.T) {
	_, err := rulesForArch("nonexistent", domain.VariantClassic, "")
	assert.Error(t, err)
}

func TestRulesForArch_InvalidVariant(t *testing.T) {
	_, err := rulesForArch(domain.ArchHexagonal, "nonexistent", "")
	assert.Error(t, err)
}

func TestRulesForArch_DDD(t *testing.T) {
	rules, err := rulesForArch(domain.ArchDDD, domain.VariantClassic, "")
	require.NoError(t, err)
	assert.NotEmpty(t, rules)

	ruleMap := make(map[string]LayerRule)
	for _, r := range rules {
		ruleMap[r.Name] = r
	}

	// In the DDD architecture, domain and port share the same package (internal/order/domain),
	// so no domain-no-port rule is generated (it would be a self-rule).
	_, hasDomainNoPort := ruleMap["domain-no-port"]
	assert.False(t, hasDomainNoPort, "domain-no-port rule must not exist when Domain == Port")

	// The domain-to-app rule should still be present.
	domainNoApp, ok := ruleMap["domain-no-app"]
	require.True(t, ok, "expected domain-no-app rule")
	assert.Equal(t, "internal/order/domain", domainNoApp.FromLayer)
	assert.Equal(t, "internal/order/application", domainNoApp.ToLayer)
	assert.Equal(t, domain.SeverityError, domainNoApp.Severity)
}
