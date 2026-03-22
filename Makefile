.PHONY: build test vet web-install web-lint web-build wasm-hcldiag

build:
	go build -o bin/omnigraph ./cmd/omnigraph

# Build browser HCL diagnostics (requires Go 1.22+). wasm_exec.js is vendored under web/public/wasm/.
wasm-hcldiag:
	cd wasm/hcldiag && GOOS=js GOARCH=wasm go build -trimpath -o ../../web/public/wasm/hcldiag.wasm .

test:
	go test ./...

vet:
	go vet ./...

web-install:
	cd web && npm ci

web-lint:
	cd web && npm run lint

web-build:
	cd web && npm run build
