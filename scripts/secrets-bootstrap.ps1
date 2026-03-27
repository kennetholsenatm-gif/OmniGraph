param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ArgsFromCaller
)

$deployScript = "C:\GiTeaRepos\Deploy\scripts\secrets-bootstrap.ps1"
if (-not (Test-Path $deployScript)) {
    throw "Moved script not found: $deployScript"
}

Write-Host "secrets-bootstrap.ps1 moved to Deploy repo -> $deployScript"
& $deployScript @ArgsFromCaller
