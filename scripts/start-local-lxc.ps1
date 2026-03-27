param(
    [string]$Inventory = "inventory/lxc.example.yml",
    [string]$ApplyNamesJson = '["devsecops-iam","devsecops-messaging"]'
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$ansibleDir = Join-Path $repoRoot "ansible"

Write-Host "Starting local LXC target via Ansible..."
Write-Host "Inventory: $Inventory"
Write-Host "Instances: $ApplyNamesJson"

Push-Location $ansibleDir
try {
    ansible-galaxy collection install -r collections/requirements.yml
    ansible-playbook -i $Inventory playbooks/deploy-devsecops-lxc.yml -e "lxd_apply_names=$ApplyNamesJson"
}
finally {
    Pop-Location
}
