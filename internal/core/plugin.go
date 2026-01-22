// Copyright 2026 Riley Rice
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
