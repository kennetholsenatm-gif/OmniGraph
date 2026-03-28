import type { ChangeEvent } from 'react'
import { useCallback, useMemo, useState } from 'react'

import { fetchLocalSecurityScan } from './omnigraphApi'

export type SecurityScanV1 = {
  apiVersion: string
  kind: string
  metadata: {
    generatedAt: string
    target?: string
    ansibleHost?: string
    transport?: string
    profile?: string
    disclaimer?: string
  }
  spec: {
    summary: {
      modulesRun: number
      vulnerable: number
      notVulnerable: number
      errors: number
      notApplicable: number
    }
    results: SecurityModuleResult[]
  }
}

export type SecurityModuleResult = {
  moduleId: string
  techniqueId: string
  techniqueName: string
  tactic: string
  severity: string
  status: string
  summary: string
  evidence?: string
  remediation?: string
  complianceTags?: string[]
}

function parseSecurityDoc(raw: string): { ok: true; doc: SecurityScanV1 } | { ok: false; err: string } {
  try {
    const j: unknown = JSON.parse(raw)
    if (!j || typeof j !== 'object') {
      return { ok: false, err: 'Document must be a JSON object' }
    }
    const o = j as Record<string, unknown>
    if (o.apiVersion !== 'omnigraph/security/v1' || o.kind !== 'SecurityScan') {
      return { ok: false, err: 'Expected omnigraph/security/v1 SecurityScan' }
    }
    return { ok: true, doc: j as SecurityScanV1 }
  } catch (e: unknown) {
    const m = e instanceof Error ? e.message : String(e)
    return { ok: false, err: m }
  }
}

function severityClass(s: string): string {
  switch (s) {
    case 'critical':
      return 'text-rose-400'
    case 'high':
      return 'text-orange-400'
    case 'medium':
      return 'text-amber-400'
    case 'low':
      return 'text-yellow-400'
    default:
      return 'text-gray-400'
  }
}

export type PostureTabProps = {
  securityJsonText: string
  onSecurityJsonTextChange: (s: string) => void
}

