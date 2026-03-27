# Shared helpers for stack-manifest.json (merged docker compose).
# Dot-source from docker-compose/*.ps1: . (Join-Path $PSScriptRoot 'DevSecOpsStackManifest.ps1')

function Get-DevSecOpsStackManifest {
    param(
        [Parameter(Mandatory = $true)]
        [string]$ComposeDirectory
    )
    $manifestPath = Join-Path $ComposeDirectory "stack-manifest.json"
    if (-not (Test-Path $manifestPath)) {
        throw "Stack manifest not found: $manifestPath"
    }
    Get-Content -LiteralPath $manifestPath -Raw -Encoding utf8 | ConvertFrom-Json
}

function Get-MergedCoreComposeFileList {
    param(
        [Parameter(Mandatory = $true)]
        [object]$Manifest,
        [switch]$IncludeSdnTelemetry
    )
    $list = New-Object System.Collections.Generic.List[string]
    if ($IncludeSdnTelemetry) {
        foreach ($f in $Manifest.sdnTelemetry.files) {
            $list.Add([string]$f)
        }
    }
    foreach ($f in $Manifest.coreStack.files) {
        $list.Add([string]$f)
    }
    return , $list.ToArray()
}

function Invoke-DevSecOpsDockerComposeUp {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$ComposeFiles,
        [string[]]$EnvFileArg = @(),
        [switch]$RemoveOrphans,
        [string]$WorkingDirectory
    )
    if ($ComposeFiles.Count -eq 0) {
        throw "Invoke-DevSecOpsDockerComposeUp: no compose files."
    }
    $dockerArgs = New-Object System.Collections.Generic.List[string]
    $dockerArgs.Add("compose")
    foreach ($f in $ComposeFiles) {
        $dockerArgs.Add("-f")
        $dockerArgs.Add($f)
    }
    foreach ($a in $EnvFileArg) { $dockerArgs.Add($a) }
    $dockerArgs.Add("up")
    $dockerArgs.Add("-d")
    if ($RemoveOrphans) {
        $dockerArgs.Add("--remove-orphans")
    }
    $argArray = $dockerArgs.ToArray()
    if ($WorkingDirectory) {
        Push-Location $WorkingDirectory
        try {
            & docker @argArray
        } finally {
            Pop-Location
        }
    } else {
        & docker @argArray
    }
}

function Get-OptionalStackComposeFiles {
    param(
        [Parameter(Mandatory = $true)]
        [object]$Manifest,
        [Parameter(Mandatory = $true)]
        [string]$StackKey
    )
    $opt = $Manifest.optionalStacks.$StackKey
    if (-not $opt) {
        throw "Unknown optional stack key: $StackKey"
    }
    $files = @()
    foreach ($f in $opt.composeFiles) { $files += [string]$f }
    return , $files
}
