/**
 * Normalizes inventory-like data from multiple backends (OpenTofu state, plan JSON, Ansible INI)
 * so the UI can browse one merged view. Logic mirrors internal/state and internal/plan (aws_instance + outputs).
 */

export type InventorySourceKind = 'terraform-state' | 'plan-json' | 'ansible-ini'

export type InventoryRow = {
  /** Unique row id for React keys */
  id: string
  /** Host / resource label shown to operators */
  name: string
  /** Target address when known (ansible_host) */
  ansibleHost: string
  source: InventorySourceKind
  /** Ansible group when parsed from INI */
  group?: string
  /** Extra vars from INI lines */
  vars?: Record<string, string>
  /** Repo-relative path when row came from a repository scan */
  originPath?: string
}

function stringifyValue(v: unknown): string | null {
  if (v === null || v === undefined) {
    return null
  }
  if (typeof v === 'string') {
    return v
  }
  if (typeof v === 'number' && Number.isFinite(v)) {
    return Number.isInteger(v) ? String(v) : String(v)
  }
  if (typeof v === 'boolean') {
    return String(v)
  }
  return null
}

type TfResource = {
  address?: string
  mode?: string
  type?: string
  name?: string
  values?: Record<string, unknown>
}

type TfRootModule = { resources?: TfResource[] }

type TfValues = {
  outputs?: Record<string, { value?: unknown }>
  root_module?: TfRootModule
}

function extractFromValues(values: TfValues | undefined, source: InventorySourceKind): InventoryRow[] {
  const rows: InventoryRow[] = []
  if (!values) {
    return rows
  }
  if (values.outputs) {
    for (const [name, ov] of Object.entries(values.outputs)) {
      const s = stringifyValue(ov?.value)
      if (s && s !== '' && s !== 'null') {
        const key = `output.${name}`
        rows.push({
          id: `${source}:${key}`,
          name: key,
          ansibleHost: s,
          source,
        })
      }
    }
  }
  const resources = values.root_module?.resources
  if (!resources) {
    return rows
  }
  for (const res of resources) {
    if (res.mode !== 'managed' || res.type !== 'aws_instance') {
      continue
    }
    const host = res.address ?? res.name ?? 'instance'
    let ip: string | null = null
    if (res.values) {
      const pub = stringifyValue(res.values.public_ip)
      if (pub && pub !== '' && pub !== 'null') {
        ip = pub
      } else {
        const priv = stringifyValue(res.values.private_ip)
        if (priv && priv !== '' && priv !== 'null') {
          ip = priv
        }
      }
    }
    if (ip) {
      rows.push({
        id: `${source}:${host}`,
        name: host,
        ansibleHost: ip,
        source,
      })
    }
  }
  return rows
}

export function extractHostsFromTfStateJson(text: string): { rows: InventoryRow[]; error?: string } {
  const t = text.trim()
  if (!t) {
    return { rows: [] }
  }
  try {
    const parsed: unknown = JSON.parse(t)
    if (typeof parsed !== 'object' || parsed === null) {
      return { rows: [], error: 'State JSON must be an object' }
    }
    const values = (parsed as { values?: TfValues }).values
    return { rows: extractFromValues(values, 'terraform-state') }
  } catch (e) {
    const m = e instanceof Error ? e.message : String(e)
    return { rows: [], error: m }
  }
}

/** `terraform show -json` plan shape (planned_values). */
export function extractHostsFromPlanJson(text: string): { rows: InventoryRow[]; error?: string } {
  const t = text.trim()
  if (!t) {
    return { rows: [] }
  }
  try {
    const parsed: unknown = JSON.parse(t)
    if (typeof parsed !== 'object' || parsed === null) {
      return { rows: [], error: 'Plan JSON must be an object' }
    }
    const planned = (parsed as { planned_values?: TfValues }).planned_values
    return { rows: extractFromValues(planned, 'plan-json') }
  } catch (e) {
    const m = e instanceof Error ? e.message : String(e)
    return { rows: [], error: m }
  }
}

/**
 * Best-effort Ansible INI parse: [groups], host lines, ansible_host= / key=value pairs.
 */
export function parseAnsibleIni(text: string): { rows: InventoryRow[]; error?: string } {
  const rows: InventoryRow[] = []
  let group = 'ungrouped'
  const lines = text.split(/\r?\n/)
  let lineNo = 0
  for (const raw of lines) {
    lineNo++
    const line = raw.trim()
    if (!line || line.startsWith('#') || line.startsWith(';')) {
      continue
    }
    const section = /^\[([^\]]+)\]\s*$/.exec(line)
    if (section) {
      group = section[1].trim()
      continue
    }
    const parts = line.split(/\s+/)
    const hostToken = parts[0]
    if (!hostToken || hostToken.includes('=')) {
      continue
    }
    const vars: Record<string, string> = {}
    for (let i = 1; i < parts.length; i++) {
      const p = parts[i]
      const eq = p.indexOf('=')
      if (eq > 0) {
        const k = p.slice(0, eq)
        const v = p.slice(eq + 1)
        vars[k] = v
      }
    }
    const ansibleHost = vars.ansible_host ?? vars.ip ?? ''
    const id = `ansible-ini:${group}:${hostToken}:${lineNo}`
    rows.push({
      id,
      name: hostToken,
      ansibleHost,
      source: 'ansible-ini',
      group,
      vars: Object.keys(vars).length ? vars : undefined,
    })
  }
  return { rows }
}

export function mergeInventoryRows(
  stateRows: InventoryRow[],
  planRows: InventoryRow[],
  iniRows: InventoryRow[],
): InventoryRow[] {
  return [...stateRows, ...planRows, ...iniRows]
}

/** Same shape as internal/inventory BuildINI for omnigraph group. */
export function buildOmnigraphIni(rows: InventoryRow[]): string {
  const withAddr = rows.filter((r) => r.ansibleHost.trim() !== '')
  if (withAddr.length === 0) {
    return '[omnigraph]\n'
  }
  const keys = [...new Set(withAddr.map((r) => sanitizeName(r.name)))].sort()
  const byName = new Map<string, string>()
  for (const r of withAddr) {
    const k = sanitizeName(r.name)
    if (!byName.has(k)) {
      byName.set(k, r.ansibleHost)
    }
  }
  let b = '[omnigraph]\n'
  for (const k of keys) {
    const host = byName.get(k)
    if (host) {
      b += `${k} ansible_host=${host}\n`
    }
  }
  return b
}

function sanitizeName(name: string): string {
  const s = name
    .split('')
    .map((c) => {
      if (/[a-zA-Z0-9_-]/.test(c)) {
        return c
      }
      return '_'
    })
    .join('')
  return s || 'host'
}

export function sourceLabel(s: InventorySourceKind): string {
  switch (s) {
    case 'terraform-state':
      return 'Terraform/OpenTofu state'
    case 'plan-json':
      return 'Plan JSON'
    case 'ansible-ini':
      return 'Ansible INI'
    default:
      return s
  }
}
