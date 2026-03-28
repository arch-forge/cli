package generator

import (
	"path"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/arch-forge/cli/internal/domain"
)

// buildFuncMap returns the complete template.FuncMap used by Engine.
// All functions are pure (no I/O, no mutable state).
func buildFuncMap() template.FuncMap {
	return template.FuncMap{
		"camelCase":   camelCase,
		"pascalCase":  pascalCase,
		"snakeCase":   snakeCase,
		"kebabCase":   kebabCase,
		"upperCase":   strings.ToUpper,
		"lowerCase":   strings.ToLower,
		"domainPath":  func(p domain.ResolvedPaths) string { return p.Domain },
		"portPath":    func(p domain.ResolvedPaths) string { return p.Port },
		"appPath":     func(p domain.ResolvedPaths) string { return p.App },
		"adapterPath": func(p domain.ResolvedPaths) string { return p.Adapter },
		"handlerPath": func(p domain.ResolvedPaths) string { return p.Handler },
		"repoPath":    func(p domain.ResolvedPaths) string { return p.Repository },
		"joinPath":    func(elems ...string) string { return path.Join(elems...) },
		"hasModule":   hasModule,
		"goPackage":   goPackage,
		"currentYear": func() int { return time.Now().Year() },
		"quote":       func(s string) string { return `"` + s + `"` },
		"trimSuffix":  strings.TrimSuffix,
		"replace":     func(s, old, new string) string { return strings.ReplaceAll(s, old, new) },
		"default":     templateDefault,
	}
}

// splitWords splits s into lowercase words on word boundaries:
// underscores, hyphens, and transitions between lowercase and uppercase letters.
func splitWords(s string) []string {
	var words []string
	var current strings.Builder

	runes := []rune(s)
	for i, r := range runes {
		if r == '_' || r == '-' {
			if current.Len() > 0 {
				words = append(words, strings.ToLower(current.String()))
				current.Reset()
			}
			continue
		}

		// Detect lowercase→uppercase transition (e.g. "myField": 'y'→'F')
		if i > 0 && unicode.IsUpper(r) && unicode.IsLower(runes[i-1]) {
			if current.Len() > 0 {
				words = append(words, strings.ToLower(current.String()))
				current.Reset()
			}
		}

		// Detect start of an acronym → word boundary (e.g. "XMLParser": 'L'→'P')
		if i > 0 && unicode.IsUpper(r) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
			if current.Len() > 1 {
				words = append(words, strings.ToLower(current.String()))
				current.Reset()
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, strings.ToLower(current.String()))
	}

	return words
}

// camelCase converts snake_case or kebab-case to camelCase.
// Example: "my_field" → "myField", "my-field" → "myField".
func camelCase(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return s
	}
	var b strings.Builder
	b.WriteString(words[0])
	for _, w := range words[1:] {
		if len(w) == 0 {
			continue
		}
		b.WriteString(strings.ToUpper(w[:1]) + w[1:])
	}
	return b.String()
}

// pascalCase converts any casing to PascalCase.
// Example: "my_field" → "MyField".
func pascalCase(s string) string {
	words := splitWords(s)
	var b strings.Builder
	for _, w := range words {
		if len(w) == 0 {
			continue
		}
		b.WriteString(strings.ToUpper(w[:1]) + w[1:])
	}
	return b.String()
}

// snakeCase converts PascalCase or camelCase to snake_case.
// Example: "MyField" → "my_field".
func snakeCase(s string) string {
	words := splitWords(s)
	return strings.Join(words, "_")
}

// kebabCase converts PascalCase or camelCase to kebab-case.
// Example: "MyField" → "my-field".
func kebabCase(s string) string {
	words := splitWords(s)
	return strings.Join(words, "-")
}

// hasModule reports whether name is present in the modules slice.
func hasModule(modules []string, name string) bool {
	for _, m := range modules {
		if m == name {
			return true
		}
	}
	return false
}

// templateDefault returns fallback if value is the zero value for its type
// (empty string, nil, zero int, false). Mirrors the Helm/Sprig `default` function.
func templateDefault(fallback, value any) any {
	if value == nil {
		return fallback
	}
	switch v := value.(type) {
	case string:
		if v == "" {
			return fallback
		}
	case bool:
		if !v {
			return fallback
		}
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		if v == 0 {
			return fallback
		}
	}
	return value
}

// goPackage returns the last path component of p as a valid Go identifier,
// replacing hyphens with underscores.
func goPackage(p string) string {
	base := path.Base(p)
	return strings.ReplaceAll(base, "-", "_")
}
