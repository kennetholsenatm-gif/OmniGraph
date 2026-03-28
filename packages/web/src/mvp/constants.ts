import sampleGraph from '../graph/sampleGraph.json'
import sampleSecurity from './sample.security.json'

export const defaultSchema = `apiVersion: omnigraph/v1alpha1
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

export const defaultGraphJson = JSON.stringify(sampleGraph, null, 2)

export const defaultPostureSecurityJson = JSON.stringify(sampleSecurity, null, 2)

export const defaultHcl = `resource "null_resource" "example" {
  triggers = {
    always = timestamp()
  }
}
`

export type MvpTab = 'visualizer' | 'schema' | 'ide' | 'inventory' | 'pipeline' | 'posture'

/** Header and breadcrumbs: human-facing name for persisted tab id. */
export function mvpTabDisplayName(tab: MvpTab): string {
  switch (tab) {
    case 'visualizer':
      return 'Topology'
    case 'schema':
      return 'Schema Contract'
    case 'ide':
      return 'Web IDE'
    case 'inventory':
      return 'Inventory'
    case 'pipeline':
      return 'GitOps Pipeline'
    case 'posture':
      return 'Posture'
    default:
      return tab
  }
}
