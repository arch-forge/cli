package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arch-forge/cli/internal/adapter/cli/tui"
	"github.com/arch-forge/cli/internal/app"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newInitCmd(uc *app.InitUseCase) *cobra.Command {
	var (
		arch       string
		variant    string
		modules    []string
		modulePath string
		goVersion  string
		dryRun     bool
		preset     string
	)

	cmd := &cobra.Command{
		Use:   "init <project-name>",
		Short: "Create a new Go project with the chosen architecture",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// If no arch specified and not using a preset, launch the interactive wizard.
			if arch == "" && preset == "" && !cmd.Flags().Changed("arch") {
				if isInteractiveTerminal() {
					result, wizErr := tui.Run(tui.WizardInput{
						ProjectName: name,
						ModulePath:  modulePath,
						GoVersion:   goVersion,
					})
					if wizErr != nil {
						return fmt.Errorf("wizard: %w", wizErr)
					}
					if result.Cancelled {
						fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
						return nil
					}
					// Apply wizard results.
					arch = string(result.Architecture)
					variant = string(result.Variant)
					modules = result.Modules
					if modulePath == "" && result.ModulePath != "" {
						modulePath = result.ModulePath
					}
				} else {
					fmt.Fprintln(cmd.ErrOrStderr(), "No interactive terminal detected. Using defaults (hexagonal/modular).")
					fmt.Fprintln(cmd.ErrOrStderr(), "Tip: use --arch, --variant, and --modules flags to customize, or --preset for a quick start.")
				}
			}

			// Default module path if not specified.
			if modulePath == "" {
				modulePath = fmt.Sprintf("github.com/your-org/%s", name)
			}
			if goVersion == "" {
				goVersion = "1.23"
			}

			// Apply arch/variant defaults only when no preset is provided.
			// When a preset is given, defaults are applied inside the use case
			// so that explicit flags can override preset values.
			if preset == "" {
				if arch == "" {
					arch = "hexagonal"
				}
				if variant == "" {
					variant = "modular"
				}
			}

			var dryFs afero.Fs
			if dryRun {
				dryFs = afero.NewMemMapFs()
			}

			opts := app.InitOptions{
				Name:       name,
				ModulePath: modulePath,
				Arch:       domain.Architecture(arch),
				Variant:    domain.Variant(variant),
				Modules:    modules,
				DryRun:     dryRun,
				OutputDir:  name,
				GoVersion:  goVersion,
				Preset:     preset,
				Fs:         dryFs,
			}

			if err := uc.Execute(cmd.Context(), opts); err != nil {
				return fmt.Errorf("init: %w", err)
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Dry run — files that would be created:\n\n")
				fmt.Fprintf(cmd.OutOrStdout(), "%s/\n", name)
				if err := renderFsTree(dryFs, ".", "", cmd.OutOrStdout()); err != nil {
					return fmt.Errorf("render dry-run tree: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout())
				fmt.Fprintln(cmd.OutOrStdout(), "No files written.")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Project %q created successfully.\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&arch, "arch", "", "Architecture pattern (hexagonal, clean, standard, ddd, modular_monolith, cqrs, microservice)")
	cmd.Flags().StringVar(&variant, "variant", "", "Architecture variant (classic, modular)")
	cmd.Flags().StringSliceVar(&modules, "modules", nil, "Comma-separated list of modules to include")
	cmd.Flags().StringVar(&modulePath, "module-path", "", "Go module path (e.g. github.com/acme/myapp)")
	cmd.Flags().StringVar(&goVersion, "go-version", "1.23", "Go version to use")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview generated files without writing to disk")
	cmd.Flags().StringVar(&preset, "preset", "", "Use a named preset (starter, production-api, microservice)")

	// Register dynamic completion for --arch flag.
	_ = cmd.RegisterFlagCompletionFunc("arch", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		archs := domain.AllArchitectures()
		completions := make([]string, len(archs))
		for i, a := range archs {
			completions[i] = string(a.Value) + "\t" + a.Description
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	})

	// Register dynamic completion for --variant flag.
	_ = cmd.RegisterFlagCompletionFunc("variant", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"classic\tBy-the-book implementation with canonical nomenclature",
			"modular\tOrganized by business domain (default)",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	// Register dynamic completion for --modules flag.
	_ = cmd.RegisterFlagCompletionFunc("modules", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"api\tHTTP API server with chi router",
			"database\tPostgreSQL connection and pool",
			"logging\tStructured logging with slog",
			"docker\tMulti-stage Dockerfile and docker-compose",
			"makefile\tStandard Makefile targets",
			"auth\tJWT authentication middleware",
			"cache\tRedis cache client",
			"grpc\tgRPC server with interceptors",
			"crud\tCRUD scaffold for an entity",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	// Register dynamic completion for --preset flag.
	_ = cmd.RegisterFlagCompletionFunc("preset", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		presets := domain.AllPresets()
		completions := make([]string, len(presets))
		for i, p := range presets {
			completions[i] = p.Name + "\t" + p.Description
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

// renderFsTree walks fs and prints a directory tree with box-drawing connectors.
// It uses afero.Walk to avoid relying on implicit directory entries in MemMapFs.
func renderFsTree(fs afero.Fs, root string, _ string, w io.Writer) error {
	// Collect all entries: map from dir → sorted list of children.
	type fsEntry struct {
		name  string
		isDir bool
	}
	children := map[string][]fsEntry{}

	err := afero.Walk(fs, root, func(p string, info os.FileInfo, err error) error {
		if err != nil || p == root || p == "." {
			return err
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil || rel == "." {
			return nil
		}
		dir := filepath.Dir(rel)
		if dir == "." {
			dir = ""
		}
		children[dir] = append(children[dir], fsEntry{name: filepath.Base(rel), isDir: info.IsDir()})
		return nil
	})
	if err != nil {
		return err
	}

	// Sort each directory's children: dirs first, then files, both alphabetically.
	for dir := range children {
		sort.Slice(children[dir], func(i, j int) bool {
			a, b := children[dir][i], children[dir][j]
			if a.isDir != b.isDir {
				return a.isDir // dirs before files
			}
			return a.name < b.name
		})
	}

	// Render recursively.
	var render func(dir, prefix string)
	render = func(dir, prefix string) {
		entries := children[dir]
		for i, e := range entries {
			isLast := i == len(entries)-1
			connector := "├── "
			childPrefix := prefix + "│   "
			if isLast {
				connector = "└── "
				childPrefix = prefix + "    "
			}
			if e.isDir {
				fmt.Fprintf(w, "%s%s%s/\n", prefix, connector, e.name)
				childDir := e.name
				if dir != "" {
					childDir = dir + string(filepath.Separator) + e.name
				}
				render(childDir, childPrefix)
			} else {
				fmt.Fprintf(w, "%s%s%s\n", prefix, connector, e.name)
			}
		}
	}
	render("", "")
	return nil
}

// keep strings import used by renderFsTree's sort closure.
var _ = strings.Join

// isInteractiveTerminal reports whether a controlling terminal is available.
// It mirrors what bubbletea does internally: attempt to open /dev/tty.
func isInteractiveTerminal() bool {
	f, err := os.Open("/dev/tty")
	if err != nil {
		return false
	}
	f.Close()
	return true
}
