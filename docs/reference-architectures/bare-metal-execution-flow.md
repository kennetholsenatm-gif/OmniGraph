# Bare-Metal Execution Flow (Reference)

This is an example execution flow for bare-metal onboarding with OmniGraph.

## Flow

1. Validate infrastructure intent and credentials references
2. Verify BMC/network reachability
3. Start temporary provisioning services
4. Boot targets through configured boot mechanism
5. Wait for OS readiness and collect inventory
6. Run configuration management
7. Run verification and publish artifacts

## Example inventory snippet

```ini
[baremetal]
server-a ansible_host=<TARGET_NODE_IP_A>
server-b ansible_host=<TARGET_NODE_IP_B>
```

## Environment assumptions

Replace assumptions with your own standards:

- CI platform: `<CI_PLATFORM>`
- Artifact host: `https://artifacts.example.com`
- Git provider: `https://git.example.com/<org>/<repo>`
