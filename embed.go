// Package archforge exposes the embedded templates filesystem.
package archforge

import "embed"

// TemplateFS is the embedded filesystem containing all arch_forge templates.
// The pattern uses "all:" to include hidden sentinel files (e.g. .gitkeep).
//
//go:embed all:templates/go
var TemplateFS embed.FS
