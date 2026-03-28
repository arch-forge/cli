package generator

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"
	"text/template"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/spf13/afero"
)

// Engine implements port.Generator using text/template.
type Engine struct {
	funcs template.FuncMap
}

// NewEngine constructs a ready-to-use Engine with all template funcs registered.
func NewEngine() *Engine {
	return &Engine{funcs: buildFuncMap()}
}

// Generate implements port.Generator.
func (e *Engine) Generate(
	ctx context.Context,
	tctx domain.TemplateContext,
	templates []port.TemplateFile,
	fs afero.Fs,
) error {
	for _, tmpl := range templates {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		rendered, err := e.render(tmpl, tctx)
		if err != nil {
			return fmt.Errorf("render %s: %w", tmpl.RelPath, err)
		}

		dest := destPath(tmpl.RelPath)
		if err := writeFile(fs, dest, rendered); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}
	}
	return nil
}

// render executes a single template file with tctx as the data value.
func (e *Engine) render(tmpl port.TemplateFile, tctx domain.TemplateContext) ([]byte, error) {
	t, err := template.New(tmpl.RelPath).Funcs(e.funcs).Parse(string(tmpl.Content))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, tctx); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return buf.Bytes(), nil
}

// destPath strips the ".tmpl" suffix from a relative template path.
func destPath(relPath string) string {
	return strings.TrimSuffix(relPath, ".tmpl")
}

// writeFile writes content to path in fs, creating intermediate directories.
func writeFile(fs afero.Fs, filePath string, content []byte) error {
	dir := path.Dir(filePath)
	if err := fs.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}
	f, err := fs.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file %s: %w", filePath, err)
	}
	defer f.Close()
	if _, err := f.Write(content); err != nil {
		return fmt.Errorf("write file %s: %w", filePath, err)
	}
	return nil
}
