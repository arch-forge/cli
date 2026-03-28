package port

import (
	"context"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/spf13/afero"
)

// TemplateFile is a single template entry returned by TemplateRepository.
type TemplateFile struct {
	// RelPath is the template path relative to the template root,
	// e.g. "cmd/api/main.go.tmpl"
	RelPath string
	// Content is the raw template source bytes.
	Content []byte
}

// Generator renders a set of TemplateFiles into a filesystem.
type Generator interface {
	// Generate renders all templates using tctx and writes the output
	// files to fs. The destination path for each file is derived by
	// stripping the ".tmpl" suffix from TemplateFile.RelPath.
	Generate(ctx context.Context, tctx domain.TemplateContext, templates []TemplateFile, fs afero.Fs) error
}
