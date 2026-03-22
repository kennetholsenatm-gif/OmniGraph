package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kennetholsenatm-gif/omnigraph/internal/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "OmniGraph control plane — orchestrates provisioning, configuration, and handoff.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [flags]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println(version.String())
		return
	}

	if flag.NArg() > 0 {
		flag.Usage()
		os.Exit(2)
	}

	fmt.Println("omnigraph: no command yet; use -version or -h")
}
