import { useEffect, useMemo, useState } from 'react'
import { validateOmnigraphText } from './validateOmnigraph'
import { wasmSpikeAdd, wasmSpikeEnabled } from './wasmSpike'

const defaultSchema = `apiVersion: omnigraph/v1alpha1
kind: Project
metadata:
  name: demo
  environment: staging
spec:
  network:
    vpcCidr: 10.0.0.0/16
    publicPorts: [80, 443]
  tags:
    app: web
`

function App() {
  const [text, setText] = useState(defaultSchema)
  const [debounced, setDebounced] = useState(text)
  const [wasmNote, setWasmNote] = useState<string | null>(null)

  useEffect(() => {
    const t = setTimeout(() => setDebounced(text), 250)
    return () => clearTimeout(t)
  }, [text])

  const validation = useMemo(() => validateOmnigraphText(debounced), [debounced])

  useEffect(() => {
    if (!wasmSpikeEnabled()) {
      return
    }
    let cancelled = false
    wasmSpikeAdd(2, 3)
      .then((n) => {
        if (!cancelled) {
          setWasmNote(`Wasm spike: add(2,3) = ${n} (feature flag VITE_ENABLE_WASM_SPIKE)`)
        }
      })
      .catch((e: unknown) => {
        if (!cancelled) {
          const m = e instanceof Error ? e.message : String(e)
          setWasmNote(`Wasm spike failed: ${m}`)
        }
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="flex min-h-screen flex-col bg-slate-950 px-6 py-10 text-slate-100">
      <header className="mx-auto w-full max-w-3xl">
        <h1 className="text-3xl font-semibold tracking-tight text-white">OmniGraph</h1>
        <p className="mt-2 text-sm leading-relaxed text-slate-400">
          State-aware DevSecOps orchestration: OpenTofu, Ansible, and telemetry in one GitOps flow.
          Edit <code className="text-slate-300">.omnigraph.schema</code> — validation runs locally
          (JSON Schema draft 2020-12).
        </p>
      </header>

      <main className="mx-auto mt-8 flex w-full max-w-3xl flex-1 flex-col gap-6">
        <section className="flex flex-col gap-2">
          <label htmlFor="schema" className="text-sm font-medium text-slate-300">
            Project schema
          </label>
          <textarea
            id="schema"
            className="min-h-56 w-full resize-y rounded-lg border border-slate-700 bg-slate-900/80 p-3 font-mono text-sm text-slate-100 outline-none ring-emerald-500/40 focus:ring-2"
            spellCheck={false}
            value={text}
            onChange={(e) => setText(e.target.value)}
            aria-describedby="schema-status"
          />
          <div
            id="schema-status"
            className={`rounded-md border px-3 py-2 text-sm ${
              validation.ok
                ? 'border-emerald-800 bg-emerald-950/40 text-emerald-200'
                : 'border-rose-800 bg-rose-950/40 text-rose-200'
            }`}
            role="status"
          >
            {validation.ok ? 'Schema valid.' : validation.message}
          </div>
        </section>

        {wasmNote ? (
          <p className="text-xs text-slate-500" data-testid="wasm-spike-note">
            {wasmNote}
          </p>
        ) : null}

        <section
          className="flex h-56 w-full items-center justify-center rounded-xl border border-dashed border-slate-600 bg-slate-900/50 text-sm text-slate-500"
          role="img"
          aria-label="Dependency graph placeholder"
        >
          Graph canvas placeholder (D3.js / SVG)
        </section>
      </main>
    </div>
  )
}

export default App
