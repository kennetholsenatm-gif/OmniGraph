# ADR 006: QminiWasm-Core Agent Mesh Integration

## Status

Proposed

## Context

OmniGraph currently orchestrates infrastructure through a state engine that bridges declarative provisioning (OpenTofu), imperative configuration (Ansible), and telemetry (NetBox/Zabbix). However, complex routing decisions, policy evaluations, and edge ML inference require capabilities beyond traditional scripting.

QminiWasm-Core provides:
- **Deterministic WebAssembly execution** via WasmEdge for sandboxed ML inference
- **Ternary-weight inference** for memory-constrained edge nodes
- **Quantum-assisted routing** (QAOA-style) for combinatorial optimization
- **Zero-Trust Enclave Enrollment (ZTEE)** for secure agent lifecycle

## Decision

We will implement an **event-driven agent mesh** pattern (inspired by n8n + Solace) that allows OmniGraph to delegate complex decisions to QminiWasm agents through an asynchronous topic-based communication layer.

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     OmniGraph Control Plane                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    State Engine                           │  │
│  │  • OpenTofu Plan/Apply                                   │  │
│  │  • Ansible Playbooks                                     │  │
│  │  • Telemetry Collection                                  │  │
│  └─────────────────────┬────────────────────────────────────┘  │
│                        │                                        │
│  ┌─────────────────────▼────────────────────────────────────┐  │
│  │              Event Mesh Broker (Solace-style)            │  │
│  │  • Topic-based pub/sub                                   │  │
│  │  • Async message routing                                 │  │
│  │  • Dead letter queues                                    │  │
│  │  • Event persistence                                     │  │
│  └─────────────────────┬────────────────────────────────────┘  │
│                        │                                        │
└────────────────────────┼────────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  Wasm Agent 1   │ │  Wasm Agent 2   │ │  Wasm Agent N   │
│  (WasmEdge)     │ │  (WasmEdge)     │ │  (WasmEdge)     │
│                 │ │                 │ │                 │
│ • ML Inference  │ │ • Route Optimize│ │ • Policy Engine │
│ • Ternary-weight│ │ • QAOA routing  │ │ • Compliance    │
│ • ZTEE enclave  │ │ • ZTEE enclave  │ │ • ZTEE enclave  │
└─────────────────┘ └─────────────────┘ └─────────────────┘
```

### Key Components

#### 1. Event Mesh Broker (`internal/agentmesh/broker.go`)

The broker implements Solace-style async communication:

```go
type EventBroker interface {
    // Publish emits an event to a topic
    Publish(ctx context.Context, topic string, event Event) error
    
    // Subscribe registers a handler for topic events
    Subscribe(ctx context.Context, topic string, handler EventHandler) (Subscription, error)
    
    // Request sends a request and waits for response
    Request(ctx context.Context, topic string, payload interface{}, timeout time.Duration) (*Event, error)
    
    // SubscribeWithFilter subscribes with topic wildcards
    SubscribeWithFilter(ctx context.Context, pattern string, handler EventHandler) (Subscription, error)
}
```

**Topic Hierarchy:**
```
omnigraph.state.changed    # Infrastructure state changes
omnigraph.policy.evaluate  # Policy evaluation requests
omnigraph.route.optimize   # Route optimization requests
omnigraph.ml.inference     # ML inference requests
omnigraph.telemetry.alert  # Telemetry alerts
qminiwasm.agent.heartbeat  # Agent health checks
qminiwasm.agent.result     # Agent computation results
```

#### 2. Wasm Agent Lifecycle Manager (`internal/agentmesh/lifecycle.go`)

Manages WasmEdge sandboxed environments:

```go
type WasmAgentManager interface {
    // SpawnAgent creates a new Wasm agent with ZTEE enrollment
    SpawnAgent(ctx context.Context, config AgentConfig) (*Agent, error)
    
    // TerminateAgent gracefully stops an agent
    TerminateAgent(ctx context.Context, agentID string) error
    
    // LoadModule loads a Wasm module into an agent
    LoadModule(ctx context.Context, agentID string, modulePath string) error
    
    // ExecuteInference runs ML inference in sandboxed environment
    ExecuteInference(ctx context.Context, agentID string, input InferenceInput) (*InferenceResult, error)
    
    // GetAgentStatus returns agent health and metrics
    GetAgentStatus(ctx context.Context, agentID string) (*AgentStatus, error)
}
```

#### 3. State Interceptor (`internal/agentmesh/interceptor.go`)

Bridges OmniGraph telemetry with QminiWasm inference:

```go
type StateInterceptor interface {
    // InterceptStateChange captures infrastructure state changes
    InterceptStateChange(ctx context.Context, change StateChange) error
    
    // RegisterAgent registers an agent for specific state patterns
    RegisterAgent(agentID string, patterns []StatePattern) error
    
    // RouteToAgent routes state changes to appropriate agents
    RouteToAgent(ctx context.Context, change StateChange) ([]string, error)
    
    // CollectAgentResults gathers results from agents
    CollectAgentResults(ctx context.Context, agentIDs []string, timeout time.Duration) ([]AgentResult, error)
}
```

#### 4. Quantum Router (`internal/agentmesh/quantum.go`)

Integrates QminiWasm's quantum-assisted routing:

```go
type QuantumRouter interface {
    // OptimizeInventoryPath finds optimal inventory traversal
    OptimizeInventoryPath(ctx context.Context, nodes []InventoryNode) (*OptimizedPath, error)
    
    // OptimizeAnsibleExecution orders playbooks for minimal execution time
    OptimizeAnsibleExecution(ctx context.Context, playbooks []Playbook, hosts []Host) (*ExecutionPlan, error)
    
    // OptimizeNetworkTraffic shapes traffic based on telemetry
    OptimizeNetworkTraffic(ctx context.Context, telemetry []TelemetryData) (*TrafficPolicy, error)
}
```

### Schema Extensions

Extend `.omnigraph.schema` to support agent nodes:

```yaml
apiVersion: omnigraph/v1
kind: Project
metadata:
  name: my-project
