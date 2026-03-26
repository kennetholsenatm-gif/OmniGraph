import { AlertTriangle, ChevronDown, FileCode, FileJson, Lock, RefreshCw, ShieldAlert, TerminalSquare, Upload } from 'lucide-react'
import { useEffect, useMemo, useRef, useState, type ChangeEvent } from 'react'

import { validateOmnigraphText } from '../validateOmnigraph'
import { CodeBlock } from './CodeBlock'
import { CopyableCommand } from './CopyableCommand'
import { shellQuote } from './shellQuote'

function SchemaFieldReadOnly({ name, typeLabel, value }: { name: string; typeLabel: string; value: string }) {
  return (
    <div className="rounded-xl border border-gray-800 bg-gray-900/50 p-4 opacity-70">
      <div className="mb-3 flex items-center justify-between">
        <label className="font-mono text-sm text-gray-200">{name}</label>
        <span className="rounded border border-blue-500/20 bg-blue-500/20 px-2 py-0.5 text-[10px] font-bold text-blue-400">
          {typeLabel}
        </span>
      </div>
      <input
        type="text"
        readOnly
        value={value}
        className="w-full cursor-not-allowed rounded-lg border border-gray-800 bg-gray-950 px-3 py-2 font-mono text-sm text-gray-500"
      />
    </div>
  )
}

export type SchemaTabProps = {
  schemaText: string
  onSchemaChange: (value: string) => void
  schemaCliPath: string
  onSchemaCliPathChange: (value: string) => void
  schemaFileNameHint?: string
  onSchemaFileNameHintChange?: (value: string | undefined) => void
}

