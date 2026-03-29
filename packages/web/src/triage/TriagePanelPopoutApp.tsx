import { useEffect, useState } from 'react'

import type { GraphNodeSelection } from '../graph/GraphCanvas'
import { NodeContextPanel } from './NodeContextPanel'
import type { NodeTriageState } from './nodeTriageTypes'
import { createTriagePanelChannel, isTriagePanelMessage } from './triagePanelPopoutChannel'

function stubSelection(nodeId: string): GraphNodeSelection {
  return {
    id: nodeId,
    label: nodeId,
    kind: '',
    state: '',
    subtitle: '',
    debugLog: [],
    enclave: '',
    attributes: {},
  }
}

/** Detached triage panel window: stays in sync via BroadcastChannel with the main workspace. */
export default function TriagePanelPopoutApp() {
  const [triageByNodeId, setTriageByNodeId] = useState<Record<string, NodeTriageState>>({})
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [streamStatus, setStreamStatus] = useState('detached')

  useEffect(() => {
    const ch = createTriagePanelChannel()
    if (!ch) {
      return
    }
    const onMsg = (ev: MessageEvent) => {
      if (!isTriagePanelMessage(ev.data)) {
        return
      }
      if (ev.data.type === 'full_sync') {
        setTriageByNodeId(ev.data.triageByNodeId)
        setSelectedNodeId(ev.data.selectedNodeId)
        setStreamStatus(ev.data.triageModeEnabled ? 'synced' : 'idle')
      } else if (ev.data.type === 'selection') {
        setSelectedNodeId(ev.data.selectedNodeId)
      }
    }
    ch.addEventListener('message', onMsg)
    return () => {
      ch.removeEventListener('message', onMsg)
      ch.close()
    }
  }, [])

  const selectedNode = selectedNodeId ? stubSelection(selectedNodeId) : null
  const triage = selectedNodeId ? triageByNodeId[selectedNodeId] : undefined

  return (
    <div className="flex h-dvh min-h-dvh flex-col bg-gray-950 text-gray-100">
      <NodeContextPanel
        open
        selectedNode={selectedNode}
        triage={triage}
        streamStatus={streamStatus}
        onClose={() => window.close()}
        onDetach={() => {}}
      />
    </div>
  )
}
