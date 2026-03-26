# Bare-Metal Provisioning Execution Flow

## Overview

This document details the step-by-step execution flow for bare-metal provisioning through OmniGraph, from the initial `omnigraph pipeline run` command to the successful SSH connection.

## Execution Pipeline

### Phase 1: Initialization and Validation

```
┌─────────────────────────────────────────────────────────────┐
│ omnigraph pipeline run deploy-infrastructure                │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Parse .omnigraph.schema                                     │
│ - Validate schema syntax                                    │
│ - Load bare-metal target definitions                        │
│ - Resolve profile references                                │
│ - Validate BMC credentials (check Vault)                    │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Validate Network Configuration                              │
│ - Check DHCP/ProxyDHCP configuration                        │
│ - Verify HTTP boot server accessibility                     │
│ - Validate DNS resolution                                   │
│ - Check VLAN configuration                                  │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Validate BMC Connectivity                                   │
│ - Test BMC reachability (ping/HTTPS)                        │
│ - Verify BMC credentials                                    │
│ - Check BMC firmware version                                │
│ - Validate Redfish/IPMI access                              │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Generate Provisioning Artifacts                             │
│ - Generate iPXE scripts for each target                     │
│ - Generate Cloud-Init user-data/meta-data                   │
│ - Generate Ignition configs (if CoreOS)                     │
│ - Create provisioning timeline                              │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Phase 1 Complete                                            │
│ - Schema validation: PASSED                                 │
│ - Network validation: PASSED                                │
│ - BMC connectivity: PASSED                                  │
│ - Artifacts generated: SUCCESS                              │
└─────────────────────────────────────────────────────────────┘
```

### Phase 2: Pre-flight Checks

```
┌─────────────────────────────────────────────────────────────┐
│ Start Network Services                                      │
│ - Start ProxyDHCP service                                   │
│ - Start HTTP boot server                                    │
│ - Verify services are listening                             │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Verify Target State                                         │
│ - Check target power state (should be OFF)                  │
│ - Verify boot order configuration                           │
│ - Check disk configuration                                  │
│ - Validate network interfaces                               │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Pre-flight Checks Complete                                  │
│ - ProxyDHCP: RUNNING                                        │
│ - HTTP Boot Server: RUNNING                                 │
│ - Target state: VERIFIED                                    │
└─────────────────────────────────────────────────────────────┘
```

### Phase 3: Hardware Preparation

```
┌─────────────────────────────────────────────────────────────┐
│ For Each Target (Parallel Execution)                        │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 3.1 Discover Hardware                                        │
│ - Connect to BMC via Redfish/IPMI                           │
│ - Retrieve hardware inventory                               │
│ - Identify disk devices                                     │
│ - Get MAC addresses                                         │
│ - Check firmware versions                                   │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 3.2 Update Firmware (if policy = latest)                    │
│ - Download firmware from repository                         │
│ - Apply BIOS update                                         │
│ - Apply BMC firmware update                                 │
│ - Reboot BMC if required                                    │
│ - Verify firmware version                                   │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 3.3 Configure RAID                                          │
│ - Identify disks for RAID                                   │
│ - Create RAID array                                         │
│ - Verify RAID status                                        │
│ - Create logical volumes                                    │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 3.4 Configure Network Boot                                  │
│ - Set boot order (HTTP > Disk)                              │
│ - Configure network interface                               │
│ - Set PXE/HTTP boot parameters                              │
│ - Verify boot configuration                                 │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 3.5 Power On and Trigger Boot                               │
│ - Power on server via BMC                                   │
│ - Monitor boot sequence                                     │
│ - Verify DHCP Discover received                             │
│ - Verify ProxyDHCP response sent                            │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Hardware Preparation Complete                               │
│ - Discovery: SUCCESS                                        │
│ - Firmware: UPDATED                                         │
│ - RAID: CONFIGURED                                          │
│ - Network Boot: CONFIGURED                                  │
│ - Power State: ON                                           │
└─────────────────────────────────────────────────────────────┘
```

### Phase 4: Network Boot Sequence

