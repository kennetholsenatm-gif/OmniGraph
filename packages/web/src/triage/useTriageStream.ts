/* eslint-disable react-hooks/set-state-in-effect -- WebSocket and demo timer map external events to stream status state */
import { useEffect, useState } from 'react'

import type { NodeTriagePatch, NodeTriageState, TriageWsMessage } from './nodeTriageTypes'

function parseWsPayload(raw: string): TriageWsMessage | null {
  try {
    return JSON.parse(raw) as TriageWsMessage
  } catch {
    return null
  }
}

export type TriageStreamHandlers = {
  onPatch: (patch: NodeTriagePatch) => void
  onSnapshot: (byNodeId: Record<string, NodeTriageState>) => void
}

export type TriageStreamStatus = 'idle' | 'connecting' | 'live' | 'demo' | 'error'

/**
 * WebSocket feed for triage patches; falls back to a soft demo ticker when no URL is set.
 * Set `VITE_TRIAGE_WS_URL` in env to connect to your gateway.
 */
export function useTriageStream(
  wsUrl: string | undefined,
  nodeIds: string[],
  { onPatch, onSnapshot }: TriageStreamHandlers,
  enabled: boolean,
): { status: TriageStreamStatus; lastError: string | null } {
  const [status, setStatus] = useState<TriageStreamStatus>('idle')
  const [lastError, setLastError] = useState<string | null>(null)

  useEffect(() => {
    if (!enabled || nodeIds.length === 0) {
      setStatus('idle')
      setLastError(null)
      return
    }

    if (wsUrl && wsUrl.startsWith('ws')) {
      setStatus('connecting')
      setLastError(null)
      const ws = new WebSocket(wsUrl)
      const onOpen = () => {
        setStatus('live')
        setLastError(null)
      }
      const onMsg = (ev: MessageEvent) => {
        const msg = parseWsPayload(String(ev.data))
        if (!msg) {
          return
        }
        if (msg.type === 'node_triage_patch') {
          onPatch(msg.patch)
        } else if (msg.type === 'triage_snapshot') {
          onSnapshot(msg.byNodeId)
        }
      }
      const onErr = () => {
        setStatus('error')
        setLastError('WebSocket error')
      }
      ws.addEventListener('open', onOpen)
      ws.addEventListener('message', onMsg)
      ws.addEventListener('error', onErr)
      return () => {
        ws.removeEventListener('open', onOpen)
        ws.removeEventListener('message', onMsg)
        ws.removeEventListener('error', onErr)
        ws.close()
      }
    }

    setStatus('demo')
    setLastError(null)
    let i = 0
    const id = window.setInterval(() => {
      const nid = nodeIds[i % nodeIds.length]
      i += 1
      const rollups = ['healthy', 'degraded', 'drift', 'policy', 'failing'] as const
      const r = rollups[i % rollups.length]
      const patch: NodeTriagePatch = {
        nodeId: nid,
        rollup: r,
        updatedAt: new Date().toISOString(),
        streamEpoch: i,
        reconciliationLogs: [
          {
            id: `demo-${i}`,
            ts: new Date().toISOString(),
            phase: 'ansible',
            level: r === 'failing' ? 'error' : 'info',
            message:
              r === 'drift'
                ? 'State reconciliation: observed image tag differs from desired'
                : r === 'policy'
                  ? 'Posture module flagged medium finding correlated to this host'
                  : 'Handoff log (demo stream)',
          },
        ],
        postureFindings:
          r === 'policy'
            ? [
                {
                  findingId: `demo-policy-${i}`,
                  title: 'CIS baseline deviation',
                  severity: 'warning',
                  summary: 'OpenSSH hardening recommendation not satisfied',
                  nodeId: nid,
                },
              ]
            : [],
        drift:
          r === 'drift'
            ? {
                active: true,
                intendedLabel: 'declared',
                intendedSubtitle: 'image:app:1.2.3',
                fieldDeltas: [
                  { path: 'spec.containers[0].image', expected: 'app:1.2.3', actual: 'app:1.2.0' },
                ],
              }
            : { active: false, fieldDeltas: [] },
      }
      onPatch(patch)
    }, 9000)

    return () => window.clearInterval(id)
  }, [enabled, wsUrl, nodeIds, onPatch, onSnapshot])

  return { status, lastError }
}