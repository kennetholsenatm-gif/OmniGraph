package schemas

import _ "embed"

// OmniGraphSchemaJSON is the canonical JSON Schema for .omnigraph.schema documents.
//
//go:embed omnigraph.schema.json
var OmniGraphSchemaJSON []byte

// GraphV1SchemaJSON is the JSON Schema for omnigraph/graph/v1 documents.
//
//go:embed graph.v1.schema.json
var GraphV1SchemaJSON []byte

// SecurityV1SchemaJSON is the JSON Schema for omnigraph/security/v1 documents.
//
//go:embed security.v1.schema.json
var SecurityV1SchemaJSON []byte

// LockV1SchemaJSON is the JSON Schema for omnigraph/lock/v1 state lease documents.
//
//go:embed lock.v1.schema.json
var LockV1SchemaJSON []byte

// InventorySourceV1SchemaJSON is the JSON Schema for omnigraph/inventory-source/v1 snapshots.
//
//go:embed inventory-source.v1.schema.json
var InventorySourceV1SchemaJSON []byte

// RunV1SchemaJSON is the JSON Schema for omnigraph/run/v1 pipeline run timelines.
//
//go:embed run.v1.schema.json
var RunV1SchemaJSON []byte

// IrV1SchemaJSON is the JSON Schema for omnigraph/ir/v1 InfrastructureIntent documents.
//
//go:embed ir.v1.schema.json
var IrV1SchemaJSON []byte
