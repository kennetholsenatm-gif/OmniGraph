import Ajv2020 from 'ajv/dist/2020'
import { parse as parseYaml } from 'yaml'
import schemaRaw from '../../schemas/omnigraph.schema.json?raw'

const ajv = new Ajv2020({ allErrors: true, strict: false })
const validate = ajv.compile(JSON.parse(schemaRaw) as object)

export function parseOmnigraphText(text: string): unknown {
  const t = text.trim()
  if (!t) {
    return {}
  }
  if (t.startsWith('{')) {
    return JSON.parse(t) as unknown
  }
  return parseYaml(t) as unknown
}

export type ValidateResult =
  | { ok: true }
  | { ok: false; message: string }

export function validateOmnigraphInstance(data: unknown): ValidateResult {
  const ok = validate(data) as boolean
  if (ok) {
    return { ok: true }
  }
  const errs = validate.errors ?? []
  const msg =
    errs
      .map((e) => `${e.instancePath === '' ? '/' : e.instancePath} ${e.message ?? ''}`)
      .join('\n') || 'invalid'
  return { ok: false, message: msg }
}

export function validateOmnigraphText(text: string): ValidateResult {
  try {
    const data = parseOmnigraphText(text)
    return validateOmnigraphInstance(data)
  } catch (e) {
    const m = e instanceof Error ? e.message : String(e)
    return { ok: false, message: m }
  }
}
