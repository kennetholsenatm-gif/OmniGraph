import { type ReactNode } from 'react'

import type { TriageRollup } from './nodeTriageTypes'
import { useOptionalTriageSession } from './TriageSessionContext'

function rollupBadge(rollup: TriageRollup): { text: string; className: string; pulse?: boolean } {
  switch (rollup) {
    case 'failing':
      return { text: 'FAIL', className: 'bg-rose-600 text-white', pulse: true }
    case 'drift':
      return { text: 'DRIFT', className: 'bg-amber-600 text-amber-950', pulse: true }
    case 'policy':
      return { text: 'SEC', className: 'bg-rose-500/90 text-white' }
    case 'degraded':
      return { text: 'DEG', className: 'bg-amber-500/80 text-amber-950' }
    default:
      return { text: '', className: '' }
  }
}

function rollupGlow(rollup: TriageRollup): string {
  switch (rollup) {
    case 'failing':
      return 'shadow-[0_0_22px_rgba(244,63,94,0.45)] ring-2 ring-rose-500/80'
    case 'drift':
      return 'shadow-[0_0_20px_rgba(251,191,36,0.4)] ring-2 ring-amber-400/70'
    case 'policy':
      return 'shadow-[0_0_18px_rgba(244,63,94,0.35)] ring-2 ring-rose-400/60'
    case 'degraded':
      return 'ring-2 ring-amber-500/50'
    default:
      return ''
  }
}

/**
 * TopologyNode: wraps each React Flow node body with triage glow, badges, and drift silhouette.
 * Safe outside `TriageSessionProvider` (no-op).
 */
export function TopologyNodeFrame({ nodeId, children }: { nodeId: string; children: ReactNode }) {
  const triageCtx = useOptionalTriageSession()
  const triage = triageCtx?.triageModeEnabled ? triageCtx.triageByNodeId[nodeId] : undefined
  const rollup = triage?.rollup ?? 'healthy'
  const glow = rollup !== 'healthy' ? rollupGlow(rollup) : ''
  const badge = rollup !== 'healthy' ? rollupBadge(rollup) : null
  const drift = triage?.drift

  return (
    <div className="relative">
      {drift?.active ? (
        <div
          aria-hidden
          className="pointer-events-none absolute -inset-x-1 -inset-y-1 -z-10 rounded-lg border-2 border-dashed border-amber-500/40 bg-amber-400/5 opacity-80"
          style={{ transform: 'translate(-8px, -6px)' }}
        >
          <div className="px-3 py-1 font-mono text-[9px] uppercase tracking-wider text-amber-200/60">
            Intended{drift.intendedLabel ? ` · ${drift.intendedLabel}` : ''}
            {drift.intendedSubtitle ? ` · ${drift.intendedSubtitle}` : ''}
          </div>
        </div>
      ) : null}
      <div className={`relative rounded-lg transition-shadow duration-300 ${glow}`}>
        {badge?.text ? (
          <span
            className={`absolute -right-1 -top-2 z-10 rounded px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wide ${badge.className} ${badge.pulse ? 'animate-pulse' : ''}`}
          >
            {badge.text}
          </span>
        ) : null}
        {children}
      </div>
    </div>
  )
}
