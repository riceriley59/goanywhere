package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/riceriley59/goanywhere/internal/core"
	"github.com/riceriley59/goanywhere/internal/core/factory"

	// Register plugins
	_ "github.com/riceriley59/goanywhere/plugins/cgo"
	_ "github.com/riceriley59/goanywhere/plugins/python"
)

type generateOptions struct {
	OutputFile string
	ImportPath string
	Plugin     string
	Verbose    bool
}

// NewGenerateCmd creates the generate subcommand
func NewGenerateCmd() *cobra.Command {
	opts := &generateOptions{}

	// Build supported plugins list for help text
	pluginList := strings.Join(factory.List(), ", ")

	cmd := &cobra.Command{
		Use:   "generate <input-directory>",
		Short: "Generate plugin code for a Go package",
		Long: fmt.Sprintf(`Generate plugin code that exposes Go functions and structs to other languages.

The generator processes all .go files in the specified directory and creates
plugin code for the specified target language.

Supported plugins: %s

Example:
  goanywhere generate ./mypackage -o plugin.go
  goanywhere generate ./mypackage --import-path github.com/user/mypackage
  goanywhere generate ./mypackage --plugin cgo`, pluginList),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(args[0], opts)
		},
	}

	cmd.Flags().StringVarP(&opts.OutputFile, "output", "o", "",
		"Output file path (default: <input>/<plugin>_plugin/main.go)")
	cmd.Flags().StringVarP(&opts.ImportPath, "import-path", "i", "",
		"Import path for the target package (required for proper imports)")
	cmd.Flags().StringVarP(&opts.Plugin, "plugin", "p", "cgo",
		fmt.Sprintf("Plugin type to generate (%s)", pluginList))
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false,
		"Verbose output showing parsed constructs and skipped items")

	return cmd
}

func runGenerate(inputDir string, opts *generateOptions) error {
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

	// Get the plugin from factory
	plugin, err := factory.Get(opts.Plugin, opts.Verbose)
	if err != nil {
		return err
	}

	// Create parser and parse package
	parser := core.NewParser(opts.Verbose)

	if opts.Verbose {
		fmt.Printf("Parsing package at: %s\n", inputPath)
	}

	pkg, err := parser.ParsePackage(inputPath)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Set import path if provided
	if opts.ImportPath != "" {
		pkg.ImportPath = opts.ImportPath
	} else {
		// Try to infer import path from go.mod
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
		fmt.Printf("Plugin: %s\n", plugin.Name())
		fmt.Printf("Functions: %d\n", len(pkg.Functions))
		fmt.Printf("Structs: %d\n", len(pkg.Structs))
		for _, fn := range pkg.Functions {
			fmt.Printf("  - Function: %s\n", fn.Name)
		}
		for _, st := range pkg.Structs {
			fmt.Printf("  - Struct: %s (%d methods)\n", st.Name, len(st.Methods))
		}
	}

	// Generate plugin code using the plugin interface
	code, err := plugin.Generate(pkg)
	if err != nil {
		return fmt.Errorf("generation error: %w", err)
	}

	// Determine output path
	outputPath := opts.OutputFile
	if outputPath == "" {
		// Default to a subdirectory based on plugin type
		switch plugin.Name() {
		case "python":
			outputPath = filepath.Join(inputPath, plugin.Name()+"_plugin", pkg.Name+".py")
		default:
			outputPath = filepath.Join(inputPath, plugin.Name()+"_plugin", "main.go")
		}
	}

	// Make output path absolute
	outputPath, err = filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputPath, code, 0644); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	fmt.Printf("Generated %s plugin: %s\n", plugin.Name(), outputPath)

	// Print build instructions based on plugin type
	if plugin.Name() == "cgo" {
		fmt.Println("\nTo build as shared library:")
		fmt.Printf("  CGO_ENABLED=1 go build -buildmode=c-shared -o lib%s.so %s\n", pkg.Name, outputPath)
	}

	return nil
}

// inferImportPath tries to determine the import path from go.mod
func inferImportPath(pkgDir string) (string, error) {
	// Walk up to find go.mod
	dir := pkgDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod, read module name
			content, err := os.ReadFile(goModPath)
			if err != nil {
				return "", err
			}

			// Parse module line
			moduleName := ""
			lines := splitLines(string(content))
			for _, line := range lines {
				if len(line) > 7 && line[:7] == "module " {
					moduleName = line[7:]
					break
				}
			}

			if moduleName != "" {
				// Calculate relative path from module root to package
				relPath, err := filepath.Rel(dir, pkgDir)
				if err != nil {
					return moduleName, nil
				}
				if relPath == "." {
					return moduleName, nil
				}
				return filepath.ToSlash(filepath.Join(moduleName, relPath)), nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("go.mod not found")
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