export function SchemaTab({
  schemaText,
  onSchemaChange,
  schemaCliPath,
  onSchemaCliPathChange,
  schemaFileNameHint,
  onSchemaFileNameHintChange,
}: SchemaTabProps) {
  const [debounced, setDebounced] = useState(schemaText)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [env, setEnv] = useState<'dev' | 'prod'>('dev')
  const [dbPort, setDbPort] = useState('5432')
  const [portError, setPortError] = useState('')

  useEffect(() => {
    const t = setTimeout(() => setDebounced(schemaText), 250)
    return () => clearTimeout(t)
  }, [schemaText])

  const documentValidation = useMemo(() => validateOmnigraphText(debounced), [debounced])

  const pathArg = shellQuote(schemaCliPath.trim() || '.omnigraph.schema')

  const handlePortChange = (e: ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value
    setDbPort(val)
    const n = Number(val)
    if (val === '' || Number.isNaN(n) || n < 1024 || n > 65535) {
      setPortError('Tutorial sample: port must be an integer between 1024 and 65535.')
    } else {
      setPortError('')
    }
  }

  const portNum = portError ? '5432' : dbPort

  const onPickFile = () => fileInputRef.current?.click()

  const onFileSelected = (e: ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0]
    e.target.value = ''
    if (!f) {
      return
    }
    const reader = new FileReader()
    reader.onload = () => {
      const text = typeof reader.result === 'string' ? reader.result : ''
      onSchemaChange(text)
      onSchemaFileNameHintChange?.(f.name)
    }
    reader.readAsText(f)
  }

  const downloadSchema = () => {
    const blob = new Blob([schemaText], { type: 'text/yaml;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = schemaFileNameHint?.endsWith('.yml') || schemaFileNameHint?.endsWith('.yaml')
      ? schemaFileNameHint
      : '.omnigraph.schema'
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="flex h-full min-h-0 w-full flex-col overflow-hidden lg:flex-row">
      <div className="flex w-full flex-col overflow-y-auto border-b border-gray-800 p-6 lg:w-1/2 lg:border-b-0 lg:border-r">
        <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h2 className="text-xl font-bold text-gray-100">.omnigraph.schema</h2>
            <p className="mt-1 text-sm text-gray-500">
              Validated in-browser (JSON Schema). Coercion output comes from the Go CLI — use the commands on the right in your
              terminal.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <input
              ref={fileInputRef}
              type="file"
              accept=".yaml,.yml,.json,text/plain,text/*"
              className="hidden"
              aria-hidden
              onChange={onFileSelected}
            />
            <button
              type="button"
              onClick={onPickFile}
              className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-200 hover:bg-gray-800"
            >
              <Upload size={16} aria-hidden />
              Open file
            </button>
            <button
              type="button"
              onClick={downloadSchema}
              className="rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-200 hover:bg-gray-800"
            >
              Download
            </button>
          </div>
        </div>
        {schemaFileNameHint ? (
          <p className="mb-2 font-mono text-xs text-gray-500">Loaded: {schemaFileNameHint}</p>
        ) : null}

        <label htmlFor="schema-cli-path" className="mb-1 text-xs font-medium text-gray-400">
          Path for CLI examples (on disk)
        </label>
        <input
          id="schema-cli-path"
          type="text"
          value={schemaCliPath}
          onChange={(e) => onSchemaCliPathChange(e.target.value)}
          spellCheck={false}
          className="mb-4 w-full rounded-lg border border-gray-800 bg-gray-950 px-3 py-2 font-mono text-sm text-gray-200 outline-none focus:ring-2 focus:ring-blue-500/40"
          placeholder=".omnigraph.schema"
        />

        <label htmlFor="mvp-schema-doc" className="mb-2 text-sm font-medium text-gray-300">
          Project document
        </label>
        <textarea
          id="mvp-schema-doc"
          spellCheck={false}
          value={schemaText}
          onChange={(e) => onSchemaChange(e.target.value)}
          className="mb-4 min-h-48 w-full resize-y rounded-lg border border-gray-800 bg-gray-900/80 p-3 font-mono text-sm text-gray-100 outline-none focus:ring-2 focus:ring-blue-500/40"
        />
        <div
          className={`mb-6 rounded-md border px-3 py-2 text-sm ${
            documentValidation.ok
              ? 'border-emerald-800 bg-emerald-950/40 text-emerald-200'
              : 'border-rose-800 bg-rose-950/40 text-rose-200'
          }`}
          role="status"
        >
          {documentValidation.ok ? 'Schema document valid.' : documentValidation.message}
        </div>

        <details className="rounded-lg border border-gray-800 bg-gray-900/40">
          <summary className="flex cursor-pointer list-none items-center gap-2 px-3 py-2 text-sm text-gray-400 [&::-webkit-details-marker]:hidden">
            <ChevronDown size={16} className="shrink-0 transition-transform [[open]_&]:rotate-180" aria-hidden />
            Tutorial sample fields (not from your file)
          </summary>
          <div className="space-y-6 border-t border-gray-800 p-4">
            <div className="flex rounded-lg border border-gray-800 bg-gray-900 p-1">
              <button
                type="button"
                onClick={() => setEnv('dev')}
                className={`rounded-md px-4 py-1.5 text-sm font-medium transition-all ${
                  env === 'dev' ? 'bg-blue-600 text-white shadow' : 'text-gray-400 hover:text-gray-200'
                }`}
              >
                Dev
              </button>
              <button
                type="button"
                onClick={() => setEnv('prod')}
                className={`rounded-md px-4 py-1.5 text-sm font-medium transition-all ${
                  env === 'prod' ? 'bg-rose-600 text-white shadow' : 'text-gray-400 hover:text-gray-200'
                }`}
              >
                Prod
              </button>
            </div>
            <SchemaFieldReadOnly name="AWS_REGION" typeLabel="STRING" value="us-east-1" />
            <div className="rounded-xl border border-gray-800 bg-gray-900/50 p-4">
              <div className="mb-3 flex items-center justify-between">
                <label className="font-mono text-sm text-gray-200">DB_PORT</label>
                <span className="rounded border border-blue-500/20 bg-blue-500/20 px-2 py-0.5 text-[10px] font-bold text-blue-400">
                  INT
                </span>
              </div>
              <input
                type="text"
                value={dbPort}
                onChange={handlePortChange}
                className={`w-full rounded-lg border bg-gray-950 px-3 py-2 font-mono text-sm text-gray-200 outline-none transition-all focus:ring-1 ${
                  portError ? 'border-rose-500 focus:ring-rose-500' : 'border-gray-700 focus:ring-blue-500'
                }`}
              />
              {portError ? (
                <p className="mt-2 flex items-center gap-1 text-xs text-rose-400">
                  <AlertTriangle size={12} aria-hidden />
                  {portError}
                </p>
              ) : null}
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-900/50 p-4">
              <div className="mb-3 flex items-center justify-between">
                <label className="flex items-center gap-2 font-mono text-sm text-gray-200">
                  DB_PASSWORD <Lock size={14} className="text-amber-400" aria-hidden />
                </label>
                <span className="rounded border border-amber-500/20 bg-amber-500/20 px-2 py-0.5 text-[10px] font-bold text-amber-400">
                  SECRET
                </span>
              </div>
              <div className="flex w-full cursor-not-allowed items-center gap-2 rounded-lg border border-gray-800 bg-gray-950 px-3 py-2 font-mono text-sm text-gray-500">
                <ShieldAlert size={14} aria-hidden />
                {env === 'dev' ? 'vault/dev/database/password' : 'vault/prod/database/password'}
              </div>
              <p className="mt-2 text-xs text-gray-500">Secrets are injected as env at execution time (ADR 003).</p>
            </div>
          </div>
        </details>
      </div>

      <div className="flex w-full flex-col overflow-y-auto bg-gray-900/30 p-6 lg:w-1/2">
        <h2 className="mb-2 text-lg font-bold text-gray-100">Run in terminal</h2>
        <p className="mb-4 text-xs text-gray-500">
          Run these from a directory where <code className="text-gray-400">{schemaCliPath || '.omnigraph.schema'}</code> exists.
          Adjust the path field on the left if needed.
        </p>

        <div className="mb-6 space-y-3">
          <CopyableCommand label="Validate" command={`omnigraph validate ${pathArg}`} />
          <CopyableCommand label="Coerce — all formats" command={`omnigraph coerce --format=all -f ${pathArg}`} />
          <CopyableCommand label="Coerce — Terraform tfvars JSON" command={`omnigraph coerce --format=tfvars -f ${pathArg}`} />
          <CopyableCommand label="Coerce — Ansible group_vars" command={`omnigraph coerce --format=groupvars -f ${pathArg}`} />
          <CopyableCommand label="Coerce — env lines" command={`omnigraph coerce --format=env -f ${pathArg}`} />
        </div>

        <h3 className="mb-2 flex items-center gap-2 text-sm font-semibold text-gray-300">
          <RefreshCw className="text-emerald-400" size={16} aria-hidden />
          Illustrative coercion snippets
        </h3>
        <p className="mb-4 text-xs text-gray-500">
          Fictional variables for onboarding only — not generated from your document. Use the CLI commands above for real output.
        </p>

        {portError ? (
          <div className="flex flex-col items-center justify-center gap-4 py-8 text-rose-400">
            <ShieldAlert size={40} className="opacity-50" aria-hidden />
            <p className="text-center text-sm">Fix the tutorial DB_PORT range to show sample blocks below.</p>
          </div>
        ) : (
          <div className="space-y-4">
            <CodeBlock title="OpenTofu (HCL)" icon={<FileCode size={16} className="text-purple-400" aria-hidden />}>
              {`variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "db_port" {
  type    = number
  default = ${portNum}
}

variable "db_password" {
  type      = string
  sensitive = true
}`}
            </CodeBlock>
            <CodeBlock title="Ansible (YAML)" icon={<FileJson size={16} className="text-blue-400" aria-hidden />}>
              {`---
aws_region: "us-east-1"
db_port: ${portNum}
db_password: "{{ omnigraph.vault.db_password }}"`}
            </CodeBlock>
            <CodeBlock title="Podman / Docker (.env)" icon={<TerminalSquare size={16} className="text-emerald-400" aria-hidden />}>
              {`AWS_REGION=us-east-1
DB_PORT=${portNum}
DB_PASSWORD=********`}
            </CodeBlock>
          </div>
        )}
      </div>
    </div>
  )
}
