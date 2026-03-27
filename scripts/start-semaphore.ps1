#Requires -Version 5.1
<#
.SYNOPSIS
  Provisions Semaphore on Incus (LXC) via Ansible — no Docker.
.EXAMPLE
  .\scripts\start-semaphore.ps1
  .\scripts\start-semaphore.ps1 -Inventory inventory\lxc.example.yml
#>
param(
  [string]$Inventory = "inventory/lxc.example.yml",
  [string[]]$ExtraAnsibleArgs = @()
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location "$repoRoot\ansible"

if (Get-Command ansible-galaxy -ErrorAction SilentlyContinue) {
  ansible-galaxy collection install -r collections/requirements.yml | Out-Null
}

$ansibleArgs = @(
  "-i", $Inventory,
  "playbooks/deploy-semaphore-incus.yml",
  "-e", "lxd_become=false",
  "-e", "lxd_manage_daemon=false",
  "-e", "lxd_ensure_idmap=false",
  "-e", "lxd_incus_socket=/run/incus/unix.socket",
  "-e", 'lxd_apply_names=["devsecops-semaphore"]'
) + $ExtraAnsibleArgs

& ansible-playbook @ansibleArgs

Write-Host "Open Semaphore UI (default): http://127.0.0.1:3001"
