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
