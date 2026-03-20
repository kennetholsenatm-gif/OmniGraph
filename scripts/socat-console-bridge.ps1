<#
.SYNOPSIS
  Run socat inside WSL: ISR serial device <-> TCP 127.0.0.1 (for Ansible telnet, not SSH).

.PARAMETER WslDevice
  Serial device in WSL, e.g. /dev/ttyUSB0 (after usbipd-win attach) or /dev/ttyS2.

.PARAMETER Port
  Local TCP listen port (default 3322).

.PARAMETER Baud
  Console baud (default 9600).

.EXAMPLE
  usbipd wsl attach --busid 1-5
  .\scripts\socat-console-bridge.ps1 -WslDevice /dev/ttyUSB0
#>
param(
  [string] $WslDevice = "/dev/ttyUSB0",
  [int] $Port = 3322,
  [int] $Baud = 9600
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command wsl.exe -ErrorAction SilentlyContinue)) {
  Write-Error "WSL not found. Use a Linux host with scripts/socat-console-bridge.sh or install WSL2."
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$drive = $repoRoot.Substring(0, 1).ToLowerInvariant()
$tail = $repoRoot.Substring(2) -replace '\\', '/'
$wslRepo = "/mnt/$drive$tail"

# Single-line -lc argument (newlines break PowerShell -> wsl argument parsing).
$bashOneLine = "export ISR_CONSOLE_DEVICE='$WslDevice' ISR_CONSOLE_PORT=$Port ISR_CONSOLE_BAUD=$Baud ISR_CONSOLE_BIND=127.0.0.1; cd '$wslRepo'; chmod +x scripts/socat-console-bridge.sh 2>/dev/null; exec ./scripts/socat-console-bridge.sh"

Write-Host "WSL socat: 127.0.0.1:$Port <-> $WslDevice (${Baud} baud)"
Write-Host "Repo (WSL): $wslRepo"
Write-Host "Test:  wsl -e bash -lc `"nc -v 127.0.0.1 $Port`""
Write-Host ""

& wsl.exe -e bash -lc $bashOneLine
