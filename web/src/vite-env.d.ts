/// <reference types="vite/client" />

declare module '*?raw' {
  const src: string
  export default src
}

interface ImportMetaEnv {
  readonly VITE_ENABLE_WASM_SPIKE?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
