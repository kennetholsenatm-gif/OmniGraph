# Integrations

OmniGraph supports optional integration with secret stores, CMDB systems, telemetry
sources, and identity providers.

## Integration Categories

- Secret backends (runtime retrieval, no committed credentials)
- Inventory and CMDB sources
- Telemetry enrichment for graph and run context
- Identity and authorization providers

## Provider Neutrality

Provider names in examples are illustrative. Teams can map OmniGraph workflows to
their selected stack (for example GitHub, GitLab, Gitea, or self-hosted CI).

Use placeholders in examples:

- `https://git.example.com/<org>/<repo>`
- `https://id.example.com`
- `https://inventory.example.com/api`
