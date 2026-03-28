import { AlertTriangle, ArrowRight, CheckCircle, Play, RefreshCw } from 'lucide-react'
import { useMemo, useState } from 'react'

import { tryParseGraphDocument } from '../graph/graphConventions'
import { CopyableCommand } from './CopyableCommand'
import { buildOrchestrateCommand, looksAbsoluteHostPath } from './pipelineCommand'

type MatrixRow = {
  stage: string
  target: string
  runner: string
  ansible: string
  outputs: string
}

const SIM_STEPS = [
  { id: 1, title: 'Validate schema', hint: 'OmniGraph + schema file' },
  { id: 2, title: 'Coerce intent', hint: 'Tool inputs' },
  { id: 3, title: 'OpenTofu plan', hint: 'plan output' },
  { id: 4, title: 'Ansible check', hint: 'Check mode' },
  { id: 5, title: 'Approval gate', hint: 'TTY or --auto-approve' },
  { id: 6, title: 'OpenTofu apply', hint: 'Apply' },
  { id: 7, title: 'Ansible apply', hint: 'Apply mode' },
  { id: 8, title: 'Artifacts', hint: 'graph / telemetry' },
] as const

function trimCell(s: string): string {
  const t = s.trim()
  return t || '—'
}

function phaseBadgeClass(status: string): string {
  const u = status.toLowerCase()
  if (u === 'ok' || u === 'done' || u === 'success') {
    return 'border-emerald-500/40 bg-emerald-500/10 text-emerald-200/90'
  }
  if (u === 'pending' || u === 'running') {
    return 'border-amber-500/40 bg-amber-500/10 text-amber-200/90'
  }
  if (u === 'err' || u === 'error' || u === 'failed') {
    return 'border-rose-500/40 bg-rose-500/10 text-rose-200/90'
  }
  return 'border-gray-700 bg-gray-800/80 text-gray-300'
}

