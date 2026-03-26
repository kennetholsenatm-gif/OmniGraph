import { AlertTriangle, ArrowRight, CheckCircle, ChevronDown, Play, RefreshCw } from 'lucide-react'
import { useMemo, useState } from 'react'

import { CopyableCommand } from './CopyableCommand'
import { buildOrchestrateCommand, looksAbsoluteHostPath } from './pipelineCommand'

type StepDef = { id: number; title: string; tool: string }

const demoSteps: StepDef[] = [
  { id: 1, title: 'PR Opened', tool: 'GitHub' },
  { id: 2, title: 'Schema Validation', tool: 'OmniGraph Wasm' },
  { id: 3, title: 'Unified Plan', tool: 'Tofu + Ansible' },
  { id: 4, title: 'Phase 1: Provision', tool: 'OpenTofu' },
  { id: 5, title: 'State Handoff', tool: 'Memory Injection' },
  { id: 6, title: 'Phase 2: Configure', tool: 'Ansible' },
  { id: 7, title: 'CMDB Sync', tool: 'NetBox API' },
]

export type PipelineTabProps = {
  workdir: string
  onWorkdirChange: (v: string) => void
  ansibleRoot: string
  onAnsibleRootChange: (v: string) => void
  playbookRel: string
  onPlaybookRelChange: (v: string) => void
  playbookOverride: string
  onPlaybookOverrideChange: (v: string) => void
  usePlaybookOverride: boolean
  onUsePlaybookOverrideChange: (v: boolean) => void
  schema: string
  onSchemaChange: (v: string) => void
  tfBinary: string
  onTfBinaryChange: (v: string) => void
  planFile: string
  onPlanFileChange: (v: string) => void
  stateFile: string
  onStateFileChange: (v: string) => void
  runner: 'exec' | 'container'
  onRunnerChange: (v: 'exec' | 'container') => void
  containerRuntime: string
  onContainerRuntimeChange: (v: string) => void
  autoApprove: boolean
  onAutoApproveChange: (v: boolean) => void
  skipAnsible: boolean
  onSkipAnsibleChange: (v: boolean) => void
  graphOut: string
  onGraphOutChange: (v: string) => void
  telemetryFile: string
  onTelemetryFileChange: (v: string) => void
  iacEngine: string
  onIacEngineChange: (v: string) => void
  tofuImage: string
  onTofuImageChange: (v: string) => void
  ansibleImage: string
  onAnsibleImageChange: (v: string) => void
  pulumiImage: string
  onPulumiImageChange: (v: string) => void
}

