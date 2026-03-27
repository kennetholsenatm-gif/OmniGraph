#Requires -Version 5.1
<#
.SYNOPSIS
  Runs the WSL bash helper to tune Incus/LXD bridge DNS (NAT + DHCP + upstream resolvers).
.EXAMPLE
  .\scripts\setup-incus-network-dns.ps1
  $env:INCUS_BRIDGE = "incusbr0"; $env:DNS_SERVERS = "1.1.1.1,8.8.8.8"; .\scripts\setup-incus-network-dns.ps1
#>
$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$unixPath = $repoRoot -replace '\\', '/'
if ($unixPath -match '^([A-Za-z]):') {
  $unixPath = '/mnt/' + $Matches[1].ToLower() + $unixPath.Substring(2)
}
wsl -e bash -lc "cd '$unixPath' && chmod +x ./scripts/setup-incus-network-dns.sh 2>/dev/null; ./scripts/setup-incus-network-dns.sh"
