package app

import (
	"context"
	"fmt"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
)

// ArchSummary is a display-ready descriptor for an architecture.
type ArchSummary struct {
	Value       domain.Architecture
	DisplayName string
	Description string
}

// ModuleSummary is a display-ready descriptor for a module.
type ModuleSummary struct {
	Name        string
	Category    domain.ModuleCategory
	Description string
	Version     string
}

// ListUseCase provides read-only queries over available architectures
// and modules.
type ListUseCase struct {
	repo port.TemplateRepository
}

// NewListUseCase constructs a ListUseCase.
func NewListUseCase(repo port.TemplateRepository) *ListUseCase {
	return &ListUseCase{repo: repo}
}

// Architectures returns all supported architectures with display metadata.
// Data comes from domain.AllArchitectures() — no error possible.
func (uc *ListUseCase) Architectures() []ArchSummary {
	infos := domain.AllArchitectures()
	summaries := make([]ArchSummary, len(infos))
	for i, info := range infos {
		summaries[i] = ArchSummary{
			Value:       info.Value,
			DisplayName: info.DisplayName,
			Description: info.Description,
		}
	}
	return summaries
}

// Presets returns all available presets.
func (uc *ListUseCase) Presets() []domain.Preset {
	return domain.AllPresets()
}

// Modules returns all modules available in the repository, optionally
// filtered by category. Pass empty string for all categories.
func (uc *ListUseCase) Modules(ctx context.Context, category string) ([]ModuleSummary, error) {
	names, err := uc.repo.ListModules()
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}

	var summaries []ModuleSummary
	for _, name := range names {
		mod, err := uc.repo.LoadModuleManifest(name)
		if err != nil {
			return nil, fmt.Errorf("list modules: %w", err)
		}

		if category != "" && mod.Category != domain.ModuleCategory(category) {
			continue
		}

		summaries = append(summaries, ModuleSummary{
			Name:        mod.Name,
			Category:    mod.Category,
			Description: mod.Description,
			Version:     mod.Version,
		})
	}

	return summaries, nil
}
