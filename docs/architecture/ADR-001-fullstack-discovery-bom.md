# ADR-001: Full-Stack Discovery & Comprehensive BOM Generation

## Status

Proposed. Orchestrator-selected stack to be implemented by deploy-fullstack-discovery playbook and OpenTofu.

## Context

Automated, full-stack discovery and inventory must support:

- **HBOM** — Hardware Bill of Materials (network devices, physical/virtual hosts).
- **SBOM** — Software Bill of Materials (container images, app dependencies).
- **Services BOM (SaaSBOM/API)** — Service endpoints, APIs, and dependencies.
- **CBOM** — Cryptographic Bill of Materials (primitives in discovered software); must integrate with SBOM for PQC readiness.

Requirements: containerizable tools, programmatic API access for n8n/SAM, operation on an isolated Docker bridge network.

## Decision (Orchestrator Selection)

| Concern | Selected Tool(s) | Justification |
|--------|------------------|----------------|
| **Network & HBOM discovery** | **NetDISCO** | Containerizable, SNMP/LLDP, API; runZero/Scanopy as optional augment. |
| **SBOM & services discovery** | **Syft** + **Trivy** | Syft for SBOM generation from images/filesystems; Trivy for vuln scanning and SBOM; both CLI/API, container-friendly. |
| **SBOM aggregation & vuln tracking** | **Dependency-Track** | Central SBOM store, API, CVE mapping; integrates with Syft/Trivy output. |
| **Source of Truth (SoT)** | **NetBox** | DCIM/IPAM, API-first, containerizable; Device42/DefectDojo as optional (DefectDojo for vuln mapping in SoT). |

## Integration (Unified HBOM, SBOM, Services BOM)

1. **NetDISCO** → Discovers network topology and devices → Pushes device/host inventory to **NetBox** via API (HBOM SoT).
2. **Syft/Trivy** → Run against container images and artifact stores → SBOM (CycloneDX/SPDX) → Ingested into **Dependency-Track** (SBOM SoT).
3. **Services BOM** → n8n workflows consume NetBox (devices/services) and Dependency-Track (components/APIs) APIs; optional service mesh or API catalog export.
4. **CBOM** → Existing CBOM generation logic consumes SBOM from Dependency-Track; flags legacy crypto in dependencies (PQC readiness).

## CMMC 2.0 Mapping

- **CM (Configuration Management)** — HBOM/SBOM provide configuration and component baseline; SoT supports change control.
- **SA (System and Services Acquisition)** — SBOM and vuln data support supply chain and acquisition risk assessment.

## Compliance (Critique Agent)

- Security Critique phase must verify: HBOM/SBOM coverage for CM/SA controls; CBOM integration with SBOM for legacy crypto flagging (PQC readiness).

## References

- deploy-fullstack-discovery.yml (Ansible), OpenTofu discovery networks, fullstack-discovery.env.schema (Varlock).
- Doc_Agent: BOM synthesis (chunked HBOM, SBOM, Services BOM tables); System Glossary; RAG-optimized chunking.
