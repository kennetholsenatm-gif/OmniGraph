/**
 * Portable workspace description for GitOps: commit `omnigraph.workspace.json` at your repo root
 * and point teammates or CI at the same layout.
 */

export type OmnigraphWorkspaceManifestV1 = {
  $schema?: string
  version: 1
  /** Local clone root (the folder that contains your tofu/ansible checkouts). */
  gitRepositoryRoot: string
  openTofu: {
    /** Absolute path or path relative to gitRepositoryRoot */
    path: string
  }
  ansible?: {
    path: string
    playbook: string
  }
  schema?: {
    /** Path relative to OpenTofu root unless absolute */
    path: string
  }
}

export function joinDisplayPath(base: string, segment: string): string {
  const b = base.trim().replace(/[/\\]+$/, '')
  const s = segment.trim().replace(/^[/\\]+/, '')
  if (!b) {
    return s
  }
  if (!s) {
    return b
  }
  const sep = b.includes('\\') ? '\\' : '/'
  return `${b}${sep}${s.replace(/\//g, sep)}`
}

/** If `absolute` is under `repoRoot` (case-insensitive), return a relative path with `/` separators. */
export function tryRelativeToRepo(absolute: string, repoRoot: string): string | null {
  const abs = absolute.trim().replace(/\\/g, '/')
  const repo = repoRoot.trim().replace(/\\/g, '/').replace(/\/+$/, '')
  if (!abs || !repo) {
    return null
  }
  const absL = abs.toLowerCase()
  const repoL = repo.toLowerCase()
  if (absL === repoL) {
    return ''
  }
  const prefix = `${repoL}/`
  if (!absL.startsWith(prefix)) {
    return null
  }
  return abs.slice(repo.length).replace(/^\/+/, '')
}

export function buildWorkspaceManifest(args: {
  gitRepositoryRoot: string
  pipelineWorkdir: string
  pipelineAnsibleRoot: string
  pipelinePlaybookRel: string
  schemaCliPath: string
}): OmnigraphWorkspaceManifestV1 {
  const root = args.gitRepositoryRoot.trim()
  const wd = args.pipelineWorkdir.trim()
  const ar = args.pipelineAnsibleRoot.trim()

  const tofuPath = wd ? (root ? tryRelativeToRepo(wd, root) ?? wd : wd) : ''
  const ansiblePath = ar ? (root ? tryRelativeToRepo(ar, root) ?? ar : ar) : ''

  return {
    version: 1,
    gitRepositoryRoot: root || '',
    openTofu: { path: tofuPath || '.' },
    ...(ansiblePath
      ? {
          ansible: {
            path: ansiblePath,
            playbook: (args.pipelinePlaybookRel.trim() || 'site.yml').replace(/\\/g, '/'),
          },
        }
      : {}),
    schema: {
      path: args.schemaCliPath.trim() || '.omnigraph.schema',
    },
  }
}

export function manifestToJson(m: OmnigraphWorkspaceManifestV1): string {
  return `${JSON.stringify(m, null, 2)}\n`
}
