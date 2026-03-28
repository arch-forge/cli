package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/arch-forge/cli/internal/domain"
	"gopkg.in/yaml.v3"
)

// ModuleCreateOptions carries parameters for creating a new local module.
type ModuleCreateOptions struct {
	Name     string // module name (snake_case)
	Category string // module category (default: "custom")
	Dir      string // base directory for the module (default: "modules/")
}

// ModuleValidateOptions carries parameters for validating a local module.
type ModuleValidateOptions struct {
	ModuleDir string // path to the module directory
}

// ModuleUseCase handles custom module scaffold, watch, and validation.
type ModuleUseCase struct{}

// NewModuleUseCase creates a ModuleUseCase.
func NewModuleUseCase() *ModuleUseCase { return &ModuleUseCase{} }

// Create scaffolds a new local module directory with manifest, prompts, base/, and hooks/.
func (uc *ModuleUseCase) Create(opts ModuleCreateOptions) error {
	dir := opts.Dir
	if dir == "" {
		dir = "modules/"
	}

	category := opts.Category
	if category == "" {
		category = "custom"
	}

	moduleDir := filepath.Join(dir, opts.Name)

	if _, err := os.Stat(moduleDir); err == nil {
		return fmt.Errorf("create module: directory %q already exists", moduleDir)
	}

	// Create directories.
	if err := os.MkdirAll(filepath.Join(moduleDir, "base"), 0o755); err != nil {
		return fmt.Errorf("create module base dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(moduleDir, "hooks"), 0o755); err != nil {
		return fmt.Errorf("create module hooks dir: %w", err)
	}

	// Write module.yaml.
	moduleYAML := fmt.Sprintf(`name: %s
version: 1.0.0
description: "TODO: describe what this module does"
category: %s

architectures:
  - hexagonal
  - clean
  - standard

variants:
  - classic
  - modular

dependencies:
  required: []
  optional: []

options: []
patches: []
`, opts.Name, category)

	if err := os.WriteFile(filepath.Join(moduleDir, "module.yaml"), []byte(moduleYAML), 0o644); err != nil {
		return fmt.Errorf("write module.yaml: %w", err)
	}

	// Write prompts.yaml.
	promptsYAML := "prompts: []\n"
	if err := os.WriteFile(filepath.Join(moduleDir, "prompts.yaml"), []byte(promptsYAML), 0o644); err != nil {
		return fmt.Errorf("write prompts.yaml: %w", err)
	}

	// Write base/README.md.
	readme := fmt.Sprintf(`# Module: %s

Place your template files here under `+"`base/`"+`.

Template files should be named `+"`<target_filename>.tmpl`"+`.
For example: `+"`base/internal/%s/service.go.tmpl`"+`

## Available template variables

- `+"`{{ .Module }}`"+` — Go module path (e.g., github.com/acme/myapp)
- `+"`{{ .Project.Name }}`"+` — project name
- `+"`{{ .GoVersion }}`"+` — Go version
- `+"`{{ .Arch }}`"+` — architecture (hexagonal, clean, standard, etc.)
- `+"`{{ .Variant }}`"+` — variant (classic, modular)
- `+"`{{ index .Options \"key\" }}`"+` — module option value

## Hooks

Place hook templates in `+"`hooks/`"+` to inject code into existing files.
Hooks use anchor comments: `+"`// arch_forge:<anchor_name>`"+`
`, opts.Name, opts.Name)

	if err := os.WriteFile(filepath.Join(moduleDir, "base", "README.md"), []byte(readme), 0o644); err != nil {
		return fmt.Errorf("write base/README.md: %w", err)
	}

	// Write hooks/.gitkeep.
	if err := os.WriteFile(filepath.Join(moduleDir, "hooks", ".gitkeep"), []byte{}, 0o644); err != nil {
		return fmt.Errorf("write hooks/.gitkeep: %w", err)
	}

	return nil
}

// localManifest is used to parse a module.yaml for validation purposes.
type localManifest struct {
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	Architectures []string `yaml:"architectures"`
	Variants      []string `yaml:"variants"`
}

// Validate checks a module directory for manifest completeness and template correctness.
func (uc *ModuleUseCase) Validate(opts ModuleValidateOptions) (domain.LocalModuleValidation, error) {
	moduleDir := opts.ModuleDir
	moduleName := filepath.Base(moduleDir)

	result := domain.LocalModuleValidation{
		ModuleName: moduleName,
		ModuleDir:  moduleDir,
	}

	// Check module directory exists.
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		return result, fmt.Errorf("validate module: directory %q does not exist", moduleDir)
	}

	manifestPath := filepath.Join(moduleDir, "module.yaml")
	manifestExists := true

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		manifestExists = false
		result.Issues = append(result.Issues, domain.LocalModuleIssue{
			Kind:    "error",
			Message: "module.yaml not found",
			Field:   "module.yaml",
		})
	}

	if manifestExists {
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			return result, fmt.Errorf("read module.yaml: %w", err)
		}

		var manifest localManifest
		if err := yaml.Unmarshal(data, &manifest); err != nil {
			result.Issues = append(result.Issues, domain.LocalModuleIssue{
				Kind:    "error",
				Message: fmt.Sprintf("module.yaml parse error: %s", err.Error()),
				Field:   "module.yaml",
			})
		} else {
			// Validate name.
			if manifest.Name == "" {
				result.Issues = append(result.Issues, domain.LocalModuleIssue{
					Kind:    "error",
					Message: "name field is missing or empty",
					Field:   "name",
				})
			} else if manifest.Name != moduleName {
				result.Issues = append(result.Issues, domain.LocalModuleIssue{
					Kind:    "warning",
					Message: fmt.Sprintf("name %q does not match directory basename %q", manifest.Name, moduleName),
					Field:   "name",
				})
			}

			// Validate description.
			if manifest.Description == "" || strings.Contains(manifest.Description, "TODO:") {
				result.Issues = append(result.Issues, domain.LocalModuleIssue{
					Kind:    "warning",
					Message: "description is empty or still contains placeholder text",
					Field:   "description",
				})
			}

			// Validate architectures.
			if len(manifest.Architectures) == 0 {
				result.Issues = append(result.Issues, domain.LocalModuleIssue{
					Kind:    "warning",
					Message: "architectures list is empty; module will not match any architecture",
					Field:   "architectures",
				})
			}

			// Validate variants.
			if len(manifest.Variants) == 0 {
				result.Issues = append(result.Issues, domain.LocalModuleIssue{
					Kind:    "warning",
					Message: "variants list is empty; module will not match any variant",
					Field:   "variants",
				})
			}
		}
	}

	// Check prompts.yaml (optional).
	promptsPath := filepath.Join(moduleDir, "prompts.yaml")
	if _, err := os.Stat(promptsPath); os.IsNotExist(err) {
		result.Issues = append(result.Issues, domain.LocalModuleIssue{
			Kind:    "warning",
			Message: "prompts.yaml not found (optional but recommended)",
			Field:   "prompts.yaml",
		})
	}

	// Check base/ directory.
	baseDir := filepath.Join(moduleDir, "base")
	baseExists := true
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		baseExists = false
		result.Issues = append(result.Issues, domain.LocalModuleIssue{
			Kind:    "warning",
			Message: "base/ directory not found",
			Field:   "base/",
		})
	}

	// Walk base/ and validate .tmpl files.
	if baseExists {
		tmplCount, tmplErrors := validateTemplateFiles(baseDir)
		result.Issues = append(result.Issues, tmplErrors...)
		_ = tmplCount
	}

	// Check hooks/ directory.
	hooksDir := filepath.Join(moduleDir, "hooks")
	hooksExists := true
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		hooksExists = false
		result.Issues = append(result.Issues, domain.LocalModuleIssue{
			Kind:    "warning",
			Message: "hooks/ directory not found",
			Field:   "hooks/",
		})
	}

	// Walk hooks/ and validate .go.tmpl files.
	if hooksExists {
		_, hookErrors := validateTemplateFiles(hooksDir)
		result.Issues = append(result.Issues, hookErrors...)
	}

	// Determine overall status.
	hasErrors := false
	hasWarnings := false
	for _, issue := range result.Issues {
		switch issue.Kind {
		case "error":
			hasErrors = true
		case "warning":
			hasWarnings = true
		}
	}

	switch {
	case hasErrors:
		result.Status = domain.LocalModuleStatusInvalid
	case hasWarnings:
		result.Status = domain.LocalModuleStatusWarning
	default:
		result.Status = domain.LocalModuleStatusValid
	}

	return result, nil
}

// validateTemplateFiles walks dir and tries parsing each .tmpl file as a Go template.
// Returns the count of successfully parsed templates and a slice of error issues.
func validateTemplateFiles(dir string) (int, []domain.LocalModuleIssue) {
	var issues []domain.LocalModuleIssue
	count := 0

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".tmpl") {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			issues = append(issues, domain.LocalModuleIssue{
				Kind:    "error",
				Message: fmt.Sprintf("cannot read template file: %s", readErr.Error()),
				Field:   path,
			})
			return nil
		}

		if _, parseErr := template.New("").Parse(string(content)); parseErr != nil {
			issues = append(issues, domain.LocalModuleIssue{
				Kind:    "error",
				Message: fmt.Sprintf("template parse error: %s", parseErr.Error()),
				Field:   path,
			})
			return nil
		}

		count++
		return nil
	})

	if err != nil {
		issues = append(issues, domain.LocalModuleIssue{
			Kind:    "error",
			Message: fmt.Sprintf("walk template directory: %s", err.Error()),
			Field:   dir,
		})
	}

	return count, issues
}
