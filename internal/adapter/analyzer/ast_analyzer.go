package analyzer

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
)

// ASTAnalyzer implements port.Analyzer using Go's go/parser for import analysis.
type ASTAnalyzer struct{}

// NewASTAnalyzer constructs an ASTAnalyzer.
func NewASTAnalyzer() *ASTAnalyzer {
	return &ASTAnalyzer{}
}

// Analyze walks the project directory and checks all Go files for architecture violations.
func (a *ASTAnalyzer) Analyze(req port.AnalysisRequest) (domain.Report, error) {
	report := domain.Report{
		ProjectPath: req.ProjectDir,
		Arch:        req.Arch,
		Variant:     req.Variant,
	}

	var rules []LayerRule
	var err error

	if req.Variant == domain.VariantModular {
		// For modular variant, discover module names from subdirs of internal/.
		rules, err = rulesForModular(req.Arch, req.Variant, req.ProjectDir, req.ModulePath)
	} else {
		rules, err = rulesForArch(req.Arch, req.Variant, req.ModulePath)
	}
	if err != nil {
		return report, fmt.Errorf("analyze: build rules: %w", err)
	}

	// Collect all .go files to analyze.
	var goFiles []string
	err = filepath.WalkDir(req.ProjectDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		// Skip directories that should not be analyzed.
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip test files and non-Go files.
		if strings.HasSuffix(path, "_test.go") || !strings.HasSuffix(path, ".go") {
			return nil
		}
		goFiles = append(goFiles, path)
		return nil
	})
	if err != nil {
		return report, fmt.Errorf("analyze: walk dir: %w", err)
	}

	fset := token.NewFileSet()
	totalRuleEvaluations := len(rules) * len(goFiles)
	report.TotalRules = totalRuleEvaluations

	for _, filePath := range goFiles {
		astFile, parseErr := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
		if parseErr != nil {
			// Skip files that fail to parse (generated or invalid).
			continue
		}

		// Get relative path from project dir.
		relFilePath := strings.TrimPrefix(filePath, req.ProjectDir+"/")

		// Collect import paths for this file.
		var imports []string
		for _, imp := range astFile.Imports {
			rawPath := strings.Trim(imp.Path.Value, `"`)
			// Strip module path prefix to get the relative import path.
			importRelPath := strings.TrimPrefix(rawPath, req.ModulePath+"/")
			imports = append(imports, importRelPath)
		}

		// Evaluate each rule against this file's imports.
		for _, rule := range rules {
			if rule.Name == "adapter-no-lateral" {
				// Special lateral adapter check: if the file is under an adapter subtree,
				// check if it imports a different adapter subtree.
				if !strings.HasPrefix(relFilePath, rule.FromLayer+"/") && relFilePath != rule.FromLayer {
					continue
				}
				// Determine the immediate subdirectory of adapter this file belongs to.
				fileAdapterSub := adapterSubtree(relFilePath, rule.FromLayer)
				for _, imp := range imports {
					if !strings.HasPrefix(imp, rule.ToLayer+"/") && imp != rule.ToLayer {
						continue
					}
					importAdapterSub := adapterSubtree(imp, rule.ToLayer)
					if fileAdapterSub != "" && importAdapterSub != "" && fileAdapterSub != importAdapterSub {
						pos := fset.Position(astFile.Pos())
						report.Violations = append(report.Violations, domain.Violation{
							File:     relFilePath,
							Line:     pos.Line,
							Rule:     rule.Name,
							Message:  fmt.Sprintf("adapter subtree %q must not import adapter subtree %q", fileAdapterSub, importAdapterSub),
							Severity: rule.Severity,
						})
					}
				}
				continue
			}

			// Standard directional rule: check if this file is in the FromLayer.
			if !strings.HasPrefix(relFilePath, rule.FromLayer+"/") && relFilePath != rule.FromLayer {
				continue
			}

			// Check if any import is from the forbidden ToLayer.
			for _, imp := range imports {
				if strings.HasPrefix(imp, rule.ToLayer+"/") || imp == rule.ToLayer {
					pos := fset.Position(astFile.Pos())
					report.Violations = append(report.Violations, domain.Violation{
						File:     relFilePath,
						Line:     pos.Line,
						Rule:     rule.Name,
						Message:  fmt.Sprintf("package in %q must not import %q", rule.FromLayer, rule.ToLayer),
						Severity: rule.Severity,
					})
					break // one violation per rule per file is enough
				}
			}
		}
	}

	report.ComputeScore()
	return report, nil
}

// adapterSubtree returns the immediate subtree name under the adapter root for a given path.
// E.g. "internal/adapter/cli/root.go" with adapterRoot "internal/adapter" → "cli".
// Returns the full relative path segment if there is no subtree.
func adapterSubtree(relPath, adapterRoot string) string {
	trimmed := strings.TrimPrefix(relPath, adapterRoot+"/")
	if trimmed == relPath {
		return ""
	}
	parts := strings.SplitN(trimmed, "/", 2)
	return parts[0]
}

// rulesForModular discovers module names from immediate subdirectories of internal/
// and returns combined rules for all modules.
func rulesForModular(arch domain.Architecture, variant domain.Variant, projectDir, modulePath string) ([]LayerRule, error) {
	internalDir := filepath.Join(projectDir, "internal")
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		// If internal/ doesn't exist, fall back to default rules with empty module name.
		return rulesForArch(arch, variant, "")
	}

	var allRules []LayerRule
	seen := make(map[string]bool)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		moduleRules, err := rulesForArch(arch, variant, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("rulesForModular: module %s: %w", entry.Name(), err)
		}
		for _, r := range moduleRules {
			key := r.Name + "|" + r.FromLayer + "|" + r.ToLayer
			if !seen[key] {
				seen[key] = true
				allRules = append(allRules, r)
			}
		}
	}

	if len(allRules) == 0 {
		return rulesForArch(arch, variant, "")
	}
	return allRules, nil
}
