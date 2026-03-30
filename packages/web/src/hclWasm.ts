export type HclDiagnostic = {
  severity: string
  summary: string
  detail?: string
  line?: number
  column?: number
}

type GoConstructor = new () => {
  importObject: WebAssembly.Imports
  run: (instance: WebAssembly.Instance) => void
}

function loadWasmExecScript(): Promise<void> {
  const w = globalThis as unknown as { Go?: GoConstructor }
  if (w.Go) {
    return Promise.resolve()
  }
  return new Promise((resolve, reject) => {
    const s = document.createElement('script')
    s.src = '/wasm/wasm_exec.js'
    s.async = true
    s.onload = () => resolve()
    s.onerror = () => reject(new Error('failed to load /wasm/wasm_exec.js'))
    document.head.appendChild(s)
  })
}

let initPromise: Promise<void> | null = null

async function waitForWasmExports(timeoutMs: number): Promise<void> {
  const t0 = Date.now()
  while (Date.now() - t0 < timeoutMs) {
    const w = globalThis as unknown as {
      omnigraphHclValidate?: unknown
      omnigraphTfPatternLint?: unknown
    }
    if (typeof w.omnigraphHclValidate === 'function' && typeof w.omnigraphTfPatternLint === 'function') {
      return
    }
    await new Promise((r) => setTimeout(r, 20))
  }
  throw new Error('timeout waiting for Wasm exports (hcldiag main did not register)')
}

/** Loads wasm_exec.js and hcldiag.wasm; registers omnigraphHclValidate on globalThis. */
export function initHclWasm(): Promise<void> {
  if (!initPromise) {
    initPromise = (async () => {
      await loadWasmExecScript()
      const w = globalThis as unknown as { Go: GoConstructor }
      const Go = w.Go
      if (!Go) {
        throw new Error('Go runtime missing after wasm_exec.js')
      }
      const go = new Go()
      const res = await WebAssembly.instantiateStreaming(fetch('/wasm/hcldiag.wasm'), go.importObject)
      void go.run(res.instance)
      await waitForWasmExports(15000)
    })()
  }
  return initPromise
}

export function validateHclText(src: string): HclDiagnostic[] {
  try {
    const fn = (globalThis as unknown as { omnigraphHclValidate?: (s: string) => string }).omnigraphHclValidate
    if (typeof fn !== 'function') {
      throw new Error('omnigraphHclValidate not registered; call initHclWasm() first')
    }
    const raw = fn(src)
    return JSON.parse(raw) as HclDiagnostic[]
  } catch (e) {
    return [
      {
        severity: 'error',
        summary: 'HCL Wasm bridge error',
        detail: e instanceof Error ? e.message : String(e),
      },
    ]
  }
}

export function formatHclDiagnostics(ds: HclDiagnostic[]): string {
  if (ds.length === 0) {
    return 'HCL parse OK.'
  }
  return ds
    .map((d) => {
      const loc = d.line ? `line ${d.line}:${d.column ?? 0}` : 'input'
      return `[${d.severity}] ${loc}: ${d.summary}${d.detail ? ` — ${d.detail}` : ''}`
    })
    .join('\n')
}

/** Policy-style pattern scan (same Wasm binary as HCL parse; checkov-style subset). */
export function lintTfPatterns(src: string): HclDiagnostic[] {
  try {
    const fn = (globalThis as unknown as { omnigraphTfPatternLint?: (s: string) => string }).omnigraphTfPatternLint
    if (typeof fn !== 'function') {
      throw new Error('omnigraphTfPatternLint not registered; call initHclWasm() first')
    }
    const raw = fn(src)
    return JSON.parse(raw) as HclDiagnostic[]
  } catch (e) {
    return [
      {
        severity: 'error',
        summary: 'pattern lint bridge error',
        detail: e instanceof Error ? e.message : String(e),
      },
    ]
  }
}

export function formatPatternDiagnostics(ds: HclDiagnostic[]): string {
  if (ds.length === 0) {
    return 'No plaintext secret patterns detected (subset rules).'
  }
  return ds
    .map((d) => {
      const loc = d.line ? `line ${d.line}` : 'input'
      return `[${d.severity}] ${loc}: ${d.summary}${d.detail ? ` — ${d.detail}` : ''}`
    })
    .join('\n')
}
