import { Check, CheckCircle, Code, ShieldAlert, TerminalSquare } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'

import {
  formatHclDiagnostics,
  formatPatternDiagnostics,
  lintTfPatterns,
  validateHclText,
} from '../hclWasm'
type WasmStatus = 'loading' | 'ok' | 'err'

type DemoScenario = {
  title: string
  toolLabel: string
  file: string
  badCode: string
  goodCode: string
  errorMsg: string
  errorLine: number
}

const demoScenarios: Record<string, DemoScenario> = {
  checkov: {
    title: 'Security flaw (illustrative)',
    toolLabel: 'Checkov (not shipped in Wasm)',
    file: 'main.tf',
    badCode: `resource "aws_security_group" "web" {
  name        = "web-sg"
  description = "Allow SSH"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # TODO: Restrict
  }
}`,
    goodCode: `resource "aws_security_group" "web" {
  name        = "web-sg"
  description = "Allow SSH"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }
}`,
    errorMsg: 'CKV_AWS_24: Ensure no security groups allow ingress from 0.0.0.0:0 to port 22 (demo only).',
    errorLine: 8,
  },
  tflint: {
    title: 'Invalid instance type (illustrative)',
    toolLabel: 'TFLint (not shipped in Wasm)',
    file: 'compute.tf',
    badCode: `resource "aws_instance" "app" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t1.supermicro"

  tags = {
    Name = "AppServer"
  }
}`,
    goodCode: `resource "aws_instance" "app" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t3.micro"

  tags = {
    Name = "AppServer"
  }
}`,
    errorMsg: 'Aws_instance invalid type (demo only).',
    errorLine: 3,
  },
}

export type WebIDETabProps = {
  hclWasm: WasmStatus
  hclText: string
  onHclChange: (value: string) => void
}

