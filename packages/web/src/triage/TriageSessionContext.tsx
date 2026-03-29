/* eslint-disable react-refresh/only-export-components -- context module exports hooks + imperative helpers */
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'

import type { NodeTriagePatch, NodeTriageState } from './nodeTriageTypes'
import {
  createTriagePanelChannel,
  type TriagePanelBroadcastMessage,
} from './triagePanelPopoutChannel'
import { useTriageStream } from './useTriageStream'

export type TriageSessionValue = {
  triageByNodeId: Record<string, NodeTriageState>
  triageModeEnabled: boolean
  setTriageModeEnabled: (v: boolean) => void
  panelDetached: boolean
  setPanelDetached: (v: boolean) => void
  applyPatch: (patch: NodeTriagePatch) => void
  replaceAll: (m: Record<string, NodeTriageState>) => void
  streamStatus: 'idle' | 'connecting' | 'live' | 'demo' | 'error'
  streamError: string | null
}

const Ctx = createContext<TriageSessionValue | null>(null)

function mergePatch(prev: Record<string, NodeTriageState>, patch: NodeTriagePatch): Record<string, NodeTriageState> {
  const cur = prev[patch.nodeId]
  const base: NodeTriageState =
    cur ??
    ({
      nodeId: patch.nodeId,
      rollup: 'healthy',
      reconciliationLogs: [],
      postureFindings: [],
      updatedAt: new Date().toISOString(),
    } satisfies NodeTriageState)
  const next: NodeTriageState = {
    ...base,
    ...patch,
    reconciliationLogs: patch.reconciliationLogs ?? base.reconciliationLogs,
    postureFindings: patch.postureFindings ?? base.postureFindings,
    drift: patch.drift !== undefined ? patch.drift : base.drift,
    nodeId: patch.nodeId,
    updatedAt: patch.updatedAt ?? new Date().toISOString(),
  }
  return { ...prev, [patch.nodeId]: next }
}

export function useTriageSession(): TriageSessionValue {
  const v = useContext(Ctx)
  if (!v) {
    throw new Error('useTriageSession must be used under TriageSessionProvider')
  }
  return v
}

export function useOptionalTriageSession(): TriageSessionValue | null {
  return useContext(Ctx)
}

export function TriageSessionProvider({
  children,
  graphText,
  syncSelectionId,
}: {
  children: ReactNode
  graphText: string
  syncSelectionId: string | null
}) {
  const [triageByNodeId, setTriageByNodeId] = useState<Record<string, NodeTriageState>>({})
  const [triageModeEnabled, setTriageModeEnabled] = useState(false)
  const [panelDetached, setPanelDetached] = useState(false)

  const nodeIds = useMemo(() => {
    const t = graphText.trim()
    if (!t) {
      return []
    }
    try {
      const doc = JSON.parse(t) as { spec?: { nodes?: { id: string }[] } }
      const nodes = doc.spec?.nodes
      if (!Array.isArray(nodes)) {
        return []
      }
      return nodes.map((n) => n.id).filter(Boolean)
    } catch {
      return []
    }
  }, [graphText])

  const applyPatch = useCallback((patch: NodeTriagePatch) => {
    setTriageByNodeId((p) => mergePatch(p, patch))
  }, [])

  const replaceAll = useCallback((m: Record<string, NodeTriageState>) => {
    setTriageByNodeId(m)
  }, [])

  const wsUrl = import.meta.env.VITE_TRIAGE_WS_URL as string | undefined

  const handlers = useMemo(
    () => ({
      onPatch: applyPatch,
      onSnapshot: replaceAll,
    }),
    [applyPatch, replaceAll],
  )

  const streamEnabled = triageModeEnabled && nodeIds.length > 0
  const { status: streamStatus, lastError: streamError } = useTriageStream(wsUrl, nodeIds, handlers, streamEnabled)

  const bcRef = useRef<BroadcastChannel | null>(null)
  useEffect(() => {
    bcRef.current = createTriagePanelChannel()
    return () => {
      bcRef.current?.close()
      bcRef.current = null
    }
  }, [])

  const postPanel = useCallback(
    (msg: TriagePanelBroadcastMessage) => {
      try {
        bcRef.current?.postMessage(msg)
      } catch {
        /* ignore */
      }
    },
    [],
  )

  useEffect(() => {
    if (!panelDetached) {
      return
    }
    postPanel({
      type: 'full_sync',
      triageByNodeId,
      selectedNodeId: syncSelectionId,
      triageModeEnabled,
    })
  }, [panelDetached, postPanel, triageByNodeId, triageModeEnabled, syncSelectionId])

  const value = useMemo(
    () =>
      ({
        triageByNodeId,
        triageModeEnabled,
        setTriageModeEnabled,
        panelDetached,
        setPanelDetached,
        applyPatch,
        replaceAll,
        streamStatus,
        streamError,
      }) satisfies TriageSessionValue,
    [
      triageByNodeId,
      triageModeEnabled,
      panelDetached,
      applyPatch,
      replaceAll,
      streamStatus,
      streamError,
    ],
  )

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}

export function postTriageSelectionDetached(selectedNodeId: string | null): void {
  try {
    const ch = createTriagePanelChannel()
    if (ch) {
      ch.postMessage({ type: 'selection', selectedNodeId } satisfies TriagePanelBroadcastMessage)
      ch.close()
    }
  } catch {
    /* ignore */
  }
}