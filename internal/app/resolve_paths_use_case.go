package app

import (
	"github.com/arch-forge/cli/internal/domain"
)

// ResolvePathsUseCase resolves the canonical directory paths for a given
// architecture, variant, and module name. It is stateless.
type ResolvePathsUseCase struct{}

// NewResolvePathsUseCase constructs a ResolvePathsUseCase.
func NewResolvePathsUseCase() *ResolvePathsUseCase {
	return &ResolvePathsUseCase{}
}

// Execute resolves and returns the canonical directory paths.
func (uc *ResolvePathsUseCase) Execute(
	arch domain.Architecture,
	variant domain.Variant,
	moduleName string,
) (domain.ResolvedPaths, error) {
	return domain.ResolvePaths(arch, variant, moduleName)
}
