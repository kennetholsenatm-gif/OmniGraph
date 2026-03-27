# Bare-Metal Provisioning (Reference Architecture)

This document describes an example pattern for integrating bare-metal workflows with
OmniGraph orchestration. It is intentionally non-normative.

## Example pipeline phases

1. Validate schema and environment contracts
2. Prepare network boot artifacts
3. Trigger hardware lifecycle operations through BMC APIs
4. Handoff to post-provision configuration (for example Ansible)
5. Emit run, graph, and inventory artifacts

## Example placeholders

Use placeholders instead of local values in all automation snippets:

```yaml
apiVersion: omnigraph/ir/v1
kind: InfrastructureIntent
metadata:
  name: bare-metal-example
spec:
  targets:
    - id: server-a
      baremetal:
        bmc:
          type: redfish
          address: "<BMC_HOST_OR_IP>"
  components:
    - id: server-a-os
      componentType: omnigraph.baremetal.os
      config:
        hostname: "server-a.example.com"
        network:
          interfaces:
            - name: eth0
              ipAddress: "<NODE_IP_CIDR>"
              gateway: "<DEFAULT_GATEWAY_IP>"
```

## Notes

- Keep BMC credentials in a secret backend
- Use isolated provisioning networks where required
- Use organization-approved hardening and compliance controls
