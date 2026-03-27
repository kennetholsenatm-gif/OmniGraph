# Implementation Plan: Full-Stack Discovery & BOM

## Objective

Deliver autonomous full-stack discovery, centralized SoT, and RAG-optimized BOM documentation (HBOM, SBOM, Services BOM, CBOM) per Architecture Strategy Brief.

## Deliverables (cline_coder_agent)

| # | Item | Status / Location |
|---|------|-------------------|
| 1 | **ADR** | docs/architecture/ADR-001-fullstack-discovery-bom.md |
| 2 | **OpenTofu** | discovery_net in opentofu/main.tf; Solace queues defined on broker (not Tofu) or document in runbook |
| 3 | **Ansible** | ansible/playbooks/deploy-fullstack-discovery.yml (UFW default-deny, allowed ports, NetBox/Dependency-Track placeholders) |
| 4 | **Varlock** | fullstack-discovery.env.schema (@sensitive, @required for DB, API tokens, SNMP) |
| 5 | **Doc_Agent** | Wiki: chunked HBOM/SBOM/Services BOM tables; System Glossary; RAG chunking |

## Security Critique

- Verify HBOM/SBOM satisfy CMMC CM and SA.
- Ensure CBOM logic integrates with SBOM to flag legacy crypto (PQC readiness).

## Execution Order

1. Apply OpenTofu (discovery_net).
2. Run deploy-fullstack-discovery.yml with devsecops_secrets (Varlock).
3. Add docker-compose.discovery.yml if NetDISCO/NetBox DB/Redis need full stack.
4. Wire n8n workflows for BOM ingestion from NetBox + Dependency-Track.
5. Doc_Agent: publish BOM synthesis to repo Wiki after PR merge.
