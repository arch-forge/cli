package generator_test

import (
	"context"
	"testing"

	"github.com/arch-forge/cli/internal/adapter/generator"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_Generate_BasicTemplate(t *testing.T) {
	eng := generator.NewEngine()
	fs := afero.NewMemMapFs()

	tctx := domain.TemplateContext{
		Project: domain.ProjectInfo{Name: "testapp", ModulePath: "github.com/test/testapp"},
		Module:  "github.com/test/testapp",
	}

	templates := []port.TemplateFile{
		{
			RelPath: "cmd/app/main.go.tmpl",
			Content: []byte("package main\n// Module: {{ .Module }}\nfunc main() {}"),
		},
	}

	err := eng.Generate(context.Background(), tctx, templates, fs)
	require.NoError(t, err)

	// File should exist without .tmpl suffix.
	content, err := afero.ReadFile(fs, "cmd/app/main.go")
	require.NoError(t, err)
	assert.Contains(t, string(content), "github.com/test/testapp")
}

func TestEngine_Generate_StripsTmplSuffix(t *testing.T) {
	eng := generator.NewEngine()
	fs := afero.NewMemMapFs()

	templates := []port.TemplateFile{
		{RelPath: "internal/domain/model.go.tmpl", Content: []byte("package domain")},
	}

	err := eng.Generate(context.Background(), domain.TemplateContext{}, templates, fs)
	require.NoError(t, err)

	// .tmpl file should NOT exist.
	_, err = fs.Stat("internal/domain/model.go.tmpl")
	assert.Error(t, err, "the .tmpl file should not be written")

	// Rendered file without suffix should exist.
	_, err = fs.Stat("internal/domain/model.go")
	assert.NoError(t, err, "the rendered .go file should exist")
}

func TestEngine_Generate_ContextCancellation(t *testing.T) {
	eng := generator.NewEngine()
	fs := afero.NewMemMapFs()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	templates := []port.TemplateFile{
		{RelPath: "file.go.tmpl", Content: []byte("package main")},
	}

	err := eng.Generate(ctx, domain.TemplateContext{}, templates, fs)
	assert.Error(t, err)
}

func TestEngine_Generate_InvalidTemplate(t *testing.T) {
	eng := generator.NewEngine()
	fs := afero.NewMemMapFs()

	templates := []port.TemplateFile{
		// Malformed template: unclosed action
		{RelPath: "bad.go.tmpl", Content: []byte("package main\n{{ .NonExistentField")},
	}

	err := eng.Generate(context.Background(), domain.TemplateContext{}, templates, fs)
	assert.Error(t, err)
}

func TestEngine_Generate_TemplateFuncs(t *testing.T) {
	eng := generator.NewEngine()
	fs := afero.NewMemMapFs()

	tctx := domain.TemplateContext{
		Project: domain.ProjectInfo{Name: "my-app", ModulePath: "github.com/test/my-app"},
		Module:  "github.com/test/my-app",
	}

	templates := []port.TemplateFile{
		{
			RelPath: "output.txt.tmpl",
			Content: []byte(`pascal:{{ pascalCase "my_field" }} camel:{{ camelCase "my_field" }} snake:{{ snakeCase "MyField" }}`),
		},
	}

	err := eng.Generate(context.Background(), tctx, templates, fs)
	require.NoError(t, err)

	content, err := afero.ReadFile(fs, "output.txt")
	require.NoError(t, err)
	assert.Contains(t, string(content), "pascal:MyField")
	assert.Contains(t, string(content), "camel:myField")
	assert.Contains(t, string(content), "snake:my_field")
}

func TestEngine_Generate_MultipleTemplates(t *testing.T) {
	eng := generator.NewEngine()
	fs := afero.NewMemMapFs()

	tctx := domain.TemplateContext{
		Project: domain.ProjectInfo{Name: "testapp", ModulePath: "github.com/test/testapp"},
		Module:  "github.com/test/testapp",
	}

	templates := []port.TemplateFile{
		{RelPath: "cmd/main.go.tmpl", Content: []byte("package main")},
		{RelPath: "internal/domain/entity.go.tmpl", Content: []byte("package domain")},
		{RelPath: "go.mod.tmpl", Content: []byte("module {{ .Module }}")},
	}

	err := eng.Generate(context.Background(), tctx, templates, fs)
	require.NoError(t, err)

	for _, expected := range []string{"cmd/main.go", "internal/domain/entity.go", "go.mod"} {
		_, err := fs.Stat(expected)
		assert.NoError(t, err, "expected file %q to exist", expected)
	}

	gomod, err := afero.ReadFile(fs, "go.mod")
	require.NoError(t, err)
	assert.Equal(t, "module github.com/test/testapp", string(gomod))
}