export function PipelineTab(p: PipelineTabProps) {
  const [demoStage, setDemoStage] = useState(0)
  const [simError, setSimError] = useState(false)

  const command = useMemo(
    () =>
      buildOrchestrateCommand({
        workdir: p.workdir,
        schema: p.schema,
        ansibleRoot: p.usePlaybookOverride ? '' : p.ansibleRoot,
        playbookRel: p.playbookRel,
        playbookOverride: p.usePlaybookOverride ? p.playbookOverride : '',
        tfBinary: p.tfBinary,
        planFile: p.planFile,
        stateFile: p.stateFile,
        runner: p.runner,
        containerRuntime: p.containerRuntime,
        autoApprove: p.autoApprove,
        skipAnsible: p.skipAnsible,
        graphOut: p.graphOut,
        telemetryFile: p.telemetryFile,
        iacEngine: p.iacEngine,
        tofuImage: p.tofuImage,
        ansibleImage: p.ansibleImage,
        pulumiImage: p.pulumiImage,
      }),
    [
      p.workdir,
      p.schema,
      p.ansibleRoot,
      p.playbookRel,
      p.playbookOverride,
      p.usePlaybookOverride,
      p.tfBinary,
      p.planFile,
      p.stateFile,
      p.runner,
      p.containerRuntime,
      p.autoApprove,
      p.skipAnsible,
      p.graphOut,
      p.telemetryFile,
      p.iacEngine,
      p.tofuImage,
      p.ansibleImage,
      p.pulumiImage,
    ],
  )

  const containerSingleMountWarning =
    p.runner === 'container' &&
    !p.skipAnsible &&
    !p.ansibleRoot.trim() &&
    !p.usePlaybookOverride &&
    looksAbsoluteHostPath(p.playbookRel)

  const runDemo = () => {
    setDemoStage(1)
    let current = 1
    const interval = window.setInterval(() => {
      current++
      if (simError && current === 3) {
        setDemoStage(-1)
        window.clearInterval(interval)
        return
      }
      setDemoStage(current)
      if (current >= 7) {
        window.clearInterval(interval)
      }
    }, 1200)
  }

  const fieldClass =
    'w-full rounded-lg border border-gray-800 bg-gray-950 px-3 py-2 text-sm text-gray-200 outline-none focus:ring-1 focus:ring-blue-500'

  const containerOverrideWarning =
    p.runner === 'container' && !p.skipAnsible && p.usePlaybookOverride && looksAbsoluteHostPath(p.playbookOverride)

  return (
    <div className="relative flex h-full flex-col overflow-auto p-6 lg:p-8">
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-100">Orchestrate (CLI)</h2>
        <p className="mt-1 text-sm text-gray-500">
          Build a copy-paste command for <code className="text-gray-400">omnigraph orchestrate</code> (alias{' '}
          <code className="text-gray-400">pipeline</code>). Set an OpenTofu root and optional Ansible repo root for
          sibling checkouts; the CLI uses <code className="text-gray-400">--ansible-root</code> when both are set. Apply
          requires a TTY for confirmation unless you enable <code className="text-gray-400">--auto-approve</code>. Secrets
          stay in env only (ADR 003). See{' '}
          <a
            className="text-blue-400 underline-offset-2 hover:underline"
            href="https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/docs/architecture.md"
            target="_blank"
            rel="noreferrer"
          >
            docs/architecture.md
          </a>
          .
        </p>
      </div>

      <div className="mb-6 grid max-w-3xl grid-cols-1 gap-4 md:grid-cols-2">
        <div className="md:col-span-2">
          <label className="mb-1 block text-xs font-medium text-gray-400">OpenTofu root (--workdir, required)</label>
          <input
            className={fieldClass}
            value={p.workdir}
            onChange={(e) => p.onWorkdirChange(e.target.value)}
            placeholder="C:\GiTeaRepos\devsecops-pipeline\opentofu"
          />
        </div>
        <div className="md:col-span-2">
          <label className="mb-1 block text-xs font-medium text-gray-400">
            Ansible repo root (--ansible-root, optional)
          </label>
          <input
            className={fieldClass}
            value={p.ansibleRoot}
            onChange={(e) => p.onAnsibleRootChange(e.target.value)}
            placeholder="C:\GiTeaRepos\devsecops-pipeline\ansible"
            disabled={p.usePlaybookOverride}
          />
          <p className="mt-1 text-xs text-gray-600">
            When set, the generated command uses <code className="text-gray-500">--ansible-root</code> and{' '}
            <code className="text-gray-500">--playbook</code> is relative to that folder (e.g. <code className="text-gray-500">site.yml</code>
            ).
          </p>
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-400">
            Playbook (relative to Ansible root, or to workdir if no Ansible root)
          </label>
          <input
            className={fieldClass}
            value={p.playbookRel}
            onChange={(e) => p.onPlaybookRelChange(e.target.value)}
            disabled={p.skipAnsible || p.usePlaybookOverride}
          />
        </div>
        <div className="flex flex-col justify-end gap-2">
          <label className="flex cursor-pointer items-center gap-2 text-sm text-gray-300">
            <input
              type="checkbox"
              checked={p.usePlaybookOverride}
              onChange={(e) => p.onUsePlaybookOverrideChange(e.target.checked)}
              disabled={p.skipAnsible}
              className="rounded border-gray-700 bg-gray-800 text-blue-500"
            />
            Custom <code className="text-gray-500">--playbook</code> (literal path)
          </label>
          {p.usePlaybookOverride ? (
            <input
              className={fieldClass}
              value={p.playbookOverride}
              onChange={(e) => p.onPlaybookOverrideChange(e.target.value)}
              placeholder="Absolute or relative path passed verbatim"
              disabled={p.skipAnsible}
            />
          ) : null}
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-400">--schema</label>
          <input className={fieldClass} value={p.schema} onChange={(e) => p.onSchemaChange(e.target.value)} />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-400">--tf-binary</label>
          <input className={fieldClass} value={p.tfBinary} onChange={(e) => p.onTfBinaryChange(e.target.value)} />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-400">--runner</label>
          <select
            className={fieldClass}
            value={p.runner}
            onChange={(e) => p.onRunnerChange(e.target.value as 'exec' | 'container')}
          >
            <option value="exec">exec</option>
            <option value="container">container</option>
          </select>
        </div>
        {p.runner === 'container' ? (
          <>
            <div>
              <label className="mb-1 block text-xs font-medium text-gray-400">--container-runtime</label>
              <input
                className={fieldClass}
                value={p.containerRuntime}
                onChange={(e) => p.onContainerRuntimeChange(e.target.value)}
                placeholder="docker or podman"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-gray-400">--tofu-image</label>
              <input className={fieldClass} value={p.tofuImage} onChange={(e) => p.onTofuImageChange(e.target.value)} placeholder="optional" />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-gray-400">--ansible-image</label>
              <input
                className={fieldClass}
                value={p.ansibleImage}
                onChange={(e) => p.onAnsibleImageChange(e.target.value)}
                placeholder="optional"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-gray-400">--pulumi-image</label>
              <input
                className={fieldClass}
                value={p.pulumiImage}
                onChange={(e) => p.onPulumiImageChange(e.target.value)}
                placeholder="optional"
              />
            </div>
          </>
        ) : null}
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-400">--plan-file</label>
          <input className={fieldClass} value={p.planFile} onChange={(e) => p.onPlanFileChange(e.target.value)} />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-400">--state-file</label>
          <input className={fieldClass} value={p.stateFile} onChange={(e) => p.onStateFileChange(e.target.value)} />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-400">--graph-out</label>
          <input className={fieldClass} value={p.graphOut} onChange={(e) => p.onGraphOutChange(e.target.value)} placeholder="optional path" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-400">--telemetry-file</label>
          <input
            className={fieldClass}
            value={p.telemetryFile}
            onChange={(e) => p.onTelemetryFileChange(e.target.value)}
            placeholder="optional"
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-400">--iac-engine</label>
          <input className={fieldClass} value={p.iacEngine} onChange={(e) => p.onIacEngineChange(e.target.value)} placeholder="tofu or pulumi" />
        </div>
        <div className="flex flex-col gap-3 md:col-span-2">
          <label className="flex cursor-pointer items-center gap-2 text-sm text-gray-300">
            <input
              type="checkbox"
              checked={p.autoApprove}
              onChange={(e) => p.onAutoApproveChange(e.target.checked)}
              className="rounded border-gray-700 bg-gray-800 text-blue-500"
            />
            --auto-approve
          </label>
          <label className="flex cursor-pointer items-center gap-2 text-sm text-gray-300">
            <input
              type="checkbox"
              checked={p.skipAnsible}
              onChange={(e) => p.onSkipAnsibleChange(e.target.checked)}
              className="rounded border-gray-700 bg-gray-800 text-blue-500"
            />
            --skip-ansible
          </label>
        </div>
      </div>

      <div className="mb-8 max-w-3xl">
        <CopyableCommand label="Generated command" command={command} />
        {!p.workdir.trim() ? (
          <p className="mt-2 text-xs text-amber-600/90">Set --workdir before running; the command is still shown for editing.</p>
        ) : null}
        {containerSingleMountWarning ? (
          <p className="mt-2 text-xs text-amber-600/90">
            Container runner only mounts the OpenTofu workdir at <code className="text-gray-500">/workspace</code>. Use a
            playbook path under that tree (e.g. <code className="text-gray-500">..\ansible\site.yml</code>) or set an
            Ansible repo root so the CLI can mount it at <code className="text-gray-500">/ansible</code>.
          </p>
        ) : null}
        {containerOverrideWarning ? (
          <p className="mt-2 text-xs text-amber-600/90">
            Absolute <code className="text-gray-500">--playbook</code> outside the workdir may not exist inside the
            container; prefer the Ansible repo root fields or a path under <code className="text-gray-500">--workdir</code>.
          </p>
        ) : null}
      </div>

      <details className="rounded-lg border border-gray-800 bg-gray-900/40">
        <summary className="flex cursor-pointer list-none items-center gap-2 px-4 py-3 text-sm text-gray-400 [&::-webkit-details-marker]:hidden">
          <ChevronDown size={16} className="shrink-0 transition-transform [[open]_&]:rotate-180" aria-hidden />
          Timeline demo (non-functional animation)
        </summary>
        <div className="border-t border-gray-800 p-4">
          <div className="mb-4 flex flex-wrap items-center gap-4">
            <label className="flex cursor-pointer items-center gap-2 text-sm text-gray-400">
              <input
                type="checkbox"
                checked={simError}
                onChange={(e) => setSimError(e.target.checked)}
                className="rounded border-gray-700 bg-gray-800 text-rose-500"
              />
              Simulate schema error
            </label>
            <button
              type="button"
              onClick={runDemo}
              disabled={demoStage > 0 && demoStage !== -1}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-500 disabled:bg-gray-800 disabled:text-gray-500"
            >
              <Play size={16} aria-hidden />
              Run demo
            </button>
            <button
              type="button"
              onClick={() => setDemoStage(0)}
              className="p-2 text-gray-500 hover:text-gray-300"
              aria-label="Reset demo"
            >
              <RefreshCw size={18} />
            </button>
          </div>
          <div className="flex flex-wrap items-center justify-center gap-2 md:gap-4">
            {demoSteps.map((step, index) => {
              const isDone = demoStage > step.id
              const isActive = demoStage === step.id
              const isFailed = demoStage === -1 && step.id === 2
              const isPending = demoStage < step.id && demoStage !== -1

              let borderClass = 'border-gray-800 bg-gray-900'
              let iconClass = 'text-gray-600'

              if (isDone) {
                borderClass = 'border-emerald-500/50 bg-emerald-500/5'
                iconClass = 'text-emerald-500'
              } else if (isActive) {
                borderClass = 'border-blue-500 bg-blue-500/10 shadow-[0_0_15px_rgba(59,130,246,0.2)]'
                iconClass = 'text-blue-400'
              } else if (isFailed) {
                borderClass = 'border-rose-500 bg-rose-500/10'
                iconClass = 'text-rose-500'
              }

              return (
                <div key={step.id} className="flex items-center gap-2 md:gap-4">
                  <div className={`flex w-40 flex-col items-center rounded-xl border-2 p-3 text-center transition-all md:w-44 ${borderClass}`}>
                    <div className={`mb-2 ${iconClass} ${isActive && !isFailed ? 'animate-pulse' : ''}`}>
                      {isDone ? (
                        <CheckCircle size={24} aria-hidden />
                      ) : isFailed ? (
                        <AlertTriangle size={24} aria-hidden />
                      ) : (
                        <div className="flex h-6 w-6 items-center justify-center rounded-full border-2 border-current text-[10px] font-bold">
                          {step.id}
                        </div>
                      )}
                    </div>
                    <h3 className={`text-xs font-bold ${isPending ? 'text-gray-500' : 'text-gray-200'}`}>{step.title}</h3>
                    <p className="font-mono text-[10px] text-gray-500">{step.tool}</p>
                  </div>
                  {index < demoSteps.length - 1 ? (
                    <ArrowRight size={18} className={`hidden md:block ${isDone ? 'text-emerald-500' : 'text-gray-800'}`} aria-hidden />
                  ) : null}
                </div>
              )
            })}
          </div>
        </div>
      </details>
    </div>
  )
}
