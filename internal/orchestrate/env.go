package orchestrate

import (
	"encoding/json"
	"strconv"

	"github.com/kennetholsenatm-gif/omnigraph/internal/coerce"
)

// MergeExecutionEnv combines OMNIGRAPH_* env from coercion with TF_VAR_* derived from tfvars JSON.
func MergeExecutionEnv(art *coerce.Artifacts) map[string]string {
	if art == nil {
		return nil
	}
	out := make(map[string]string)
	for k, v := range art.Env {
		out[k] = v
	}
	for name, val := range art.TerraformTfvarsJSON {
		for k, v := range tfVarFromValue(name, val) {
			out[k] = v
		}
	}
	return out
}

func tfVarFromValue(name string, v any) map[string]string {
	key := "TF_VAR_" + name
	out := make(map[string]string)
	switch t := v.(type) {
	case string:
		out[key] = t
	case bool:
		out[key] = strconv.FormatBool(t)
	case float64:
		out[key] = strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		out[key] = strconv.Itoa(t)
	case int64:
		out[key] = strconv.FormatInt(t, 10)
	case []any, map[string]any:
		b, err := json.Marshal(v)
		if err != nil {
			return out
		}
		out[key] = string(b)
	case nil:
		// skip
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return out
		}
		out[key] = string(b)
	}
	return out
}
