# Web Frontend Development

The OmniGraph web application is implemented in React + TypeScript and built with
Vite.

## Start development server

```bash
cd web
npm ci
npm run dev
```

## Validate frontend changes

```bash
cd web
npm run lint
npm run build
```

## Optional Wasm flow

If you are changing WebAssembly-backed diagnostics, rebuild Wasm assets first, then
run the web app for integration testing.
