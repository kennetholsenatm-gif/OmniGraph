# OmniGraph glossary

Short definitions for terms used across OmniGraph docs and the root README. For product narrative, see [product-philosophy.md](product-philosophy.md).

| Term | Meaning |
|------|---------|
| **Graph-first truth model** | Infrastructure relationships are represented as explicit nodes and edges in a graph, instead of being implied only by script order or pipeline stages. |
| **GitOps pipeline context** | Plan, apply, and CI job context shown alongside the graph so changes are reviewable as intent and topology, not only as log output. |
| **Ansible handoff** | The step where declared graph/schema/inventory context is passed into Ansible-driven convergence, reducing ad-hoc glue between IaC and configuration management. |
| **`omnigraph/graph/v1`** | JSON shape consumed by the Visualizer for interactive exploration of nodes and edges. |
| **`.omnigraph.schema`** | Project schema document edited in the Schema Contract UI; defines contracts and checks for your OmniGraph project. |
| **`omnigraph/security/v1`** | Shape for security and compliance posture data kept next to the graph story in the Posture tab. |
| **WASM-backed HCL** | In the Web IDE, WebAssembly is used to give fast feedback while editing Terraform-flavored HCL. This is an **editor/runtime** concern for IaC text, not related to edge ML runtimes such as QMiniWasm. |
