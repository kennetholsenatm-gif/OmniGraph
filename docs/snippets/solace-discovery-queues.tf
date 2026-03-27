# Paste into your infra repo OpenTofu root (not applied from this repository).
locals {
  solace_vpn_discovery = "discovery_tools_mesh"

  discovery_topics = {
    sbom_ingested  = "sample-project/discovery/sbom/ingested/v1"
    vuln_critical  = "sample-project/discovery/vuln/critical/v1"
    netbox_change  = "sample-project/discovery/netbox/change/v1"
    inventory_full = "sample-project/discovery/inventory/full/v1"
  }

  discovery_queues = {
    bom_ingest    = "Q.SAMPLE.BOM.INGEST"
    netbox_events = "Q.SAMPLE.NETBOX.EVENTS"
    dlq           = "Q.SAMPLE.DISCOVERY.DLQ"
  }
}
