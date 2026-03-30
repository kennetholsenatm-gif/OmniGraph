import OmniGraphMVP from './mvp/OmniGraphMVP'
import GraphPopoutApp from './mvp/GraphPopoutApp'
import TriagePanelPopoutApp from './triage/TriagePanelPopoutApp'

function getPopoutMode(): 'graph' | 'triage-panel' | null {
  if (typeof window === 'undefined') {
    return null
  }
  const p = new URLSearchParams(window.location.search).get('popout')
  if (p === 'graph') {
    return 'graph'
  }
  if (p === 'triage-panel') {
    return 'triage-panel'
  }
  return null
}

export default function App() {
  const mode = getPopoutMode()
  if (mode === 'graph') {
    return <GraphPopoutApp />
  }
  if (mode === 'triage-panel') {
    return <TriagePanelPopoutApp />
  }
  return <OmniGraphMVP />
}