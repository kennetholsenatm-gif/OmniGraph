//go:build wasip1

// Package main builds to a WASI module: read Ansible INI from stdin, write omnigraph/graph/v1 JSON to stdout.
package main

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/inventory"
	"github.com/kennetholsenatm-gif/omnigraph/internal/repo"
)

func main() {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}
	hosts, err := inventory.ParseAnsibleINIHosts(b)
	if err != nil {
		os.Exit(2)
	}
	doc := repo.GraphV1FromAnsibleHosts(time.Now().UTC().Format(time.RFC3339), hosts)
	out, err := json.Marshal(doc)
	if err != nil {
		os.Exit(3)
	}
	if _, err := os.Stdout.Write(out); err != nil {
		os.Exit(4)
	}
}
