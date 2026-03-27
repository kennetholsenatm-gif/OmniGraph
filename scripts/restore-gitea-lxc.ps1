param(
    [string]$Inventory = "inventory/lxc.example.yml",
    [switch]$ComposeUp,
    [string]$IncusSocketPath = "/run/incus/unix.socket"
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$ansibleDir = Join-Path $repoRoot "ansible"

if (-not (Get-Command ansible-playbook -ErrorAction SilentlyContinue)) {
    throw "ansible-playbook not found in PATH."
}

Push-Location $ansibleDir
try {
    ansible-galaxy collection install -r collections/requirements.yml
    $extraVars = @(
        "lxd_apply_names=[`"devsecops-gitea`"]",
        "lxd_optional_stack_enable={gitea_lite:true}",
        "lxd_become=false",
        "lxd_manage_daemon=false",
        "lxd_ensure_idmap=false",
        "lxd_incus_socket=$IncusSocketPath",
        "lxd_install_docker_in_instance=true"
    )
    if ($ComposeUp) {
        $extraVars += "devsecops_lxc_compose_up=true"
    }
    foreach ($ev in $extraVars) {
        Write-Host "extra-var: $ev"
    }
    $args = @("-i", $Inventory, "playbooks/deploy-devsecops-lxc.yml")
    foreach ($ev in $extraVars) {
        $args += @("-e", $ev)
    }
    ansible-playbook @args
}
finally {
    Pop-Location
}

Write-Host "Gitea LXC restore command completed."
