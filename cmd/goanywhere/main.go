package main

import (
	"github.com/riceriley59/goanywhere/internal/cli"
	"os"
)

func main() {
	exitCode := cli.Execute()
	os.Exit(exitCode.ToInt())
}
