package core

// Plugin is the interface that all language plugins must implement
type Plugin interface {
	// Name returns the plugin name (e.g., "cgo", "python", "rust")
	Name() string

	// Generate produces plugin code for the given parsed package
	Generate(pkg *ParsedPackage) ([]byte, error)
}
