import { defaultGraphJson, defaultHcl, defaultPostureSecurityJson, defaultSchema } from './constants'

export const WORKSPACE_STORAGE_KEY = 'omnigraph.web.workspace.v1'

export type WorkspaceSnapshotV1 = {
  v: 1
  schemaText: string
  graphText: string
  hclText: string
  projectLabel: string
  /** Local clone root; drives portable paths and omnigraph.workspace.json export. */
  gitRepoRoot?: string
  /** Path used in generated CLI examples (on-disk path). */
  schemaCliPath: string
  schemaFileNameHint?: string
  graphFileNameHint?: string
  /** OpenTofu/Terraform root for generated orchestrate command. */
  pipelineWorkdir?: string
  /** Separate Ansible repo root; enables `--ansible-root` in the CLI when set. */
  pipelineAnsibleRoot?: string
  /** Playbook path relative to Ansible root (or to workdir if no Ansible root). */
  pipelinePlaybookRel?: string
  /** When set, used as literal `--playbook` and disables `--ansible-root` in the builder. */
  pipelinePlaybookOverride?: string
  pipelineUsePlaybookOverride?: boolean
  pipelineSchema?: string
  pipelineTfBinary?: string
  pipelinePlanFile?: string
  pipelineStateFile?: string
  pipelineRunner?: 'exec' | 'container'
  pipelineContainerRuntime?: string
  pipelineAutoApprove?: boolean
  pipelineSkipAnsible?: boolean
  pipelineGraphOut?: string
  pipelineTelemetryFile?: string
  pipelineIacEngine?: string
  pipelineTofuImage?: string
  pipelineAnsibleImage?: string
  pipelinePulumiImage?: string
  /** Pasted Terraform/OpenTofu JSON state for Inventory tab */
  inventoryTfStateText?: string
  /** Pasted plan JSON (terraform show -json) */
  inventoryPlanJsonText?: string
  /** Pasted Ansible INI */
  inventoryAnsibleIniText?: string
  /** Posture / security JSON for the Posture tab */
  postureSecurityJson?: string
}

export function defaultWorkspaceSnapshot(): WorkspaceSnapshotV1 {
  return {
    v: 1,
    schemaText: defaultSchema,
    graphText: defaultGraphJson,
    hclText: defaultHcl,
    projectLabel: 'demo',
    schemaCliPath: '.omnigraph.schema',
    postureSecurityJson: defaultPostureSecurityJson,
  }
}

function isRecord(x: unknown): x is Record<string, unknown> {
  return typeof x === 'object' && x !== null && !Array.isArray(x)
}

export function loadWorkspace(): WorkspaceSnapshotV1 | null {
  try {
    const raw = localStorage.getItem(WORKSPACE_STORAGE_KEY)
    if (!raw) {
      return null
    }
    const j: unknown = JSON.parse(raw)
    if (!isRecord(j) || j.v !== 1) {
      return null
    }
    const schemaText = typeof j.schemaText === 'string' ? j.schemaText : null
    const graphText = typeof j.graphText === 'string' ? j.graphText : null
    const hclText = typeof j.hclText === 'string' ? j.hclText : null
    const projectLabel = typeof j.projectLabel === 'string' ? j.projectLabel : null
    const schemaCliPath = typeof j.schemaCliPath === 'string' ? j.schemaCliPath : null
    if (!schemaText || !graphText || !hclText || !projectLabel || !schemaCliPath) {
      return null
    }
    const optStr = (k: string) => (typeof j[k] === 'string' ? (j[k] as string) : undefined)
    const runner =
      j.pipelineRunner === 'exec' || j.pipelineRunner === 'container' ? j.pipelineRunner : undefined
    return {
      v: 1,
      schemaText,
      graphText,
      hclText,
      projectLabel,
      gitRepoRoot: optStr('gitRepoRoot'),
      schemaCliPath,
      schemaFileNameHint: typeof j.schemaFileNameHint === 'string' ? j.schemaFileNameHint : undefined,
      graphFileNameHint: typeof j.graphFileNameHint === 'string' ? j.graphFileNameHint : undefined,
      pipelineWorkdir: optStr('pipelineWorkdir'),
      pipelineAnsibleRoot: optStr('pipelineAnsibleRoot'),
      pipelinePlaybookRel: optStr('pipelinePlaybookRel'),
      pipelinePlaybookOverride: optStr('pipelinePlaybookOverride'),
      pipelineUsePlaybookOverride: typeof j.pipelineUsePlaybookOverride === 'boolean' ? j.pipelineUsePlaybookOverride : undefined,
      pipelineSchema: optStr('pipelineSchema'),
      pipelineTfBinary: optStr('pipelineTfBinary'),
      pipelinePlanFile: optStr('pipelinePlanFile'),
      pipelineStateFile: optStr('pipelineStateFile'),
      pipelineRunner: runner,
      pipelineContainerRuntime: optStr('pipelineContainerRuntime'),
      pipelineAutoApprove: typeof j.pipelineAutoApprove === 'boolean' ? j.pipelineAutoApprove : undefined,
      pipelineSkipAnsible: typeof j.pipelineSkipAnsible === 'boolean' ? j.pipelineSkipAnsible : undefined,
      pipelineGraphOut: optStr('pipelineGraphOut'),
      pipelineTelemetryFile: optStr('pipelineTelemetryFile'),
      pipelineIacEngine: optStr('pipelineIacEngine'),
      pipelineTofuImage: optStr('pipelineTofuImage'),
      pipelineAnsibleImage: optStr('pipelineAnsibleImage'),
      pipelinePulumiImage: optStr('pipelinePulumiImage'),
      inventoryTfStateText: optStr('inventoryTfStateText'),
      inventoryPlanJsonText: optStr('inventoryPlanJsonText'),
      inventoryAnsibleIniText: optStr('inventoryAnsibleIniText'),
      postureSecurityJson: optStr('postureSecurityJson'),
    }
  } catch {
    return null
  }
}

export function persistWorkspace(s: WorkspaceSnapshotV1): void {
  try {
    localStorage.setItem(WORKSPACE_STORAGE_KEY, JSON.stringify(s))
  } catch {
    // quota or private mode
  }
}

export function clearWorkspaceStorage(): void {
  try {
    localStorage.removeItem(WORKSPACE_STORAGE_KEY)
  } catch {
    // ignore
  }
}
