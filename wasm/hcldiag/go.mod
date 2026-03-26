module github.com/kennetholsenatm-gif/omnigraph/wasm/hcldiag

go 1.22

require (
	github.com/hashicorp/hcl/v2 v2.20.1
	github.com/kennetholsenatm-gif/omnigraph/wasm/tfpattern v0.0.0
)

replace github.com/kennetholsenatm-gif/omnigraph/wasm/tfpattern => ../tfpattern
