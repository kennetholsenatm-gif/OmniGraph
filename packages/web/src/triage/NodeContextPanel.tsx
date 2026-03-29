import { ExternalLink, ShieldAlert, Terminal, X } from 'lucide-react'

import type { GraphNodeSelection } from '../graph/GraphCanvas'
import type { NodeTriageState } from './nodeTriageTypes'

export type NodeContextPanelProps = {
  open: boolean
  selectedNode: GraphNodeSelection | null
  triage: NodeTriageState | undefined
  streamStatus: string
  onClose: () => void
  onDetach: () => void
}

function severityDot(sev: string): string {
  switch (sev) {
    case 'critical':
    case 'error':
      return 'bg-rose-500'
    case 'warning':
      return 'bg-amber-500'
    default:
      return 'bg-slate-500'
  }
}

/**
 * Unified triage side context: reconciliation hand-off + posture + drift for one node id.
 * Opens when Topology triage mode is on and a node is selected.
 */
export function NodeContextPanel({
  open,
  selectedNode,
  triage,
  streamStatus,
  onClose,
  onDetach,
}: NodeContextPanelProps) {
  if (!open) {
    return null
  }

  return (
    <div
      className="flex h-full min-h-0 w-full flex-col border-gray-800 bg-gray-900/85 backdrop-blur-md lg:border-l"
      role="complementary"
      aria-label="Triage context panel"
    >
      <div className="flex items-center justify-between gap-2 border-b border-gray-800 px-4 py-3">
        <div className="min-w-0">
          <h2 className="truncate text-sm font-bold uppercase tracking-wide text-blue-400">Triage</h2>
          <p className="truncate font-mono text-[11px] text-gray-500">
            {selectedNode ? selectedNode.id : 'Select a node'}
          </p>
        </div>
        <div className="flex shrink-0 gap-1">
          <button
            type="button"
            className="rounded border border-gray-700 bg-gray-900 px-2 py-1 text-[10px] text-gray-300 hover:bg-gray-800"
            title="Open panel in a new window (same sync channel)"
            onClick={onDetach}
          >
            <ExternalLink size={14} className="inline" aria-hidden /> Detach
          </button>
          <button
            type="button"
            className="rounded border border-gray-700 p-1.5 text-gray-400 hover:bg-gray-800 lg:hidden"
            aria-label="Close triage panel"
            onClick={onClose}
          >
            <X size={16} />
          </button>
        </div>
      </div>

      <div className="flex items-center gap-2 border-b border-gray-800/80 px-4 py-2 text-[10px] text-gray-500">
        <span className="rounded bg-gray-900 px-1.5 py-0.5 font-mono text-gray-400">stream:{streamStatus}</span>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto px-4 py-4">
        {!selectedNode ? (
          <p className="text-sm text-gray-500">Click a graph node to aggregate Ansible hand-off, drift, and posture.</p>
        ) : (
          <div className="space-y-6">
            <section>
              <h3 className="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-gray-400">
                <Terminal size={14} aria-hidden /> Reconciliation / hand-off
              </h3>
              {triage && triage.reconciliationLogs.length > 0 ? (
                <ul className="space-y-2 border border-gray-800/80 bg-gray-900/50 p-2 font-mono text-[11px] text-gray-200">
                  {triage.reconciliationLogs.map((l) => (
                    <li key={l.id} className="border-b border-gray-800/60 pb-2 last:border-0">
                      <span className="text-gray-500">{l.ts}</span>{' '}
                      <span className="text-blue-400">[{l.phase}]</span>{' '}
                      <span className={l.level === 'error' ? 'text-rose-300' : l.level === 'warn' ? 'text-amber-200' : ''}>
                        {l.message}
                      </span>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="text-[11px] text-gray-600">No streamed lines for this node yet (graph `debugLog` below).</p>
              )}
              {selectedNode.debugLog.length > 0 ? (
                <pre className="mt-2 max-h-40 overflow-auto rounded border border-amber-900/40 bg-amber-950/20 p-2 text-[10px] text-amber-100/90">
                  {selectedNode.debugLog.join('\n')}
                </pre>
              ) : null}
            </section>

            <section>
              <h3 className="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-gray-400">
                <ShieldAlert size={14} aria-hidden /> Posture
              </h3>
              {triage && triage.postureFindings.length > 0 ? (
                <ul className="space-y-2">
                  {triage.postureFindings.map((f) => (
                    <li
                      key={f.findingId}
                      className="rounded border border-gray-800 bg-gray-900/60 p-2 text-[11px] text-gray-200"
                    >
                      <div className="flex items-center gap-2">
                        <span className={`h-2 w-2 rounded-full ${severityDot(f.severity)}`} aria-hidden />
                        <span className="font-medium text-gray-100">{f.title}</span>
                        <span className="text-gray-500">({f.severity})</span>
                      </div>
                      <p className="mt-1 text-gray-400">{f.summary}</p>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="text-[11px] text-gray-600">No correlated posture rows for this node.</p>
              )}
            </section>

            <section>
              <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-400">Drift</h3>
              {triage?.drift?.active && triage.drift.fieldDeltas.length > 0 ? (
                <table className="w-full border-collapse text-left text-[11px] text-gray-200">
                  <thead>
                    <tr className="border-b border-gray-800 text-gray-500">
                      <th className="py-1 pr-2">Field</th>
                      <th className="py-1 pr-2">Expected</th>
                      <th className="py-1">Actual</th>
                    </tr>
                  </thead>
                  <tbody>
                    {triage.drift.fieldDeltas.map((d) => (
                      <tr key={d.path} className="border-b border-gray-800/80">
                        <td className="py-1.5 pr-2 font-mono text-amber-200/90">{d.path}</td>
                        <td className="py-1.5 pr-2 font-mono text-emerald-200/70">{d.expected}</td>
                        <td className="py-1.5 font-mono text-rose-200/90">{d.actual}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <p className="text-[11px] text-gray-600">No active drift payload for this node.</p>
              )}
            </section>
          </div>
        )}
      </div>
    </div>
  )
}
