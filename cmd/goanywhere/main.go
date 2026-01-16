package main

import (
	"os"
	"github.com/riceriley59/goanywhere/internal/cli"
)

func main() {
	exitCode := cli.Execute()
	os.Exit(exitCode.ToInt())
}
