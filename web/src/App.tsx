function App() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-slate-950 px-6 py-12 text-slate-100">
      <header className="max-w-xl text-center">
        <h1 className="text-3xl font-semibold tracking-tight text-white">
          OmniGraph
        </h1>
        <p className="mt-3 text-sm leading-relaxed text-slate-400">
          State-aware DevSecOps orchestration: OpenTofu, Ansible, and telemetry
          in one GitOps flow. This shell will host the dependency graph (D3/SVG)
          and local schema validation.
        </p>
      </header>

      <div
        className="mt-10 flex h-56 w-full max-w-lg items-center justify-center rounded-xl border border-dashed border-slate-600 bg-slate-900/50 text-sm text-slate-500"
        role="img"
        aria-label="Dependency graph placeholder"
      >
        Graph canvas placeholder (D3.js / SVG)
      </div>
    </div>
  )
}

export default App
