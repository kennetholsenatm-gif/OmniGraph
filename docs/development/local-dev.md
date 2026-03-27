# Local Development

## Prerequisites

- Go `1.23+`
- Node.js `20+`
- Git

## Build and test CLI

```bash
go vet ./...
go test ./...
go build -o bin/omnigraph ./cmd/omnigraph
./bin/omnigraph --help
```

## Run web frontend

```bash
cd web
npm ci
npm run dev
```

## Suggested verification

```bash
cd web
npm run lint
npm run build
```
