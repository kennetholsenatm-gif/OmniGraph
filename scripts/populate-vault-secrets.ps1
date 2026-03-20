param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ArgsFromCaller
)

$deployScript = "C:\GiTeaRepos\Deploy\scripts\populate-vault-secrets.ps1"
if (-not (Test-Path $deployScript)) {
    throw "Moved script not found: $deployScript"
}

Write-Host "populate-vault-secrets.ps1 moved to Deploy repo -> $deployScript"
& $deployScript @ArgsFromCaller
