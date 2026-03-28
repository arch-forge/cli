package patcher_test

import (
	"context"
	"testing"

	"github.com/arch-forge/cli/internal/adapter/patcher"
	"github.com/arch-forge/cli/internal/port"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilePatcher_Apply_InjectAfter(t *testing.T) {
	p := patcher.NewFilePatcher()
	fs := afero.NewMemMapFs()

	content := "package main\n\nfunc main() {\n\t// arch_forge:routes\n}\n"
	err := afero.WriteFile(fs, "/project/cmd/main.go", []byte(content), 0o644)
	require.NoError(t, err)

	req := port.PatchRequest{
		TargetGlob: "cmd/main.go",
		Action:     "inject_after",
		Anchor:     "routes",
		Content:    "\tr.Mount(\"/api\", apiRouter)",
		Optional:   false,
	}

	err = p.Apply(context.Background(), "/project", []port.PatchRequest{req}, fs)
	require.NoError(t, err)

	result, err := afero.ReadFile(fs, "/project/cmd/main.go")
	require.NoError(t, err)

	resultStr := string(result)
	assert.Contains(t, resultStr, "r.Mount(\"/api\", apiRouter)")
	// Anchor comment should still be present.
	assert.Contains(t, resultStr, "// arch_forge:routes")
}

func TestFilePatcher_Apply_InjectBefore(t *testing.T) {
	p := patcher.NewFilePatcher()
	fs := afero.NewMemMapFs()

	content := "package main\n\nfunc main() {\n\t// arch_forge:imports\n}\n"
	err := afero.WriteFile(fs, "/project/cmd/main.go", []byte(content), 0o644)
	require.NoError(t, err)

	req := port.PatchRequest{
		TargetGlob: "cmd/main.go",
		Action:     "inject_before",
		Anchor:     "imports",
		Content:    "\timport \"fmt\"",
		Optional:   false,
	}

	err = p.Apply(context.Background(), "/project", []port.PatchRequest{req}, fs)
	require.NoError(t, err)

	result, err := afero.ReadFile(fs, "/project/cmd/main.go")
	require.NoError(t, err)

	resultStr := string(result)
	assert.Contains(t, resultStr, "import \"fmt\"")
	assert.Contains(t, resultStr, "// arch_forge:imports")

	// Verify injected content appears before the anchor.
	importIdx := indexOf(resultStr, "import \"fmt\"")
	anchorIdx := indexOf(resultStr, "// arch_forge:imports")
	assert.Less(t, importIdx, anchorIdx, "injected content should appear before the anchor")
}

func TestFilePatcher_Apply_Replace(t *testing.T) {
	p := patcher.NewFilePatcher()
	fs := afero.NewMemMapFs()

	content := "package main\n\n// arch_forge:version\n\nfunc main() {}\n"
	err := afero.WriteFile(fs, "/project/version.go", []byte(content), 0o644)
	require.NoError(t, err)

	req := port.PatchRequest{
		TargetGlob: "version.go",
		Action:     "replace",
		Anchor:     "version",
		Content:    "const Version = \"1.0.0\"",
		Optional:   false,
	}

	err = p.Apply(context.Background(), "/project", []port.PatchRequest{req}, fs)
	require.NoError(t, err)

	result, err := afero.ReadFile(fs, "/project/version.go")
	require.NoError(t, err)

	resultStr := string(result)
	assert.Contains(t, resultStr, "const Version = \"1.0.0\"")
	// Original anchor line replaced — should not be present.
	assert.NotContains(t, resultStr, "// arch_forge:version")
}

func TestFilePatcher_Apply_OptionalMissing(t *testing.T) {
	p := patcher.NewFilePatcher()
	fs := afero.NewMemMapFs()

	// Create the root dir so the walk can start, but no files match the glob.
	require.NoError(t, fs.MkdirAll("/project/cmd", 0o755))

	req := port.PatchRequest{
		TargetGlob: "nonexistent/*.go",
		Action:     "inject_after",
		Anchor:     "routes",
		Content:    "something",
		Optional:   true,
	}

	err := p.Apply(context.Background(), "/project", []port.PatchRequest{req}, fs)
	assert.NoError(t, err, "optional=true with no matching files should not return error")
}

func TestFilePatcher_Apply_RequiredMissingAnchor(t *testing.T) {
	p := patcher.NewFilePatcher()
	fs := afero.NewMemMapFs()

	// File exists but does not have the anchor.
	content := "package main\n\nfunc main() {}\n"
	err := afero.WriteFile(fs, "/project/main.go", []byte(content), 0o644)
	require.NoError(t, err)

	req := port.PatchRequest{
		TargetGlob: "main.go",
		Action:     "inject_after",
		Anchor:     "missing-anchor",
		Content:    "something",
		Optional:   false,
	}

	err = p.Apply(context.Background(), "/project", []port.PatchRequest{req}, fs)
	assert.Error(t, err, "optional=false with no anchor should return error")
}

func TestFilePatcher_Apply_MultipleFiles(t *testing.T) {
	p := patcher.NewFilePatcher()
	fs := afero.NewMemMapFs()

	fileContent := "package routes\n\n// arch_forge:mount\n"
	err := afero.WriteFile(fs, "/project/routes/v1.go", []byte(fileContent), 0o644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "/project/routes/v2.go", []byte(fileContent), 0o644)
	require.NoError(t, err)

	req := port.PatchRequest{
		TargetGlob: "routes/*.go",
		Action:     "inject_after",
		Anchor:     "mount",
		Content:    "// new route added",
		Optional:   false,
	}

	err = p.Apply(context.Background(), "/project", []port.PatchRequest{req}, fs)
	require.NoError(t, err)

	for _, f := range []string{"/project/routes/v1.go", "/project/routes/v2.go"} {
		result, err := afero.ReadFile(fs, f)
		require.NoError(t, err)
		assert.Contains(t, string(result), "// new route added", "file %s should be patched", f)
	}
}

// indexOf returns the byte index of substr in s, or -1 if not found.
func indexOf(s, substr string) int {
	for i := range s {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