export function PostureTab(p: PostureTabProps) {
  const [tacticFilter, setTacticFilter] = useState('')
  const [severityFilter, setSeverityFilter] = useState('')
  const [apiToken, setApiToken] = useState('')
  const [apiBusy, setApiBusy] = useState(false)
  const [apiErr, setApiErr] = useState<string | null>(null)

  const parsed = useMemo(() => parseSecurityDoc(p.securityJsonText), [p.securityJsonText])

  const filtered = useMemo(() => {
    if (!parsed.ok) {
      return []
    }
    const t = tacticFilter.trim().toLowerCase()
    const sev = severityFilter.trim().toLowerCase()
    return parsed.doc.spec.results.filter((r) => {
      if (t && r.tactic.toLowerCase() !== t) {
        return false
      }
      if (sev && r.severity.toLowerCase() !== sev) {
        return false
      }
      return true
    })
  }, [parsed, tacticFilter, severityFilter])

  const reformatJson = useCallback(() => {
    try {
      const o: unknown = JSON.parse(p.securityJsonText)
      p.onSecurityJsonTextChange(JSON.stringify(o, null, 2))
    } catch {
      window.alert('Invalid JSON — fix syntax before reformatting.')
    }
  }, [p])

  const onFile = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      const f = e.target.files?.[0]
      if (!f) {
        return
      }
      const r = new FileReader()
      r.onload = () => {
        if (typeof r.result === 'string') {
          p.onSecurityJsonTextChange(r.result)
        }
      }
      r.readAsText(f)
      e.target.value = ''
    },
    [p],
  )

  const runLocalScan = useCallback(async () => {
    setApiErr(null)
    setApiBusy(true)
    try {
      const doc = await fetchLocalSecurityScan(apiToken, { mode: 'local', profile: 'web-ui' })
      p.onSecurityJsonTextChange(JSON.stringify(doc, null, 2))
    } catch (e: unknown) {
      setApiErr(e instanceof Error ? e.message : String(e))
    } finally {
      setApiBusy(false)
    }
  }, [apiToken, p])

  return (
    <div className="flex h-full min-h-0 flex-col gap-4 overflow-auto p-4 md:p-6">
      <div>
        <h2 className="text-lg font-semibold text-gray-100">Security posture</h2>
        <p className="mt-1 max-w-3xl text-sm text-gray-400">
          Paste <span className="font-mono text-gray-300">omnigraph/security/v1</span> JSON from{' '}
          <code className="rounded bg-gray-900 px-1 py-0.5 font-mono text-xs">omnigraph security scan</code>, or call the
          serve API when <span className="font-mono">--enable-security-scan</span> is on. Authorized use only.
        </p>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <label className="cursor-pointer rounded-lg border border-gray-700 bg-gray-900 px-3 py-1.5 text-xs font-medium text-gray-200 hover:bg-gray-800">
          Load JSON file
          <input type="file" accept="application/json,.json" className="hidden" onChange={onFile} />
        </label>
        <button
          type="button"
          className="rounded-lg border border-gray-700 bg-gray-900 px-3 py-1.5 text-xs font-medium text-gray-200 hover:bg-gray-800"
          onClick={() => reformatJson()}
        >
          Reformat JSON
        </button>
      </div>

      <div className="rounded-xl border border-gray-800 bg-gray-900/40 p-4">
        <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">Optional: local scan via serve</p>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-end">
          <label className="flex min-w-0 flex-1 flex-col gap-1 text-xs text-gray-400">
            Bearer token
            <input
              type="password"
              autoComplete="off"
              value={apiToken}
              onChange={(e) => setApiToken(e.target.value)}
              placeholder="Same as --auth-token / OMNIGRAPH_SERVE_TOKEN"
              className="rounded-lg border border-gray-800 bg-gray-950 px-3 py-2 font-mono text-sm text-gray-100 placeholder:text-gray-600"
            />
          </label>
          <button
            type="button"
            disabled={apiBusy}
            onClick={() => void runLocalScan()}
            className="shrink-0 rounded-lg border border-blue-600/50 bg-blue-600/20 px-4 py-2 text-sm font-medium text-blue-200 hover:bg-blue-600/30 disabled:opacity-50"
          >
            {apiBusy ? 'Scanning…' : 'POST /api/v1/security/scan'}
          </button>
        </div>
        {apiErr ? <p className="mt-2 text-sm text-rose-400">{apiErr}</p> : null}
      </div>

      <div className="grid min-h-0 flex-1 grid-cols-1 gap-4 lg:grid-cols-2">
        <div className="flex min-h-[200px] flex-col gap-2">
          <span className="text-xs font-semibold uppercase tracking-wide text-gray-500">Raw document</span>
          <textarea
            value={p.securityJsonText}
            onChange={(e) => p.onSecurityJsonTextChange(e.target.value)}
            spellCheck={false}
            className="min-h-[280px] flex-1 resize-y rounded-xl border border-gray-800 bg-gray-950/80 p-3 font-mono text-xs leading-relaxed text-gray-200 focus:border-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500/30"
          />
        </div>
        <div className="flex min-h-0 flex-col gap-3 overflow-auto">
          {!parsed.ok ? (
            <p className="text-sm text-rose-400">{parsed.err}</p>
          ) : (
            <>
              <div className="rounded-xl border border-gray-800 bg-gray-900/50 p-4 text-sm text-gray-300">
                <p>
                  <span className="text-gray-500">Target:</span> {parsed.doc.metadata.target ?? '—'}
                </p>
                <p>
                  <span className="text-gray-500">Transport:</span> {parsed.doc.metadata.transport ?? '—'}
                </p>
                <p>
                  <span className="text-gray-500">Generated:</span> {parsed.doc.metadata.generatedAt}
                </p>
                <div className="mt-3 grid grid-cols-2 gap-2 text-xs sm:grid-cols-5">
                  {(
                    [
                      ['Run', parsed.doc.spec.summary.modulesRun],
                      ['Vuln', parsed.doc.spec.summary.vulnerable],
                      ['OK', parsed.doc.spec.summary.notVulnerable],
                      ['Err', parsed.doc.spec.summary.errors],
                      ['N/A', parsed.doc.spec.summary.notApplicable],
                    ] as const
                  ).map(([k, v]) => (
                    <div key={k} className="rounded-lg bg-gray-950/80 px-2 py-2 text-center">
                      <div className="text-gray-500">{k}</div>
                      <div className="text-lg font-semibold text-gray-100">{v}</div>
                    </div>
                  ))}
                </div>
              </div>
              <div className="flex flex-wrap gap-2">
                <input
                  value={tacticFilter}
                  onChange={(e) => setTacticFilter(e.target.value)}
                  placeholder="Filter tactic (e.g. defense_evasion)"
                  className="min-w-[12rem] flex-1 rounded-lg border border-gray-800 bg-gray-950 px-3 py-1.5 font-mono text-xs text-gray-200"
                />
                <select
                  value={severityFilter}
                  onChange={(e) => setSeverityFilter(e.target.value)}
                  className="rounded-lg border border-gray-800 bg-gray-950 px-3 py-1.5 text-xs text-gray-200"
                >
                  <option value="">All severities</option>
                  <option value="critical">critical</option>
                  <option value="high">high</option>
                  <option value="medium">medium</option>
                  <option value="low">low</option>
                  <option value="info">info</option>
                </select>
              </div>
              <ul className="flex flex-col gap-2">
                {filtered.map((r) => (
                  <li key={r.moduleId} className="rounded-xl border border-gray-800 bg-gray-900/40 p-3">
                    <div className="flex flex-wrap items-baseline gap-2">
                      <span className="font-mono text-xs text-blue-300">{r.techniqueId}</span>
                      <span className={`text-xs font-semibold uppercase ${severityClass(r.severity)}`}>{r.severity}</span>
                      <span className="text-xs text-gray-500">{r.status}</span>
                    </div>
                    <p className="mt-1 text-sm text-gray-200">{r.techniqueName}</p>
                    <p className="text-xs text-gray-400">{r.summary}</p>
                    {r.remediation ? (
                      <pre className="mt-2 max-h-40 overflow-auto rounded-lg border border-gray-800 bg-gray-950 p-2 font-mono text-xs text-gray-300">
                        {r.remediation}
                      </pre>
                    ) : null}
                  </li>
                ))}
              </ul>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
