package secrets

import "context"

// Resolver fetches secret material without persisting it to tfvars or .env files on disk.
type Resolver interface {
	Resolve(ctx context.Context, keys []string) (map[string]string, error)
}

// StaticResolver returns a fixed map (tests and local dry-runs only).
type StaticResolver map[string]string

// Resolve implements Resolver for StaticResolver.
func (s StaticResolver) Resolve(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		if v, ok := s[k]; ok {
			out[k] = v
		}
	}
	return out, nil
}

// Chain tries resolvers in order; first hit wins per key.
type Chain []Resolver

// Resolve implements Resolver.
func (c Chain) Resolve(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	need := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		need[k] = struct{}{}
	}
	for _, r := range c {
		if len(need) == 0 {
			break
		}
		remaining := make([]string, 0, len(need))
		for k := range need {
			remaining = append(remaining, k)
		}
		got, err := r.Resolve(ctx, remaining)
		if err != nil {
			return nil, err
		}
		for k, v := range got {
			if v == "" {
				continue
			}
			if _, ok := need[k]; !ok {
				continue
			}
			out[k] = v
			delete(need, k)
		}
	}
	return out, nil
}