```
┌─────────────────────────────────────────────────────────────┐
│ 4.1 DHCP Discovery                                          │
│ Target broadcasts DHCP Discover                             │
│ Corporate DHCP responds with IP address                     │
│ ProxyDHCP responds with boot parameters                     │
│ - next-server: HTTP boot server IP                          │
│ - bootfile: /ipxe.efi (or HTTP boot URL)                    │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 4.2 HTTP Boot                                               │
│ UEFI firmware initiates HTTP boot                           │
│ Downloads iPXE binary from HTTP server                      │
│ Executes iPXE                                              │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 4.3 iPXE Script Loading                                     │
│ iPXE requests boot script                                   │
│ GET /boot/?mac=00:11:22:33:44:55                            │
│ HTTP server generates dynamic script                        │
│ Script includes kernel, initrd, kernel args                 │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 4.4 OS Installer Boot                                       │
│ iPXE loads kernel and initrd                                │
│ Kernel boots with parameters                                │
│ - Static IP configuration                                   │
│ - Cloud-Init URL                                            │
│ - Auto-install parameters                                   │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 4.5 Cloud-Init Configuration                                │
│ OS installer fetches Cloud-Init                             │
│ GET /cloud-init/user-data?mac=00:11:22:33:44:55             │
│ GET /cloud-init/meta-data?mac=00:11:22:33:44:55             │
│ Cloud-Init applies configuration                            │
│ - Hostname                                                  │
│ - Users and SSH keys                                        │
│ - Packages                                                  │
│ - Run commands (install agent)                              │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 4.6 OS Installation                                         │
│ Installer runs automated installation                       │
│ - Partition disks                                           │
│ - Install OS packages                                       │
│ - Configure bootloader                                      │
│ - Apply Cloud-Init configuration                            │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Network Boot Complete                                       │
│ - DHCP: SUCCESS                                             │
│ - HTTP Boot: SUCCESS                                        │
│ - iPXE: SUCCESS                                             │
│ - Cloud-Init: SUCCESS                                       │
│ - OS Installation: SUCCESS                                  │
└─────────────────────────────────────────────────────────────┘
```

### Phase 5: First Boot and Agent Registration

```
┌─────────────────────────────────────────────────────────────┐
│ 5.1 First Boot                                              │
│ Server reboots into installed OS                            │
│ OS loads with configured network                            │
│ Cloud-Init runs on first boot                               │
│ - Configure network interfaces                              │
│ - Start SSH service                                         │
│ - Install OmniGraph agent                                   │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 5.2 Agent Installation                                      │
│ Cloud-Init executes agent installer                         │
│ curl -fsSL https://omnigraph.example.com/agent/install.sh   │
│ Agent binary downloaded and installed                       │
│ Agent starts and registers with OmniGraph                  │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 5.3 Agent Heartbeat                                         │
│ Agent sends heartbeat to OmniGraph                          │
│ POST /api/v1/agent/heartbeat                                │
│ - Target ID                                                 │
│ - IP address                                                │
│ - OS version                                                │
│ - Agent version                                             │
│ OmniGraph marks target as READY                             │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 5.4 Wait for All Targets                                    │
│ Wait for all targets to report READY                        │
│ Timeout if any target fails                                 │
│ Collect IP addresses from all targets                       │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ First Boot Complete                                         │
│ - OS Boot: SUCCESS                                          │
│ - Agent Install: SUCCESS                                    │
│ - Agent Registration: SUCCESS                               │
│ - All Targets: READY                                        │
└─────────────────────────────────────────────────────────────┘
```

### Phase 6: State Interception and Inventory

```
┌─────────────────────────────────────────────────────────────┐
│ 6.1 Collect Target Information                              │
│ Query agent for target details                              │
│ - IP addresses                                              │
│ - Hostnames                                                 │
│ - OS information                                            │
│ - Hardware information                                      │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 6.2 Update Inventory                                        │
│ Generate Ansible inventory                                  │
│ [baremetal]                                                 │
│ server-01 ansible_host=192.168.1.10                         │
│ server-02 ansible_host=192.168.1.11                         │
│ server-03 ansible_host=192.168.1.12                         │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 6.3 Emit Graph                                              │
│ Generate omnigraph/graph/v1 JSON                            │
│ Include new bare-metal nodes                                │
│ Add relationships to existing infrastructure                │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 6.4 Stop Network Services                                   │
│ Stop ProxyDHCP service                                      │
│ Stop HTTP boot server                                       │
│ Clean up temporary files                                    │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ State Interception Complete                                 │
│ - Inventory: GENERATED                                      │
│ - Graph: EMITTED                                            │
│ - Services: STOPPED                                         │
└─────────────────────────────────────────────────────────────┘
```

### Phase 7: Configuration Management (Ansible)

```
┌─────────────────────────────────────────────────────────────┐
│ 7.1 Run Ansible Playbooks                                   │
│ ansible-playbook -i inventory.ini site.yml                  │
│ Apply configuration management                              │
│ - Install additional packages                               │
│ - Configure services                                        │
│ - Apply security hardening                                  │
│ - Configure monitoring                                      │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 7.2 Verify Configuration                                    │
│ Run Ansible in check mode                                   │
│ Verify all changes applied                                  │
│ Check service status                                        │
│ Validate configuration files                                │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Configuration Management Complete                           │
│ - Ansible: SUCCESS                                          │
│ - Configuration: APPLIED                                    │
│ - Verification: PASSED                                      │
└─────────────────────────────────────────────────────────────┘
```

