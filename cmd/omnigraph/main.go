package main

import (
	"os"

	"github.com/kennetholsenatm-gif/omnigraph/internal/cli"
)

func main() {
	// Legacy: `omnigraph -version` before Cobra (single-dash flag).
	if len(os.Args) == 2 && (os.Args[1] == "-version" || os.Args[1] == "--version") {
		os.Args[1] = "--version"
	}
	cli.Execute()
}