export type PipelineTabProps = {
  /** Graph JSON from Topology (optional); used to show `spec.phases` when valid. */
  graphText: string
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

  const graphDoc = useMemo(() => tryParseGraphDocument(p.graphText), [p.graphText])

  const ansibleDesc = useMemo(() => {
    if (p.skipAnsible) {
      return 'skipped (--skip-ansible)'
    }
    if (p.usePlaybookOverride) {
      return trimCell(p.playbookOverride) || '—'
    }
    const root = p.ansibleRoot.trim()
    const rel = p.playbookRel.trim() || 'site.yml'
    if (root) {
      return `${root} → ${rel}`
    }
    return rel
  }, [p.ansibleRoot, p.playbookOverride, p.playbookRel, p.skipAnsible, p.usePlaybookOverride])

  const runnerDesc = useMemo(() => {
    if (p.runner === 'container') {
      const rt = p.containerRuntime.trim() || 'container runtime'
      return `container (${rt})`
    }
    return 'exec (host)'
  }, [p.containerRuntime, p.runner])

  const outputsDesc = useMemo(() => {
    const parts: string[] = []
    if (p.graphOut.trim()) {
      parts.push(`--graph-out`)
    }
    if (p.telemetryFile.trim()) {
      parts.push(`--telemetry-file`)
    }
    return parts.length ? parts.join(', ') : '—'
  }, [p.graphOut, p.telemetryFile])

  const matrixRows: MatrixRow[] = useMemo(() => {
    const wd = trimCell(p.workdir)
    const sch = trimCell(p.schema)
    const plan = trimCell(p.planFile)
    const st = trimCell(p.stateFile)
    const iac = p.iacEngine.trim() ? p.iacEngine.trim() : 'OpenTofu/Terraform'
    return [
      {
        stage: 'Validate schema',
        target: sch,
        runner: '—',
        ansible: '—',
        outputs: '—',
      },
      {
        stage: 'Coerce intent',
        target: wd,
        runner: runnerDesc,
        ansible: '—',
        outputs: '—',
      },
      {
        stage: `${iac} plan`,
        target: `${wd} · state ${st}`,
        runner: runnerDesc,
        ansible: '—',
        outputs: plan,
      },
      {
        stage: 'Ansible check',
        target: wd,
        runner: runnerDesc,
        ansible: ansibleDesc,
        outputs: '—',
      },
      {
        stage: 'Approval gate',
        target: p.autoApprove ? 'non-interactive (--auto-approve)' : 'TTY confirmation',
        runner: '—',
        ansible: '—',
        outputs: '—',
      },
      {
        stage: `${iac} apply`,
        target: wd,
        runner: runnerDesc,
        ansible: '—',
        outputs: st,
      },
      {
        stage: 'Ansible apply',
        target: wd,
        runner: runnerDesc,
        ansible: ansibleDesc,
        outputs: '—',
      },
      {
        stage: 'Emit graph / run artifacts',
        target: wd,
        runner: '—',
        ansible: '—',
        outputs: outputsDesc,
      },
    ]
  }, [ansibleDesc, outputsDesc, p.autoApprove, p.iacEngine, p.planFile, p.schema, p.stateFile, p.workdir, runnerDesc])

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
      if (simError && current === 1) {
        setDemoStage(-1)
        window.clearInterval(interval)
        return
      }
      current++
      if (current > SIM_STEPS.length) {
        window.clearInterval(interval)
        return
      }
      setDemoStage(current)
    }, 1100)
  }

  const fieldClass =
    'w-full rounded-lg border border-gray-800 bg-gray-950 px-3 py-2 text-sm text-gray-200 outline-none focus:ring-1 focus:ring-blue-500'

  const containerOverrideWarning =
    p.runner === 'container' && !p.skipAnsible && p.usePlaybookOverride && looksAbsoluteHostPath(p.playbookOverride)

  const thClass = 'border-b border-gray-800 bg-gray-900/90 px-3 py-2 text-left text-[10px] font-semibold uppercase tracking-wide text-gray-500'
  const tdClass = 'border-b border-gray-800/80 px-3 py-2 align-top text-xs text-gray-300'

  return (
    <div className="relative flex h-full flex-col overflow-auto p-6 lg:p-8">
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-100">Orchestrate (CLI)</h2>
        <p className="mt-1 text-sm text-gray-500">
          Build a copy-paste command for <code className="text-gray-400">omnigraph orchestrate</code> (alias{' '}
          <code className="text-gray-400">pipeline</code>). The execution matrix below mirrors the orchestration chain in{' '}
          <code className="text-gray-400">docs/core-concepts/execution-matrix.md</code> and maps each stage to your targets,
          runner, Ansible model, and artifact outputs. Apply requires a TTY unless <code className="text-gray-400">--auto-approve</code>.
          See{' '}
          <a
            className="text-blue-400 underline-offset-2 hover:underline"
            href="/docs/architecture.md"
            target="_blank"
            rel="noreferrer"
          >
            docs/architecture.md
          </a>
          .
        </p>
      </div>

      <section className="mb-8 max-w-5xl rounded-xl border border-gray-800 bg-gray-900/35">
        <div className="border-b border-gray-800 px-4 py-3">
          <h3 className="text-sm font-semibold text-gray-200">Execution matrix</h3>
          <p className="mt-1 text-xs text-gray-500">
            Rows follow validate → coerce → plan → Ansible check → approval → apply → Ansible apply → artifacts. Columns show
            how your form fields land in that chain.
          </p>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[640px] border-collapse font-mono text-[11px]">
            <thead>
              <tr>
                <th className={thClass}>Stage</th>
                <th className={thClass}>Target</th>
                <th className={thClass}>Runner</th>
                <th className={thClass}>Ansible</th>
                <th className={thClass}>Outputs</th>
              </tr>
            </thead>
            <tbody>
              {matrixRows.map((row) => (
                <tr key={row.stage} className="hover:bg-gray-800/30">
                  <td className={`${tdClass} font-medium text-gray-200`}>{row.stage}</td>
                  <td className={tdClass}>{row.target}</td>
                  <td className={tdClass}>{row.runner}</td>
                  <td className={tdClass}>{row.ansible}</td>
                  <td className={tdClass}>{row.outputs}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {graphDoc?.spec.phases?.length ? (
          <div className="border-t border-gray-800 px-4 py-3">
            <p className="mb-2 text-[10px] font-semibold uppercase tracking-wide text-gray-500">
              Phases from Topology graph JSON <span className="font-mono font-normal normal-case text-gray-600">(spec.phases)</span>
            </p>
            <div className="flex flex-wrap gap-2">
              {graphDoc.spec.phases.map((ph) => (
                <span
                  key={ph.name}
                  title={ph.detail ?? undefined}
                  className={`rounded-md border px-2.5 py-1 text-xs font-medium ${phaseBadgeClass(ph.status)}`}
                >
                  <span className="font-mono text-[10px] text-gray-400">{ph.name}</span>
                  <span className="mx-1.5 text-gray-600">·</span>
                  {ph.status}
                </span>
              ))}
            </div>
          </div>
        ) : null}
      </section>

      <div className="mb-8 max-w-5xl rounded-xl border border-gray-800 bg-gray-900/25 p-4">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
          <div>
            <h3 className="text-sm font-semibold text-gray-200">Reconciliation flow (simulation)</h3>
            <p className="text-xs text-gray-500">Walks the same stage names as the matrix; does not run the CLI.</p>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <label className="flex cursor-pointer items-center gap-2 text-xs text-gray-400">
              <input
                type="checkbox"
                checked={simError}
                onChange={(e) => setSimError(e.target.checked)}
                className="rounded border-gray-700 bg-gray-800 text-rose-500"
              />
              Fail at validate
            </label>
            <button
              type="button"
              onClick={runDemo}
              disabled={demoStage > 0 && demoStage !== -1}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-500 disabled:bg-gray-800 disabled:text-gray-500"
            >
              <Play size={14} aria-hidden />
              Run simulation
            </button>
            <button
              type="button"
              onClick={() => setDemoStage(0)}
              className="p-2 text-gray-500 hover:text-gray-300"
              aria-label="Reset simulation"
            >
              <RefreshCw size={16} />
            </button>
          </div>
        </div>
        <div className="flex flex-wrap items-stretch gap-1 md:gap-0">
          {SIM_STEPS.map((step, index) => {
            const isDone = demoStage > step.id
            const isActive = demoStage === step.id
            const isFailed = demoStage === -1 && step.id === 1
            const isPending = demoStage < step.id && demoStage !== -1

            let box = 'border-gray-800 bg-gray-900/80'
            let accent = 'text-gray-600'
            if (isDone) {
              box = 'border-emerald-500/45 bg-emerald-500/5'
              accent = 'text-emerald-500'
            } else if (isActive) {
              box = 'border-blue-500/70 bg-blue-500/10 shadow-[0_0_12px_rgba(59,130,246,0.15)]'
              accent = 'text-blue-400'
            } else if (isFailed) {
              box = 'border-rose-500/55 bg-rose-500/10'
              accent = 'text-rose-500'
            }

            return (
              <div key={step.id} className="flex min-w-0 flex-1 items-center md:flex-initial">
                <div
                  className={`flex min-w-[100px] flex-1 flex-col rounded-lg border px-2 py-2 text-center transition-all md:min-w-[108px] ${box}`}
                >
                  <div className={`mb-1 flex justify-center ${accent} ${isActive && !isFailed ? 'animate-pulse' : ''}`}>
                    {isDone ? (
                      <CheckCircle size={18} aria-hidden />
                    ) : isFailed ? (
                      <AlertTriangle size={18} aria-hidden />
                    ) : (
                      <span className="flex h-5 w-5 items-center justify-center rounded-full border border-current text-[9px] font-bold">
                        {step.id}
                      </span>
                    )}
                  </div>
                  <p className={`text-[10px] font-bold leading-tight ${isPending ? 'text-gray-500' : 'text-gray-200'}`}>
                    {step.title}
                  </p>
                  <p className="mt-0.5 font-mono text-[9px] leading-tight text-gray-600">{step.hint}</p>
                </div>
                {index < SIM_STEPS.length - 1 ? (
                  <ArrowRight
                    className={`mx-0.5 hidden h-4 w-4 shrink-0 self-center md:block ${isDone ? 'text-emerald-600/80' : 'text-gray-800'}`}
                    aria-hidden
                  />
                ) : null}
              </div>
            )
          })}
        </div>
      </div>

      <div className="mb-6 grid max-w-3xl grid-cols-1 gap-4 md:grid-cols-2">
        <div className="md:col-span-2">
          <label className="mb-1 block text-xs font-medium text-gray-400">OpenTofu root (--workdir, required)</label>
          <input
            className={fieldClass}
            value={p.workdir}
            onChange={(e) => p.onWorkdirChange(e.target.value)}
            placeholder="C:\path\to\infrastructure\opentofu"
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
            placeholder="C:\path\to\infrastructure\ansible"
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
    </div>
  )
}
