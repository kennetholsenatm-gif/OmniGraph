.PHONY: build test vet web-install web-lint web-build

build:
	go build -o bin/omnigraph ./cmd/omnigraph

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
