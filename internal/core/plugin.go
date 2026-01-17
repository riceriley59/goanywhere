package core

// BuildOptions contains configuration for the Build method
type BuildOptions struct {
	// OutputDir is the directory where build artifacts should be placed
	OutputDir string
	// LibraryName overrides the default library name (default: lib<package>)
	LibraryName string
	// BuildSystem specifies the build system for language-specific packaging (e.g., setuptools, hatch)
	BuildSystem string
	// Verbose enables verbose output during build
	Verbose bool
}

// Plugin is the interface that all language plugins must implement
type Plugin interface {
	// Name returns the plugin name (e.g., "cgo", "python", "rust")
	Name() string

	// Generate produces plugin code for the given parsed package
	Generate(pkg *ParsedPackage) ([]byte, error)

	// Build generates code and compiles/packages it for distribution
	// The inputPath is the path to the original Go package source
	Build(pkg *ParsedPackage, inputPath string, opts *BuildOptions) error
}
