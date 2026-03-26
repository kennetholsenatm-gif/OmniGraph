/**
 * Browser-side repository scan via File System Access API (Chromium / Edge).
 * Mirrors classification in internal/repo/discover.go — keep rules aligned when changing either.
 */

export type RepoScanKind =
  | 'terraform-state'
  | 'terraform-hcl'
  | 'omnigraph-schema'
  | 'ansible-cfg'
  | 'ansible-playbook'
  | 'ansible-inventory'
  | 'terraform-plan-binary'

export type RepoDiscovered = { path: string; kind: RepoScanKind }

export type RepoScanSession = {
  rootName: string
  discovered: RepoDiscovered[]
  stateFiles: { path: string; text: string }[]
  planFiles: { path: string; text: string }[]
  iniFiles: { path: string; text: string }[]
  schemas: { path: string; text: string }[]
}

const SKIP_DIRS = new Set(
  [
    '.git',
    'node_modules',
    '.terraform',
    '.venv',
    'venv',
    '__pycache__',
    'vendor',
    'dist',
    'build',
    'target',
    '.idea',
    '.vscode',
  ].map((s) => s.toLowerCase()),
)

const MAX_DEPTH = 14
const MAX_DISCOVERED = 500
const MAX_READ_BYTES = 6 * 1024 * 1024
const MAX_STATE_READ = 12
const MAX_INI_READ = 20
const MAX_SCHEMA_READ = 5

function classifyFile(relPath: string, nameLower: string): RepoScanKind | null {
  const relLower = relPath.toLowerCase()
  if (nameLower.endsWith('.tfstate')) {
    return 'terraform-state'
  }
  if (nameLower === '.omnigraph.schema' || nameLower.endsWith('.omnigraph.schema')) {
    return 'omnigraph-schema'
  }
  if (nameLower.endsWith('.tf') || nameLower.endsWith('.tofu')) {
    return 'terraform-hcl'
  }
  if (nameLower === 'ansible.cfg') {
    return 'ansible-cfg'
  }
  if (
    nameLower === 'site.yml' ||
    nameLower === 'site.yaml' ||
    nameLower === 'playbook.yml' ||
    nameLower === 'playbook.yaml'
  ) {
    return 'ansible-playbook'
  }
  if (nameLower === 'hosts' || nameLower.endsWith('.ini')) {
    if (relLower.includes('inventory') || nameLower === 'hosts') {
      return 'ansible-inventory'
    }
  }
  if (nameLower === 'tfplan' || nameLower.endsWith('.tfplan')) {
    return 'terraform-plan-binary'
  }
  return null
}

async function walkDiscover(
  dir: FileSystemDirectoryHandle,
  rel: string,
  depth: number,
  acc: RepoDiscovered[],
): Promise<void> {
  if (depth > MAX_DEPTH || acc.length >= MAX_DISCOVERED) {
    return
  }
  for await (const ent of dir.values()) {
    if (acc.length >= MAX_DISCOVERED) {
      break
    }
    const name = ent.name
    const pathSeg = rel ? `${rel}/${name}` : name
    if (ent.kind === 'directory') {
      if (SKIP_DIRS.has(name.toLowerCase())) {
        continue
      }
      const sub = await dir.getDirectoryHandle(name)
      await walkDiscover(sub, pathSeg, depth + 1, acc)
      continue
    }
    if (ent.kind !== 'file') {
      continue
    }
    const kind = classifyFile(pathSeg, name.toLowerCase())
    if (kind) {
      acc.push({ path: pathSeg.replace(/\\/g, '/'), kind })
    }
  }
}

async function readFileFromRoot(
  root: FileSystemDirectoryHandle,
  rel: string,
): Promise<{ text: string; truncated: boolean } | null> {
  const parts = rel.split('/').filter(Boolean)
  if (parts.length === 0) {
    return null
  }
  let dir = root
  for (let i = 0; i < parts.length - 1; i++) {
    try {
      dir = await dir.getDirectoryHandle(parts[i]!)
    } catch {
      return null
    }
  }
  let fh: FileSystemFileHandle
  try {
    fh = await dir.getFileHandle(parts[parts.length - 1]!)
  } catch {
    return null
  }
  const file = await fh.getFile()
  if (file.size > MAX_READ_BYTES) {
    const slice = file.slice(0, MAX_READ_BYTES)
    const text = await slice.text()
    return { text, truncated: true }
  }
  return { text: await file.text(), truncated: false }
}

export function isRepoFolderPickerSupported(): boolean {
  return typeof window !== 'undefined' && typeof window.showDirectoryPicker === 'function'
}

export async function scanRepositoryFolder(): Promise<RepoScanSession | null> {
  const picker = window.showDirectoryPicker
  if (!picker) {
    return null
  }
  let root: FileSystemDirectoryHandle
  try {
    root = await picker.call(window, { mode: 'read' })
  } catch {
    return null
  }

  const discovered: RepoDiscovered[] = []
  await walkDiscover(root, '', 0, discovered)

  const statePaths = discovered.filter((d) => d.kind === 'terraform-state').map((d) => d.path)
  const iniPaths = discovered.filter((d) => d.kind === 'ansible-inventory').map((d) => d.path)
  const schemaPaths = discovered.filter((d) => d.kind === 'omnigraph-schema').map((d) => d.path)

  const stateFiles: { path: string; text: string }[] = []
  for (const p of statePaths.slice(0, MAX_STATE_READ)) {
    const r = await readFileFromRoot(root, p)
    if (r) {
      stateFiles.push({ path: p, text: r.text })
    }
  }

  const iniFiles: { path: string; text: string }[] = []
  for (const p of iniPaths.slice(0, MAX_INI_READ)) {
    const r = await readFileFromRoot(root, p)
    if (r) {
      iniFiles.push({ path: p, text: r.text })
    }
  }

  const schemas: { path: string; text: string }[] = []
  for (const p of schemaPaths.slice(0, MAX_SCHEMA_READ)) {
    const r = await readFileFromRoot(root, p)
    if (r) {
      schemas.push({ path: p, text: r.text })
    }
  }

  return {
    rootName: root.name,
    discovered,
    stateFiles,
    planFiles: [],
    iniFiles,
    schemas,
  }
}
