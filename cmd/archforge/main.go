package main

import (
	"github.com/arch-forge/cli/internal/adapter/cli"
	configadapter "github.com/arch-forge/cli/internal/adapter/config"
	generatoradapter "github.com/arch-forge/cli/internal/adapter/generator"
	graphadapter "github.com/arch-forge/cli/internal/adapter/graph"
	patcheradapter "github.com/arch-forge/cli/internal/adapter/patcher"
	repositoryadapter "github.com/arch-forge/cli/internal/adapter/repository"
	analyzeradapter "github.com/arch-forge/cli/internal/adapter/analyzer"
	scanneradapter "github.com/arch-forge/cli/internal/adapter/scanner"
	updateradapter "github.com/arch-forge/cli/internal/adapter/updater"
	"github.com/arch-forge/cli/internal/app"
)

func main() {
	repo := repositoryadapter.NewEmbeddedRepository()
	gen := generatoradapter.NewEngine()
	cfg := configadapter.NewViperConfig()
	patcher := patcheradapter.NewFilePatcher()

	initUC := app.NewInitUseCase(repo, gen, cfg)
	addUC := app.NewAddUseCase(repo, gen, cfg, patcher)
	listUC := app.NewListUseCase(repo)
	analyzer := analyzeradapter.NewASTAnalyzer()
	doctorUC := app.NewDoctorUseCase(cfg, analyzer)

	scannerAdapter := scanneradapter.NewOsScanner()
	inspectUC := app.NewInspectUseCase(cfg, scannerAdapter)

	graphBuilder := graphadapter.NewASTGraphBuilder()
	graphUC := app.NewGraphUseCase(cfg, graphBuilder)

	moduleUC := app.NewModuleUseCase()

	updater := updateradapter.NewGithubUpdater()
	updateUC := app.NewUpdateUseCase(updater, cli.Version)

	domainUC := app.NewDomainAddUseCase(repo, gen, cfg)
	cli.Execute(initUC, addUC, listUC, doctorUC, inspectUC, graphUC, moduleUC, updateUC, domainUC)
}
