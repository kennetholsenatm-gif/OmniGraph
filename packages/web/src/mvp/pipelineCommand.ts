import { shellQuote } from './shellQuote'

export function looksAbsoluteHostPath(p: string): boolean {
  const s = p.trim()
  if (!s) {
    return false
  }
  if (s.startsWith('/')) {
    return true
  }
  return /^[A-Za-z]:[\\/]/.test(s)
}

/** Builds `omnigraph orchestrate` with optional `--ansible-root`. */
export function buildOrchestrateCommand(o: {
  workdir: string
  schema: string
  ansibleRoot: string
  playbookRel: string
  playbookOverride: string
  tfBinary: string
  planFile: string
  stateFile: string
  runner: string
  containerRuntime: string
  autoApprove: boolean
  skipAnsible: boolean
  graphOut: string
  telemetryFile: string
  iacEngine: string
  tofuImage: string
  ansibleImage: string
  pulumiImage: string
}): string {
  const parts = ['omnigraph', 'orchestrate']
  const wd = o.workdir.trim()
  parts.push('--workdir', wd ? shellQuote(wd) : "'<WORKDIR>'")
  if (o.schema.trim() && o.schema !== '.omnigraph.schema') {
    parts.push('--schema', shellQuote(o.schema))
  }

  const override = o.playbookOverride.trim()
  const ar = o.ansibleRoot.trim()
  const rel = (o.playbookRel.trim() || 'site.yml').replace(/\\/g, '/')

  if (!o.skipAnsible) {
    if (override) {
      parts.push('--playbook', shellQuote(override))
    } else if (ar) {
      parts.push('--ansible-root', shellQuote(ar))
      parts.push('--playbook', shellQuote(rel))
    } else if (rel) {
      parts.push('--playbook', shellQuote(rel))
    }
  }

  if (o.tfBinary.trim() && o.tfBinary !== 'tofu') {
    parts.push('--tf-binary', shellQuote(o.tfBinary))
  }
  if (o.planFile.trim() && o.planFile !== 'tfplan') {
    parts.push('--plan-file', shellQuote(o.planFile))
  }
  if (o.stateFile.trim() && o.stateFile !== 'terraform.tfstate') {
    parts.push('--state-file', shellQuote(o.stateFile))
  }
  if (o.runner === 'container') {
    parts.push('--runner', 'container')
    if (o.containerRuntime.trim()) {
      parts.push('--container-runtime', shellQuote(o.containerRuntime))
    }
    if (o.tofuImage.trim()) {
      parts.push('--tofu-image', shellQuote(o.tofuImage))
    }
    if (o.ansibleImage.trim()) {
      parts.push('--ansible-image', shellQuote(o.ansibleImage))
    }
    if (o.pulumiImage.trim()) {
      parts.push('--pulumi-image', shellQuote(o.pulumiImage))
    }
  }
  if (o.autoApprove) {
    parts.push('--auto-approve')
  }
  if (o.skipAnsible) {
    parts.push('--skip-ansible')
  }
  if (o.graphOut.trim()) {
    parts.push('--graph-out', shellQuote(o.graphOut))
  }
  if (o.telemetryFile.trim()) {
    parts.push('--telemetry-file', shellQuote(o.telemetryFile))
  }
  if (o.iacEngine.trim()) {
    parts.push('--iac-engine', shellQuote(o.iacEngine))
  }
  return parts.join(' ')
}
