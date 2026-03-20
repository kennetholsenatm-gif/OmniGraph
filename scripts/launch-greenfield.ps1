param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ArgsFromCaller
)

$deployScript = "C:\GiTeaRepos\Deploy\scripts\launch-greenfield.ps1"
if (-not (Test-Path $deployScript)) {
    throw "Moved script not found: $deployScript"
}

Write-Host "launch-greenfield.ps1 moved to Deploy repo -> $deployScript"
& $deployScript @ArgsFromCaller
