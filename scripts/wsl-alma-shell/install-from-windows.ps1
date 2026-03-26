param(
    [string]$Distro = "AlmaLinux-10",
    [switch]$SetDefaultDistro
)

$ErrorActionPreference = "Stop"
$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoScripts = (Resolve-Path $here).Path

wsl -d $Distro -e true 2>$null | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Host "WSL distro '$Distro' not found or failed to start. Install with (elevated PowerShell may be required):"
    Write-Host "  wsl --install -d $Distro"
    exit 1
}

if ($SetDefaultDistro) {
    wsl --set-default $Distro
    Write-Host "Default WSL distro set to $Distro"
}

$winRepoRoot = Split-Path -Parent (Split-Path -Parent $repoScripts)
# Forward slashes avoid PowerShell/native arg munging of backslashes before wslpath.
$winRepoForWSL = ($winRepoRoot -replace '\\', '/')
$unixRepo = (& wsl.exe -d $Distro -- wslpath -a -u $winRepoForWSL 2>$null | Out-String).Trim()
if ([string]::IsNullOrWhiteSpace($unixRepo)) {
    Write-Error "wslpath failed for repo root: $winRepoRoot"
}

$installSh = "$unixRepo/scripts/wsl-alma-shell/install-wsl-shell.sh"
& wsl.exe -d $Distro -- bash "$installSh"
Write-Host "Done. Optional: merge windows-terminal.profile.fragment.json into Windows Terminal settings."
