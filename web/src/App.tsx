import { useEffect, useMemo, useState } from 'react'
import { GraphCanvas } from './graph/GraphCanvas'
import sampleGraph from './graph/sampleGraph.json'
import { formatHclDiagnostics, initHclWasm, validateHclText } from './hclWasm'
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

const defaultGraphJson = JSON.stringify(sampleGraph, null, 2)

const defaultHcl = `resource "null_resource" "example" {
  triggers = {
    always = timestamp()
  }
}
`

function App() {
  const [text, setText] = useState(defaultSchema)
  const [debounced, setDebounced] = useState(text)
  const [graphJson, setGraphJson] = useState(defaultGraphJson)
  const [hclText, setHclText] = useState(defaultHcl)
  const [hclDebounced, setHclDebounced] = useState(hclText)
  const [wasmNote, setWasmNote] = useState<string | null>(null)
  const [hclWasm, setHclWasm] = useState<'loading' | 'ok' | 'err'>('loading')

  useEffect(() => {
    const t = setTimeout(() => setDebounced(text), 250)
    return () => clearTimeout(t)
  }, [text])

  useEffect(() => {
    const t = setTimeout(() => setHclDebounced(hclText), 300)
    return () => clearTimeout(t)
  }, [hclText])

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

  useEffect(() => {
    initHclWasm()
      .then(() => setHclWasm('ok'))
      .catch(() => setHclWasm('err'))
  }, [])

  const hclPanel = useMemo(() => {
    if (hclWasm === 'loading') {
      return { tone: 'muted' as const, text: 'Loading HCL Wasm…' }
    }
    if (hclWasm === 'err') {
      return {
        tone: 'muted' as const,
        text: 'HCL Wasm unavailable (run `make wasm-hcldiag` and ensure web/public/wasm/hcldiag.wasm exists).',
      }
    }
    try {
      const diags = validateHclText(hclDebounced)
      return {
        tone: diags.length === 0 ? ('ok' as const) : ('warn' as const),
        text: formatHclDiagnostics(diags),
      }
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      return { tone: 'err' as const, text: `HCL validation error: ${m}` }
    }
  }, [hclWasm, hclDebounced])

  return (
    <div className="flex min-h-screen flex-col bg-slate-950 px-6 py-10 text-slate-100">
      <header className="mx-auto w-full max-w-6xl">
        <h1 className="text-3xl font-semibold tracking-tight text-white">OmniGraph</h1>
        <p className="mt-2 text-sm leading-relaxed text-slate-400">
          State-aware DevSecOps orchestration: OpenTofu, Ansible, and telemetry in one GitOps flow.
          Edit <code className="text-slate-300">.omnigraph.schema</code> — validation runs locally
          (JSON Schema draft 2020-12). Paste <code className="text-slate-300">omnigraph graph emit</code> output
          below to explore the blast-radius graph. HCL snippets use Hashicorp HCL parse diagnostics in Wasm (ADR 001
          phase 1).
        </p>
      </header>

      <main className="mx-auto mt-8 flex w-full max-w-6xl flex-1 flex-col gap-8">
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

        <section className="flex flex-col gap-2">
          <label htmlFor="hcl" className="text-sm font-medium text-slate-300">
            Terraform / HCL snippet (Wasm diagnostics)
          </label>
          <textarea
            id="hcl"
            className="min-h-40 w-full resize-y rounded-lg border border-slate-700 bg-slate-900/80 p-3 font-mono text-sm text-slate-100 outline-none ring-violet-500/40 focus:ring-2"
            spellCheck={false}
            value={hclText}
            onChange={(e) => setHclText(e.target.value)}
            aria-describedby="hcl-status"
          />
          <div
            id="hcl-status"
            className={`rounded-md border px-3 py-2 text-sm whitespace-pre-wrap ${
              hclPanel.tone === 'ok'
                ? 'border-emerald-800 bg-emerald-950/40 text-emerald-200'
                : hclPanel.tone === 'warn'
                  ? 'border-amber-800 bg-amber-950/40 text-amber-100'
                  : hclPanel.tone === 'err'
                    ? 'border-rose-800 bg-rose-950/40 text-rose-200'
                    : 'border-slate-700 bg-slate-900/60 text-slate-400'
            }`}
            role="status"
          >
            {hclPanel.text}
          </div>
        </section>

        {wasmNote ? (
          <p className="text-xs text-slate-500" data-testid="wasm-spike-note">
            {wasmNote}
          </p>
        ) : null}

        <section className="flex flex-col gap-2">
          <label htmlFor="graph-json" className="text-sm font-medium text-slate-300">
            Graph JSON (<span className="text-slate-500">omnigraph/graph/v1</span>)
          </label>
          <textarea
            id="graph-json"
            className="min-h-40 w-full resize-y rounded-lg border border-slate-700 bg-slate-900/80 p-3 font-mono text-xs text-slate-100 outline-none ring-sky-500/40 focus:ring-2"
            spellCheck={false}
            value={graphJson}
            onChange={(e) => setGraphJson(e.target.value)}
            aria-label="Omnigraph graph v1 JSON"
          />
          <GraphCanvas graphText={graphJson} />
        </section>
      </main>
    </div>
  )
}

export default App
