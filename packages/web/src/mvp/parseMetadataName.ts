import YAML from 'yaml'

/** Best-effort parse of `metadata.name` from a project YAML document. */
export function tryParseMetadataName(src: string): string | null {
  try {
    const doc = YAML.parse(src) as unknown
    if (!doc || typeof doc !== 'object' || Array.isArray(doc)) {
      return null
    }
    const meta = (doc as { metadata?: unknown }).metadata
    if (!meta || typeof meta !== 'object' || Array.isArray(meta)) {
      return null
    }
    const name = (meta as { name?: unknown }).name
    if (typeof name !== 'string') {
      return null
    }
    const t = name.trim()
    return t.length > 0 ? t : null
  } catch {
    return null
  }
}
