.PHONY: build build-windows test vet web-install web-lint web-build wasm-hcldiag wasm-tfpattern-test

build:
	go build -o bin/omnigraph ./cmd/omnigraph

# Windows PE (Git Bash / WSL make). Default amd64; on Windows ARM64: make build-windows GOARCH=arm64
GOARCH ?= amd64
build-windows:
	GOOS=windows GOARCH=$(GOARCH) go build -trimpath -o omnigraph.exe ./cmd/omnigraph

# Build browser HCL diagnostics (requires Go 1.22+). wasm_exec.js is vendored under web/public/wasm/.
wasm-hcldiag:
	cd wasm/hcldiag && GOOS=js GOARCH=wasm go build -trimpath -o ../../web/public/wasm/hcldiag.wasm .

wasm-tfpattern-test:
	cd wasm/tfpattern && go test ./...

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
