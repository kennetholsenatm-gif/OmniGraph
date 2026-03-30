/// <reference types="vite/client" />

declare const __OMNIGRAPH_WEB_VERSION__: string

declare module '*?raw' {
  const src: string
  export default src
}

interface ImportMetaEnv {
  readonly VITE_ENABLE_WASM_SPIKE?: string
  /** When the UI is not same-origin with the workspace server, set to e.g. http://127.0.0.1:38671 */
  readonly VITE_OMNIGRAPH_API?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

/** File System Access API (Chromium) — not in all TypeScript DOM libs */
interface FileSystemDirectoryHandle {
  values(): AsyncIterableIterator<FileSystemHandle>
}

interface Window {
  showDirectoryPicker?(options?: {
    id?: string
    mode?: 'read' | 'readwrite'
    startIn?: FileSystemHandle | FileSystemDirectoryHandle
  }): Promise<FileSystemDirectoryHandle>
}