spec:
  # Existing spec properties...
  
  agentMesh:
    enabled: true
    broker:
      type: "internal"  # or "solace", "nats", "rabbitmq"
      persistence: true
      deadLetterQueue: true
    
    agents:
      - id: "route-optimizer"
        type: "qminiwasm"
        module: "wasm/route-optimizer.wasm"
        resources:
          memory: "512Mi"
          cpu: "1"
        triggers:
          - topic: "omnigraph.state.changed"
            filter: "spec.network"
        
      - id: "policy-engine"
        type: "qminiwasm"
        module: "wasm/policy-engine.wasm"
        resources:
          memory: "256Mi"
          cpu: "0.5"
        triggers:
          - topic: "omnigraph.policy.evaluate"
        
      - id: "ml-inference"
        type: "qminiwasm"
        module: "wasm/ml-inference.wasm"
        resources:
          memory: "1Gi"
          cpu: "2"
        triggers:
          - topic: "omnigraph.ml.inference"
        
      - id: "quantum-router"
        type: "qminiwasm"
        module: "wasm/quantum-router.wasm"
        resources:
          memory: "2Gi"
          cpu: "4"
        triggers:
          - topic: "omnigraph.route.optimize"
```

### Implementation Plan

#### Phase 1: Core Infrastructure
1. Create `internal/agentmesh/` module structure
2. Implement `EventBroker` interface with internal message queue
3. Create `WasmAgentManager` with WasmEdge integration
4. Add ZTEE lifecycle hooks

#### Phase 2: Schema & Integration
5. Extend `.omnigraph.schema` with `agentMesh` section
6. Create `StateInterceptor` to bridge telemetry
7. Integrate with existing orchestrate pipeline
8. Add CLI commands for agent management

#### Phase 3: Quantum & ML
9. Implement `QuantumRouter` for QAOA optimization
10. Add ML inference capabilities
11. Create example Wasm modules
12. Performance benchmarking

#### Phase 4: Production Readiness
13. Add metrics and monitoring
14. Implement dead letter queues
15. Add multi-broker support (Solace, NATS)
16. Security hardening

### Consequences

**Positive:**
- Enables complex routing decisions without scripting
- Provides sandboxed ML inference at the edge
- Quantum-assisted optimization for infrastructure
- Event-driven architecture enables loose coupling
- Extensible agent system for future capabilities

**Negative:**
- Additional complexity in orchestration
- WasmEdge dependency for runtime
- Learning curve for event-driven patterns
- Potential latency in async communication

**Risks:**
- Wasm module security requires careful sandboxing
- Quantum routing may have performance overhead
- Event mesh requires reliable message delivery