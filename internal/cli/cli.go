package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/riceriley59/goanywhere/internal/version"
)

type ExitCode int

const (
	ExitCodeSuccess ExitCode = 0
	ExitCodeError   ExitCode = 1
)

func (e ExitCode) ToInt() int {
	return int(e)
}

func NewGoAnywhereCmd() *cobra.Command {
	goAnywhereCmd := &cobra.Command{
		Use:           "goanywhere",
		Short:         "Go bindings generator",
		Long:          "Go bindings generator for multi-language support",
		Version:       version.GetVersion(),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Add subcommands
	goAnywhereCmd.AddCommand(NewGenerateCmd())
	goAnywhereCmd.AddCommand(NewBuildCmd())

	return goAnywhereCmd
}

func Execute() ExitCode {
	if err := NewGoAnywhereCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return ExitCodeError
	}

	return ExitCodeSuccess
}
