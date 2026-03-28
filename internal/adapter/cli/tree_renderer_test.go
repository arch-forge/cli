package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arch-forge/cli/internal/adapter/cli"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleSummary returns a minimal ProjectSummary for use in renderer tests.
func sampleSummary() domain.ProjectSummary {
	return domain.ProjectSummary{
		Name:             "testapp",
		ModulePath:       "github.com/acme/testapp",
		Arch:             domain.ArchHexagonal,
		Variant:          domain.VariantClassic,
		GoVersion:        "1.23",
		InstalledModules: []string{"api", "database"},
		Tree: domain.FileNode{
			Name:  "testapp",
			IsDir: true,
			Children: []domain.FileNode{
				{
					Name:  "internal",
					IsDir: true,
					Layer: domain.LayerInternal,
					Children: []domain.FileNode{
						{Name: "domain", IsDir: true, Layer: domain.LayerDomain},
						{Name: "port", IsDir: true, Layer: domain.LayerPort},
					},
				},
				{Name: "go.mod", IsDir: false},
			},
		},
		Stats: domain.ProjectStats{
			TotalFiles:       1,
			TotalDirectories: 3,
			ModuleCount:      2,
		},
	}
}

func TestRenderTree_NoColor(t *testing.T) {
	var buf bytes.Buffer
	err := cli.RenderTree(&buf, sampleSummary(), false)
	require.NoError(t, err)

	out := buf.String()

	// Header fields.
	assert.Contains(t, out, "testapp")
	assert.Contains(t, out, "github.com/acme/testapp")
	assert.Contains(t, out, "hexagonal")
	assert.Contains(t, out, "classic")
	assert.Contains(t, out, "1.23")
	assert.Contains(t, out, "api, database")

	// Tree structure uses box-drawing characters.
	assert.Contains(t, out, "├──")
	assert.Contains(t, out, "└──")

	// Footer stats.
	assert.Contains(t, out, "1 files")
	assert.Contains(t, out, "3 directories")
	assert.Contains(t, out, "2 modules")
}

func TestRenderTree_LayerAnnotations(t *testing.T) {
	var buf bytes.Buffer
	err := cli.RenderTree(&buf, sampleSummary(), false)
	require.NoError(t, err)

	out := buf.String()

	// Layer annotations appear as "[domain]", "[port]", etc.
	assert.Contains(t, out, "[domain]")
	assert.Contains(t, out, "[port]")
	assert.Contains(t, out, "[internal]")
}

func TestRenderTree_ModulesNone(t *testing.T) {
	s := sampleSummary()
	s.InstalledModules = nil

	var buf bytes.Buffer
	err := cli.RenderTree(&buf, s, false)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "none")
}

func TestRenderTree_ModuleOwner(t *testing.T) {
	s := sampleSummary()
	s.Tree.Children[0].ModuleOwner = "mymodule"

	var buf bytes.Buffer
	err := cli.RenderTree(&buf, s, false)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "{mod:mymodule}")
}

func TestRenderJSON_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	err := cli.RenderJSON(&buf, sampleSummary())
	require.NoError(t, err)

	// Output must be valid JSON.
	var raw map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &raw))
}

func TestRenderJSON_RoundTrip(t *testing.T) {
	original := sampleSummary()

	var buf bytes.Buffer
	err := cli.RenderJSON(&buf, original)
	require.NoError(t, err)

	var decoded domain.ProjectSummary
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))

	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, original.ModulePath, decoded.ModulePath)
	assert.Equal(t, original.Arch, decoded.Arch)
	assert.Equal(t, original.Variant, decoded.Variant)
	assert.Equal(t, original.GoVersion, decoded.GoVersion)
	assert.Equal(t, original.InstalledModules, decoded.InstalledModules)
	assert.Equal(t, original.Stats, decoded.Stats)
}

func TestRenderJSON_OutputEndsWithNewline(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, cli.RenderJSON(&buf, sampleSummary()))
	assert.True(t, strings.HasSuffix(buf.String(), "\n"))
}