### Phase 8: Verification and Cleanup

```
┌─────────────────────────────────────────────────────────────┐
│ 8.1 Health Checks                                           │
│ Run health checks on all targets                            │
│ - SSH connectivity                                          │
│ - Service status                                            │
│ - Network connectivity                                      │
│ - Application health                                        │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 8.2 Update CMDB                                             │
│ Sync with NetBox                                            │
│ - Update device records                                     │
│ - Update IP addresses                                       │
│ - Update device status                                      │
│ - Add provisioning metadata                                 │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 8.3 Generate Report                                         │
│ Create deployment report                                    │
│ - Target list with status                                   │
│ - Configuration summary                                     │
│ - Health check results                                      │
│ - Timeline                                                  │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ 8.4 Send Notifications                                      │
│ Notify stakeholders                                         │
│ - Slack notification                                        │
│ - Email report                                              │
│ - Update ticketing system                                   │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Pipeline Complete                                           │
│ - Health Checks: PASSED                                     │
│ - CMDB: UPDATED                                             │
│ - Report: GENERATED                                         │
│ - Notifications: SENT                                       │
│                                                            │
│ ════════════════════════════════════════════════════════ │
│ BARE-METAL PROVISIONING SUCCESSFUL                         │
│ ════════════════════════════════════════════════════════ │
└─────────────────────────────────────────────────────────────┘
```

## Timeline Example

For a 3-server deployment:

```
Time    Phase                    Duration   Status
------  -----------------------  ---------  ------
00:00   Pipeline Start           -          STARTED
00:01   Phase 1: Validation      2m         SUCCESS
00:03   Phase 2: Pre-flight      1m         SUCCESS
00:04   Phase 3: Hardware Prep   15m        SUCCESS
        - Discovery              2m
        - Firmware Update        5m
        - RAID Config            3m
        - Network Boot Config    2m
        - Power On               3m
00:19   Phase 4: Network Boot    20m        SUCCESS
        - DHCP                   30s
        - HTTP Boot              1m
        - iPXE                   30s
        - OS Install             15m
        - Cloud-Init             3m
00:39   Phase 5: First Boot      10m        SUCCESS
        - OS Boot                2m
        - Agent Install          5m
        - Agent Registration     3m
00:49   Phase 6: State Intercept 2m         SUCCESS
        - Collect Info           1m
        - Generate Inventory     30s
        - Emit Graph             30s
00:51   Phase 7: Configuration   15m        SUCCESS
        - Ansible Playbooks      15m
01:06   Phase 8: Verification    5m         SUCCESS
        - Health Checks          2m
        - CMDB Update            2m
        - Report Generation      1m
01:11   Pipeline Complete        -          SUCCESS
```

## Error Handling

### Common Error Scenarios

1. **BMC Unreachable**
   - Retry with exponential backoff
   - Check network connectivity
   - Verify BMC credentials
   - Alert operator

2. **DHCP Timeout**
   - Check ProxyDHCP service
   - Verify network configuration
   - Check VLAN settings
   - Review DHCP logs

3. **HTTP Boot Failure**
   - Verify HTTP server is running
   - Check iPXE binary availability
   - Verify network routing
   - Review server logs

4. **OS Installation Failure**
   - Check ISO/image availability
   - Verify Cloud-Init configuration
   - Review installer logs
   - Check disk space

5. **Agent Registration Failure**
   - Verify agent binary
   - Check network connectivity
   - Review agent logs
   - Verify API endpoint

### Retry Strategy

- **BMC Operations**: 3 retries with 10s backoff
- **Network Operations**: 5 retries with 5s backoff
- **HTTP Requests**: 3 retries with exponential backoff
- **SSH Connections**: 10 retries with 2s backoff

### Timeout Configuration

- **Phase 1 (Validation)**: 5 minutes
- **Phase 2 (Pre-flight)**: 2 minutes
- **Phase 3 (Hardware Prep)**: 30 minutes
- **Phase 4 (Network Boot)**: 60 minutes
- **Phase 5 (First Boot)**: 20 minutes
- **Phase 6 (State Intercept)**: 5 minutes
- **Phase 7 (Configuration)**: 30 minutes
- **Phase 8 (Verification)**: 10 minutes

**Total Pipeline Timeout**: 3 hours