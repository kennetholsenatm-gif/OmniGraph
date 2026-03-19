# Launch DevSecOps stack: merged core compose (IAM, messaging, tooling, ChatOps) matches secrets-bootstrap.ps1.
# Prefer Ansible to start stacks (env injected by Ansible; no .env).
# Export secrets from Vault into the environment, then run this script from docker-compose/.
# File list: stack-manifest.json (see scripts/verify-stack-manifest.ps1).
param(
    [switch]$IncludeSdnTelemetry,
    [switch]$SkipLlm,
    [switch]$IncludeDiscovery,
    [switch]$IncludeAiOrchestration,
    [switch]$IncludeIdentity,
    [switch]$IncludeSiem
)
$ErrorActionPreference = "Stop"

$here = Split-Path -Parent $MyInvocation.MyCommand.Path
. (Join-Path $here "DevSecOpsStackManifest.ps1")

$envFile = Join-Path (Split-Path -Parent $here) ".env"
$envFileArg = @()
if (Test-Path $envFile) {
    $envFileArg = @("--env-file", $envFile)
    Write-Host "Using repo-root .env (prefer Ansible for env injection; see docs/VARLOCK_USAGE.md)."
} else {
    Write-Host "No repo-root .env; using current environment (export from Vault per VARLOCK_USAGE.md, or use Ansible to start stacks)."
}

$includeSdn = $IncludeSdnTelemetry -or ($env:DEVSECOPS_INCLUDE_SDN_TELEMETRY -eq "1")
$includeDiscovery = $IncludeDiscovery -or ($env:DEVSECOPS_INCLUDE_DISCOVERY -eq "1")
$includeAi = $IncludeAiOrchestration -or ($env:DEVSECOPS_INCLUDE_AI_ORCHESTRATION -eq "1")
$includeId = $IncludeIdentity -or ($env:DEVSECOPS_INCLUDE_IDENTITY -eq "1")
$includeSiem = $IncludeSiem -or ($env:DEVSECOPS_INCLUDE_SIEM -eq "1")
$includeLlm = -not $SkipLlm -and ($env:DEVSECOPS_INCLUDE_LLM -ne "0")

$manifest = Get-DevSecOpsStackManifest -ComposeDirectory $here
$coreFiles = Get-MergedCoreComposeFileList -Manifest $manifest -IncludeSdnTelemetry:$includeSdn

if ($includeSdn) {
    Write-Host "Including SDN + telemetry compose files (Linux SDN host recommended)."
}
Write-Host "Starting $($manifest.coreStack.label) (single merged compose run, --remove-orphans)..."
Invoke-DevSecOpsDockerComposeUp -ComposeFiles $coreFiles -EnvFileArg $envFileArg -RemoveOrphans -WorkingDirectory $here
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

if ($includeLlm) {
    $llmFiles = Get-OptionalStackComposeFiles -Manifest $manifest -StackKey "llm"
    $nf = $manifest.optionalStacks.llm.nonFatal -eq $true
    Write-Host "Starting LLM (BitNet inference, optional)..."
    Invoke-DevSecOpsDockerComposeUp -ComposeFiles $llmFiles -EnvFileArg $envFileArg -WorkingDirectory $here
    if ($LASTEXITCODE -ne 0) {
        if ($nf) {
            Write-Warning "LLM stack failed (exit $LASTEXITCODE); ensure llm_net exists: .\scripts\create-networks.ps1."
        } else {
            exit $LASTEXITCODE
        }
    }
}

if ($includeDiscovery) {
    $df = Get-OptionalStackComposeFiles -Manifest $manifest -StackKey "discovery"
    Invoke-DevSecOpsDockerComposeUp -ComposeFiles $df -EnvFileArg $envFileArg -WorkingDirectory $here
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

if ($includeAi) {
    $af = Get-OptionalStackComposeFiles -Manifest $manifest -StackKey "aiOrchestration"
    Invoke-DevSecOpsDockerComposeUp -ComposeFiles $af -EnvFileArg $envFileArg -WorkingDirectory $here
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

if ($includeId) {
    $idf = Get-OptionalStackComposeFiles -Manifest $manifest -StackKey "identity"
    Invoke-DevSecOpsDockerComposeUp -ComposeFiles $idf -EnvFileArg $envFileArg -WorkingDirectory $here
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

if ($includeSiem) {
    $sf = Get-OptionalStackComposeFiles -Manifest $manifest -StackKey "siem"
    Write-Host "Starting Wazuh SIEM (ensure docker-compose/siem/wazuh-config exists; see docs/WAZUH_SIEM.md)..."
    Invoke-DevSecOpsDockerComposeUp -ComposeFiles $sf -EnvFileArg $envFileArg -WorkingDirectory $here
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

Write-Host "Stack started. Check: docker ps"
