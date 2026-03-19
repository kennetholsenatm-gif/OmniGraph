# Paste into your infra repo OpenTofu (not applied from devsecops-pipeline opentofu/ root).
locals {
  solace_vpn_discovery = "discovery_tools_mesh"

  discovery_topics = {
    sbom_ingested  = "devsecops/discovery/sbom/ingested/v1"
    vuln_critical  = "devsecops/discovery/vuln/critical/v1"
    netbox_change  = "devsecops/discovery/netbox/change/v1"
    inventory_full = "devsecops/discovery/inventory/full/v1"
  }

  discovery_queues = {
    bom_ingest    = "Q.DEVSECOPS.BOM.INGEST"
    netbox_events = "Q.DEVSECOPS.NETBOX.EVENTS"
    dlq           = "Q.DEVSECOPS.DISCOVERY.DLQ"
  }
}
