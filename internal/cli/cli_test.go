package cli

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Suite")
}

var _ = Describe("CLI", func() {
	Describe("ExitCode", func() {
		It("converts to int correctly", func() {
			Expect(ExitCodeSuccess.ToInt()).To(Equal(0))
			Expect(ExitCodeError.ToInt()).To(Equal(1))
		})
	})

	Describe("NewGoAnywhereCmd", func() {
		It("creates root command", func() {
			cmd := NewGoAnywhereCmd()
			Expect(cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("goanywhere"))
		})

		It("has generate subcommand", func() {
			cmd := NewGoAnywhereCmd()
			generateCmd, _, err := cmd.Find([]string{"generate"})
			Expect(err).NotTo(HaveOccurred())
			Expect(generateCmd.Use).To(ContainSubstring("generate"))
		})

		It("has build subcommand", func() {
			cmd := NewGoAnywhereCmd()
			buildCmd, _, err := cmd.Find([]string{"build"})
			Expect(err).NotTo(HaveOccurred())
			Expect(buildCmd.Use).To(ContainSubstring("build"))
		})

		It("has version flag", func() {
			cmd := NewGoAnywhereCmd()
			Expect(cmd.Version).NotTo(BeEmpty())
		})
	})

	Describe("NewGenerateCmd", func() {
		It("creates generate command with flags", func() {
			cmd := NewGenerateCmd()
			Expect(cmd).NotTo(BeNil())
			Expect(cmd.Use).To(ContainSubstring("generate"))

			// Check flags exist
			Expect(cmd.Flags().Lookup("output")).NotTo(BeNil())
			Expect(cmd.Flags().Lookup("import-path")).NotTo(BeNil())
			Expect(cmd.Flags().Lookup("plugin")).NotTo(BeNil())
			Expect(cmd.Flags().Lookup("verbose")).NotTo(BeNil())
		})

		It("has correct default plugin value", func() {
			cmd := NewGenerateCmd()
			pluginFlag := cmd.Flags().Lookup("plugin")
			Expect(pluginFlag.DefValue).To(Equal("cgo"))
		})
	})

	Describe("NewBuildCmd", func() {
		It("creates build command with flags", func() {
			cmd := NewBuildCmd()
			Expect(cmd).NotTo(BeNil())
			Expect(cmd.Use).To(ContainSubstring("build"))

			// Check flags exist
			Expect(cmd.Flags().Lookup("output")).NotTo(BeNil())
			Expect(cmd.Flags().Lookup("import-path")).NotTo(BeNil())
			Expect(cmd.Flags().Lookup("plugin")).NotTo(BeNil())
			Expect(cmd.Flags().Lookup("verbose")).NotTo(BeNil())
			Expect(cmd.Flags().Lookup("build-system")).NotTo(BeNil())
			Expect(cmd.Flags().Lookup("lib-name")).NotTo(BeNil())
		})

		It("has correct default build-system value", func() {
			cmd := NewBuildCmd()
			buildSystemFlag := cmd.Flags().Lookup("build-system")
			Expect(buildSystemFlag.DefValue).To(Equal("setuptools"))
		})
	})

	Describe("inferImportPath", func() {
		It("returns error for directory without go.mod", func() {
			tmpDir, err := os.MkdirTemp("", "no-gomod")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.RemoveAll(tmpDir) }()

			_, err = inferImportPath(tmpDir)
			Expect(err).To(HaveOccurred())
		})

		It("infers import path from go.mod", func() {
			tmpDir, err := os.MkdirTemp("", "with-gomod")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.RemoveAll(tmpDir) }()

			goModContent := "module github.com/example/mymodule\n\ngo 1.21\n"
			err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			importPath, err := inferImportPath(tmpDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(importPath).To(Equal("github.com/example/mymodule"))
		})

		It("infers import path for subdirectory", func() {
			tmpDir, err := os.MkdirTemp("", "with-gomod")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.RemoveAll(tmpDir) }()

			goModContent := "module github.com/example/mymodule\n\ngo 1.21\n"
			err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			subDir := filepath.Join(tmpDir, "pkg", "mypackage")
			err = os.MkdirAll(subDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			importPath, err := inferImportPath(subDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(importPath).To(Equal("github.com/example/mymodule/pkg/mypackage"))
		})
	})

	Describe("splitLines", func() {
		It("splits string into lines", func() {
			lines := splitLines("line1\nline2\nline3")
			Expect(lines).To(HaveLen(3))
			Expect(lines[0]).To(Equal("line1"))
			Expect(lines[1]).To(Equal("line2"))
			Expect(lines[2]).To(Equal("line3"))
		})

		It("handles string without newlines", func() {
			lines := splitLines("single line")
			Expect(lines).To(HaveLen(1))
			Expect(lines[0]).To(Equal("single line"))
		})

		It("handles empty string", func() {
			lines := splitLines("")
			Expect(lines).To(HaveLen(0))
		})

		It("handles trailing newline", func() {
			lines := splitLines("line1\nline2\n")
			Expect(lines).To(HaveLen(2))
		})
	})

	Describe("runGenerate", func() {
		var fixtureDir string

		BeforeEach(func() {
			wd, _ := os.Getwd()
			fixtureDir = filepath.Join(wd, "..", "..", "tests", "fixtures", "simple")
		})

		It("returns error for non-existent directory", func() {
			opts := &generateOptions{Plugin: "cgo"}
			err := runGenerate("/nonexistent/path", opts)
			Expect(err).To(HaveOccurred())
		})

		It("returns error when path is a file", func() {
			tmpFile, err := os.CreateTemp("", "testfile")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove(tmpFile.Name()) }()
			_ = tmpFile.Close()

			opts := &generateOptions{Plugin: "cgo"}
			err = runGenerate(tmpFile.Name(), opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not a directory"))
		})

		It("returns error for invalid plugin", func() {
			opts := &generateOptions{Plugin: "invalid"}
			err := runGenerate(fixtureDir, opts)
			Expect(err).To(HaveOccurred())
		})

		It("generates CGO code successfully", func() {
			tmpDir, err := os.MkdirTemp("", "output")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.RemoveAll(tmpDir) }()

			opts := &generateOptions{
				Plugin:     "cgo",
				OutputFile: filepath.Join(tmpDir, "main.go"),
				ImportPath: "github.com/test/simple",
			}
			err = runGenerate(fixtureDir, opts)
			Expect(err).NotTo(HaveOccurred())

			// Check output file exists
			_, err = os.Stat(filepath.Join(tmpDir, "main.go"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("generates Python code successfully", func() {
			tmpDir, err := os.MkdirTemp("", "output")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.RemoveAll(tmpDir) }()

			opts := &generateOptions{
				Plugin:     "python",
				OutputFile: filepath.Join(tmpDir, "simple.py"),
				ImportPath: "github.com/test/simple",
			}
			err = runGenerate(fixtureDir, opts)
			Expect(err).NotTo(HaveOccurred())

			// Check output file exists
			_, err = os.Stat(filepath.Join(tmpDir, "simple.py"))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("runBuild", func() {
		var fixtureDir string

		BeforeEach(func() {
			wd, _ := os.Getwd()
			fixtureDir = filepath.Join(wd, "..", "..", "tests", "fixtures", "simple")
		})

		It("returns error for non-existent directory", func() {
			opts := &buildOptions{Plugin: "cgo"}
			err := runBuild("/nonexistent/path", opts)
			Expect(err).To(HaveOccurred())
		})

		It("returns error when path is a file", func() {
			tmpFile, err := os.CreateTemp("", "testfile")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Remove(tmpFile.Name()) }()
			_ = tmpFile.Close()

			opts := &buildOptions{Plugin: "cgo"}
			err = runBuild(tmpFile.Name(), opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not a directory"))
		})

		It("returns error for unsupported plugin", func() {
			opts := &buildOptions{Plugin: "rust"}
			err := runBuild(fixtureDir, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unsupported plugin"))
		})
	})

	Describe("Execute", func() {
		It("returns success for help command", func() {
			// Save original args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = []string{"goanywhere", "--help"}
			exitCode := Execute()
			Expect(exitCode).To(Equal(ExitCodeSuccess))
		})

		It("returns success for version command", func() {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = []string{"goanywhere", "--version"}
			exitCode := Execute()
			Expect(exitCode).To(Equal(ExitCodeSuccess))
		})

		It("returns error for invalid command", func() {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = []string{"goanywhere", "invalid-command"}
			exitCode := Execute()
			Expect(exitCode).To(Equal(ExitCodeError))
		})
	})

	Describe("runGenerate verbose mode", func() {
		var fixtureDir string

		BeforeEach(func() {
			wd, _ := os.Getwd()
			fixtureDir = filepath.Join(wd, "..", "..", "tests", "fixtures", "simple")
		})

		It("generates with verbose output", func() {
			tmpDir, err := os.MkdirTemp("", "output")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.RemoveAll(tmpDir) }()

			opts := &generateOptions{
				Plugin:     "cgo",
				OutputFile: filepath.Join(tmpDir, "main.go"),
				ImportPath: "github.com/test/simple",
				Verbose:    true,
			}
			err = runGenerate(fixtureDir, opts)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("runBuild verbose mode", func() {
		var fixtureDir string

		BeforeEach(func() {
			wd, _ := os.Getwd()
			fixtureDir = filepath.Join(wd, "..", "..", "tests", "fixtures", "simple")
		})

		It("runs with verbose option", func() {
			tmpDir, err := os.MkdirTemp("", "build-output")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.RemoveAll(tmpDir) }()

			opts := &buildOptions{
				Plugin:     "cgo",
				OutputDir:  tmpDir,
				ImportPath: "github.com/test/simple",
				Verbose:    true,
			}
			// This will fail at the CGO compilation step (no C compiler),
			// but it will exercise the code path up to that point
			_ = runBuild(fixtureDir, opts)
		})
	})
})
