package omnistate

import (
	"context"
	"strings"
)

// DetectNormalizer picks a normalizer from content type and filename heuristics.
func DetectNormalizer(contentType, name string) Normalizer {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	n := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.Contains(ct, "yaml"), strings.Contains(ct, "x-yaml"),
		strings.HasSuffix(n, ".yml"), strings.HasSuffix(n, ".yaml"):
		return AnsibleYAMLNormalizer{}
	case strings.Contains(ct, "json"), strings.HasSuffix(n, ".tfstate"), strings.HasSuffix(n, ".json"):
		// Heuristic: tfstate or explicit json → terraform state attempt
		if strings.Contains(n, "tfstate") || strings.HasSuffix(n, ".tfstate") {
			return TerraformNormalizer{}
		}
		// Plain .json could be tfstate or other; try terraform first (valid state has "values")
		return TerraformNormalizer{}
	default:
		if strings.HasSuffix(n, ".ini") || strings.Contains(ct, "plain") {
			return AnsibleININormalizer{}
		}
		return AnsibleININormalizer{}
	}
}

// NormalizeOne runs detection and normalization for a single artifact.
func NormalizeOne(ctx context.Context, in NormalizerInput) OmniGraphStateFragment {
	n := DetectNormalizer(in.ContentType, in.Name)
	fr, err := n.Normalize(ctx, in)
	if err != nil {
		return OmniGraphStateFragment{
			PartialErrors: []NormalizeError{{
				Path:    in.Name,
				Code:    "E_NORMALIZE",
				Message: err.Error(),
			}},
		}
	}
	return fr
}