export function WebIDETab({ hclWasm, hclText, onHclChange }: WebIDETabProps) {
  const [mode, setMode] = useState<'live' | 'demo'>('live')
  const [hclDebounced, setHclDebounced] = useState(hclText)
  const [demoKey, setDemoKey] = useState<'checkov' | 'tflint'>('checkov')
  const [demoFixed, setDemoFixed] = useState(false)

  useEffect(() => {
    const t = setTimeout(() => setHclDebounced(hclText), 300)
    return () => clearTimeout(t)
  }, [hclText])

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

  const patternPanel = useMemo(() => {
    if (hclWasm !== 'ok') {
      return { tone: 'muted' as const, text: 'Pattern scan loads with HCL Wasm.' }
    }
    try {
      const diags = lintTfPatterns(hclDebounced)
      const hasErr = diags.some((d) => d.severity === 'error')
      const hasWarn = diags.some((d) => d.severity === 'warning')
      const tone = hasErr ? ('err' as const) : hasWarn ? ('warn' as const) : ('ok' as const)
      return { tone, text: formatPatternDiagnostics(diags) }
    } catch (e) {
      const m = e instanceof Error ? e.message : String(e)
      return { tone: 'err' as const, text: `Pattern scan error: ${m}` }
    }
  }, [hclWasm, hclDebounced])

  const activeDemo = demoScenarios[demoKey]

  return (
    <div className="flex h-full min-h-0 w-full flex-col lg:flex-row">
      <div className="flex min-h-0 min-w-0 flex-1 flex-col border-b border-gray-800 lg:border-b-0 lg:border-r">
        <div className="flex h-12 items-center justify-between border-b border-gray-800 bg-gray-900 px-4">
          <div className="flex items-center gap-2 rounded border border-gray-800 bg-gray-950 px-3 py-1 text-sm text-gray-300">
            <Code size={14} className="text-purple-400" aria-hidden />
            {mode === 'live' ? 'snippet.tf' : activeDemo.file}
          </div>
          <div className="flex items-center gap-2">
            <select
              value={mode}
              onChange={(e) => {
                const m = e.target.value as 'live' | 'demo'
                setMode(m)
                if (m === 'demo') {
                  setDemoFixed(false)
                }
              }}
              className="rounded-md border border-gray-800 bg-gray-950 px-2 py-1 text-sm text-gray-300 focus:outline-none focus:ring-1 focus:ring-blue-500"
              aria-label="Editor mode"
            >
              <option value="live">Live HCL (Wasm)</option>
              <option value="demo">Guided demo (not Wasm)</option>
            </select>
            {mode === 'demo' ? (
              <select
                value={demoKey}
                onChange={(e) => {
                  setDemoKey(e.target.value as 'checkov' | 'tflint')
                  setDemoFixed(false)
                }}
                className="rounded-md border border-gray-800 bg-gray-950 px-2 py-1 text-sm text-gray-300 focus:outline-none focus:ring-1 focus:ring-blue-500"
                aria-label="Demo scenario"
              >
                <option value="checkov">Scenario: security group</option>
                <option value="tflint">Scenario: instance type</option>
              </select>
            ) : null}
          </div>
        </div>

        {mode === 'live' ? (
          <textarea
            spellCheck={false}
            value={hclText}
            onChange={(e) => onHclChange(e.target.value)}
            className="min-h-64 flex-1 resize-none bg-[#1e1e1e] p-4 font-mono text-sm leading-loose text-gray-100 outline-none"
            aria-label="HCL editor"
          />
        ) : (
          <div className="flex-1 overflow-auto bg-[#1e1e1e] p-4 font-mono text-sm leading-loose">
            {(demoFixed ? activeDemo.goodCode : activeDemo.badCode).split('\n').map((line, i) => {
              const isErrorLine = !demoFixed && i + 1 === activeDemo.errorLine
              return (
                <div
                  key={i}
                  className={`flex ${isErrorLine ? '-mx-4 border-l-2 border-rose-500 bg-rose-500/10 px-4' : ''}`}
                >
                  <span className="inline-block w-8 flex-none select-none text-gray-600">{i + 1}</span>
                  <span
                    className={
                      isErrorLine
                        ? 'text-rose-200 underline decoration-rose-500 decoration-wavy underline-offset-4'
                        : 'text-gray-300'
                    }
                  >
                    {line}
                  </span>
                </div>
              )
            })}
          </div>
        )}
      </div>

      <aside className="flex w-full shrink-0 flex-col bg-gray-900/50 p-6 lg:w-80">
        <h2 className="mb-6 flex items-center gap-2 text-lg font-bold text-gray-100">
          <TerminalSquare className="text-blue-400" size={20} aria-hidden />
          Diagnostics
        </h2>

        {mode === 'live' ? (
          <div className="space-y-4">
            <div
              className={`rounded-xl border px-3 py-2 text-sm whitespace-pre-wrap ${
                hclPanel.tone === 'ok'
                  ? 'border-emerald-800 bg-emerald-950/40 text-emerald-200'
                  : hclPanel.tone === 'warn'
                    ? 'border-amber-800 bg-amber-950/40 text-amber-100'
                    : hclPanel.tone === 'err'
                      ? 'border-rose-800 bg-rose-950/40 text-rose-200'
                      : 'border-gray-700 bg-gray-900/60 text-gray-400'
              }`}
            >
              <span className="mb-1 block text-xs font-semibold uppercase tracking-wide text-gray-500">HCL parse</span>
              {hclPanel.text}
            </div>
            <div
              className={`rounded-xl border px-3 py-2 text-sm whitespace-pre-wrap ${
                patternPanel.tone === 'ok'
                  ? 'border-emerald-800/60 bg-emerald-950/20 text-emerald-200/90'
                  : patternPanel.tone === 'warn'
                    ? 'border-amber-800 bg-amber-950/40 text-amber-100'
                    : patternPanel.tone === 'err'
                      ? 'border-rose-800 bg-rose-950/40 text-rose-200'
                      : 'border-gray-700 bg-gray-900/60 text-gray-400'
              }`}
            >
              <span className="mb-1 block text-xs font-semibold uppercase tracking-wide text-gray-500">
                Pattern scan (Wasm)
              </span>
              {patternPanel.text}
            </div>
          </div>
        ) : !demoFixed ? (
          <div className="rounded-xl border border-rose-500/20 bg-rose-500/10 p-4 transition-opacity duration-300">
            <div className="mb-2 flex items-center gap-2 text-sm font-bold uppercase tracking-wide text-rose-400">
              <ShieldAlert size={16} aria-hidden />
              {activeDemo.toolLabel}
            </div>
            <p className="mb-2 text-sm text-gray-300">{activeDemo.title}</p>
            <p className="mb-4 text-sm text-gray-300">{activeDemo.errorMsg}</p>
            <p className="mb-4 font-mono text-xs text-gray-500">
              Line {activeDemo.errorLine} · {activeDemo.file}
            </p>
            <button
              type="button"
              onClick={() => setDemoFixed(true)}
              className="flex w-full items-center justify-center gap-2 rounded-lg border border-rose-500/50 bg-rose-500/20 py-2 text-sm font-medium text-rose-300 transition-colors hover:bg-rose-500/30"
            >
              <Check size={16} aria-hidden />
              Apply demo fix
            </button>
          </div>
        ) : (
          <div className="rounded-xl border border-emerald-500/20 bg-emerald-500/10 p-4 transition-opacity duration-300">
            <div className="mb-2 flex items-center gap-2 text-sm font-bold uppercase tracking-wide text-emerald-400">
              <CheckCircle size={16} aria-hidden />
              Demo resolved
            </div>
            <p className="text-sm text-gray-400">Illustrative only — full Checkov/TFLint are not in the browser yet (ADR 001).</p>
          </div>
        )}

        <div className="mt-auto space-y-2 border-t border-gray-800 pt-6 text-xs text-gray-500">
          <p className="flex items-center justify-between">
            <span>HCL Wasm</span>
            <span className={hclWasm === 'ok' ? 'text-emerald-400' : hclWasm === 'err' ? 'text-rose-400' : 'text-amber-400'}>
              {hclWasm === 'ok' ? 'Ready' : hclWasm === 'err' ? 'Unavailable' : 'Loading'}
            </span>
          </p>
        </div>
      </aside>
    </div>
  )
}
