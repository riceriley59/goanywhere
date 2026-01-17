package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/riceriley59/goanywhere/internal/core"
	"github.com/riceriley59/goanywhere/internal/core/factory"
)

type buildOptions struct {
	OutputDir   string
	ImportPath  string
	Plugin      string
	BuildSystem string
	LibraryName string
	Verbose     bool
}

// NewBuildCmd creates the build subcommand
func NewBuildCmd() *cobra.Command {
	opts := &buildOptions{}

	cmd := &cobra.Command{
		Use:   "build <input-directory>",
		Short: "Generate and build plugin code for a Go package",
		Long: `Generate plugin code and build it as a shared library or package.

For CGO plugin:
  Generates the CGO wrapper code and compiles it to a shared library (.so/.dylib/.dll)

For Python plugin:
  Generates CGO shared library, Python bindings, and creates a Python package
  with the specified build system configuration.

Examples:
  goanywhere build ./mypackage --plugin cgo
  goanywhere build ./mypackage --plugin cgo -o ./dist
  goanywhere build ./mypackage --plugin python --build-system setuptools`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(args[0], opts)
		},
	}

	cmd.Flags().StringVarP(&opts.OutputDir, "output", "o", "",
		"Output directory (default: <input>/<plugin>_build)")
	cmd.Flags().StringVarP(&opts.ImportPath, "import-path", "i", "",
		"Import path for the target package (required for proper imports)")
	cmd.Flags().StringVarP(&opts.Plugin, "plugin", "p", "cgo",
		"Plugin type to build (cgo, python)")
	cmd.Flags().StringVar(&opts.BuildSystem, "build-system", "setuptools",
		"Python build system (setuptools, hatch, poetry, uv)")
	cmd.Flags().StringVar(&opts.LibraryName, "lib-name", "",
		"Override the shared library name (default: lib<package>)")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false,
		"Verbose output")

	return cmd
}

func runBuild(inputDir string, opts *buildOptions) error {
	// Resolve input path
	inputPath, err := filepath.Abs(inputDir)
	if err != nil {
		return fmt.Errorf("invalid input path: %w", err)
	}

	// Check if input directory exists
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("cannot access input directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("input path is not a directory: %s", inputPath)
	}

	// Parse the Go package
	parser := core.NewParser(opts.Verbose)
	if opts.Verbose {
		fmt.Printf("Parsing package at: %s\n", inputPath)
	}

	pkg, err := parser.ParsePackage(inputPath)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Set import path
	if opts.ImportPath != "" {
		pkg.ImportPath = opts.ImportPath
	} else {
		importPath, err := inferImportPath(inputPath)
		if err == nil {
			pkg.ImportPath = importPath
		} else {
			return fmt.Errorf("could not determine import path: use --import-path flag")
		}
	}

	if opts.Verbose {
		fmt.Printf("Package: %s\n", pkg.Name)
		fmt.Printf("Import path: %s\n", pkg.ImportPath)
		fmt.Printf("Functions: %d\n", len(pkg.Functions))
		fmt.Printf("Structs: %d\n", len(pkg.Structs))
	}

	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = filepath.Join(inputPath, opts.Plugin+"_build")
	}
	outputDir, err = filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	// Get the plugin
	plugin, err := factory.Get(opts.Plugin, opts.Verbose)
	if err != nil {
		return fmt.Errorf("unsupported plugin for build: %s (supported: cgo, python)", opts.Plugin)
	}

	// Build using the plugin
	buildOpts := &core.BuildOptions{
		OutputDir:   outputDir,
		LibraryName: opts.LibraryName,
		BuildSystem: opts.BuildSystem,
		Verbose:     opts.Verbose,
	}

	return plugin.Build(pkg, inputPath, buildOpts)
}
