# Inventory Sources

OmniGraph can aggregate inventory context from multiple sources to support planning,
graph generation, and post-apply operations.

## Common Inputs

- IaC state outputs
- Static or generated inventory files
- CMDB/device APIs
- Runtime telemetry snapshots

## Contract Reference

Use versioned schema contracts from `schemas/` for machine validation and exchange.
Keep source-specific field mappings in environment documentation, not in core docs.
