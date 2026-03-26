package security

import (
	"context"
	"strings"
)

const maxEvidence = 4000

func clip(s string) string {
	if len(s) <= maxEvidence {
		return s
	}
	return s[:maxEvidence] + "…"
}

// T1082SystemInfo collects kernel identity (baseline discovery).
type T1082SystemInfo struct{}

func (T1082SystemInfo) ID() string            { return "T1082_system_info" }
func (T1082SystemInfo) TechniqueID() string   { return "T1082" }
func (T1082SystemInfo) TechniqueName() string { return "System Information Discovery" }
func (T1082SystemInfo) Tactic() string        { return "discovery" }
func (T1082SystemInfo) Severity() string      { return SeverityInfo }

func (T1082SystemInfo) Check(ctx context.Context, h Host, targetLinux bool) ModuleResult {
	if !targetLinux {
		return ModuleResult{ModuleID: T1082SystemInfo{}.ID(), TechniqueID: T1082SystemInfo{}.TechniqueID(), TechniqueName: T1082SystemInfo{}.TechniqueName(), Tactic: T1082SystemInfo{}.Tactic(), Severity: SeverityInfo, Status: StatusNotApplicable, Summary: "Target is not Linux; module skipped"}
	}
	out, stderr, code, err := h.Run(ctx, []string{"uname", "-a"})
	if err != nil {
		return ModuleResult{Status: StatusError, Summary: "failed to run uname", Evidence: err.Error()}
	}
	if code != 0 {
		return ModuleResult{Status: StatusError, Summary: "uname non-zero exit", Evidence: clip(stderr)}
	}
	return ModuleResult{
		Status:      StatusNotVulnerable,
		Summary:     "Kernel identity collected for baseline (expected on a scanned Linux host)",
		Evidence:    clip(strings.TrimSpace(out)),
		Remediation: "Restrict exposure of build/kernel details where organizational policy requires minimal disclosure.",
		ComplianceTags: []string{"baseline"},
	}
}

// SELinuxMode checks enforcing vs permissive/disabled.
type SELinuxMode struct{}

func (SELinuxMode) ID() string            { return "selinux_mode" }
func (SELinuxMode) TechniqueID() string { return "T1562.001" }
func (SELinuxMode) TechniqueName() string {
	return "Impair Defenses: Disable or Modify Tools"
}
func (SELinuxMode) Tactic() string   { return "defense_evasion" }
func (SELinuxMode) Severity() string { return SeverityHigh }

func (SELinuxMode) Check(ctx context.Context, h Host, targetLinux bool) ModuleResult {
	if !targetLinux {
		return na(SELinuxMode{}, "Target is not Linux")
	}
	out, _, code, err := h.Run(ctx, []string{"getenforce"})
	mode := strings.TrimSpace(strings.ToLower(out))
	if err != nil || code != 0 || mode == "" {
		out2, _, c2, _ := h.Run(ctx, []string{"cat", "/sys/fs/selinux/enforce"})
		v := strings.TrimSpace(out2)
		if c2 != 0 || v == "" {
			return ModuleResult{Status: StatusNotApplicable, Summary: "SELinux not present or not readable", Evidence: clip(out)}
		}
		if v == "1" {
			mode = "enforcing"
		} else {
			mode = "permissive"
		}
	}
	switch mode {
	case "enforcing":
		return ModuleResult{Status: StatusNotVulnerable, Summary: "SELinux is enforcing", Evidence: mode, Remediation: "Keep enforcing mode for RHEL workloads unless documented exception.", ComplianceTags: []string{"stig", "cis"}}
	case "permissive", "disabled":
		return ModuleResult{Status: StatusVulnerable, Summary: "SELinux is not enforcing", Evidence: mode, Remediation: "Set SELinux to enforcing; fix policy denials rather than lowering mode.", ComplianceTags: []string{"stig", "cis"}}
	default:
		return ModuleResult{Status: StatusUnknown, Summary: "Unexpected getenforce output", Evidence: mode}
	}
}

// FirewalldActive checks whether firewalld is active (host firewall).
type FirewalldActive struct{}

func (FirewalldActive) ID() string            { return "firewalld_active" }
func (FirewalldActive) TechniqueID() string   { return "T1562.004" }
func (FirewalldActive) TechniqueName() string { return "Impair Defenses: Disable or Modify System Firewall" }
func (FirewalldActive) Tactic() string        { return "defense_evasion" }
func (FirewalldActive) Severity() string      { return SeverityMedium }

func (FirewalldActive) Check(ctx context.Context, h Host, targetLinux bool) ModuleResult {
	if !targetLinux {
		return na(FirewalldActive{}, "Target is not Linux")
	}
	_, _, code, err := h.Run(ctx, []string{"systemctl", "is-active", "firewalld"})
	if err != nil {
		return ModuleResult{Status: StatusError, Summary: "systemctl failed", Evidence: err.Error()}
	}
	if code == 0 {
		return ModuleResult{Status: StatusNotVulnerable, Summary: "firewalld is active", Evidence: "active", Remediation: "Maintain host firewall rules aligned with change management.", ComplianceTags: []string{"cis"}}
	}
	return ModuleResult{Status: StatusVulnerable, Summary: "firewalld is not active", Evidence: "inactive", Remediation: "Enable and configure firewalld (or equivalent) consistent with network segmentation design.", ComplianceTags: []string{"cis"}}
}

// SSHDPermitRoot inspects sshd_config for PermitRootLogin yes/prohibit-password.
type SSHDPermitRoot struct{}

func (SSHDPermitRoot) ID() string            { return "sshd_permit_root" }
func (SSHDPermitRoot) TechniqueID() string   { return "T1021.001" }
func (SSHDPermitRoot) TechniqueName() string { return "Remote Services: SSH" }
func (SSHDPermitRoot) Tactic() string        { return "lateral_movement" }
func (SSHDPermitRoot) Severity() string     { return SeverityHigh }

func (SSHDPermitRoot) Check(ctx context.Context, h Host, targetLinux bool) ModuleResult {
	if !targetLinux {
		return na(SSHDPermitRoot{}, "Target is not Linux")
	}
	out, stderr, code, err := h.Run(ctx, []string{"grep", "-E", "^[[:space:]]*(#?)[Pp]ermit[Rr]oot[Ll]ogin", "/etc/ssh/sshd_config"})
	if err != nil {
		return ModuleResult{Status: StatusError, Summary: "grep failed", Evidence: err.Error()}
	}
	if code != 0 && code != 1 {
		return ModuleResult{Status: StatusError, Summary: "unexpected grep exit", Evidence: clip(stderr)}
	}
	if code == 1 || strings.TrimSpace(out) == "" {
		return ModuleResult{Status: StatusNotVulnerable, Summary: "No explicit PermitRootLogin (sshd default often deny root)", Evidence: "(no match)", Remediation: "Explicitly set PermitRootLogin no in sshd_config.", ComplianceTags: []string{"stig", "cis"}}
	}
	line := strings.ToLower(out)
	if strings.Contains(line, "without-password") || strings.Contains(line, "forced-commands-only") {
		return ModuleResult{Status: StatusNotVulnerable, Summary: "Root SSH limited (key/forced-commands)", Evidence: clip(strings.TrimSpace(out)), ComplianceTags: []string{"stig"}}
	}
	if strings.Contains(line, "permitrootlogin no") || strings.Contains(line, "permitrootlogin prohibit-password") {
		return ModuleResult{Status: StatusNotVulnerable, Summary: "Root SSH login restricted", Evidence: clip(strings.TrimSpace(out)), ComplianceTags: []string{"stig", "cis"}}
	}
	if strings.Contains(line, "permitrootlogin yes") {
		return ModuleResult{Status: StatusVulnerable, Summary: "PermitRootLogin allows password/root SSH", Evidence: clip(strings.TrimSpace(out)), Remediation: "Set PermitRootLogin no; use sudo and SSH keys.", ComplianceTags: []string{"stig", "cis"}}
	}
	return ModuleResult{Status: StatusUnknown, Summary: "Could not classify PermitRootLogin", Evidence: clip(strings.TrimSpace(out))}
}

// SysctlASLR checks /proc/sys/kernel/randomize_va_space.
type SysctlASLR struct{}

func (SysctlASLR) ID() string            { return "sysctl_aslr" }
func (SysctlASLR) TechniqueID() string   { return "T1068" }
func (SysctlASLR) TechniqueName() string { return "Exploitation for Privilege Escalation" }
func (SysctlASLR) Tactic() string       { return "privilege_escalation" }
func (SysctlASLR) Severity() string     { return SeverityMedium }

func (SysctlASLR) Check(ctx context.Context, h Host, targetLinux bool) ModuleResult {
	if !targetLinux {
		return na(SysctlASLR{}, "Target is not Linux")
	}
	out, _, code, err := h.Run(ctx, []string{"cat", "/proc/sys/kernel/randomize_va_space"})
	if err != nil || code != 0 {
		return ModuleResult{Status: StatusError, Summary: "cannot read randomize_va_space", Evidence: clip(out)}
	}
	v := strings.TrimSpace(out)
	if v == "2" {
		return ModuleResult{Status: StatusNotVulnerable, Summary: "ASLR enabled (randomize_va_space=2)", Evidence: v, Remediation: "Keep kernel.randomize_va_space=2.", ComplianceTags: []string{"cis", "stig"}}
	}
	if v == "1" || v == "0" {
		return ModuleResult{Status: StatusVulnerable, Summary: "ASLR not fully enabled", Evidence: v, Remediation: "Set kernel.randomize_va_space=2 via sysctl.d.", ComplianceTags: []string{"cis", "stig"}}
	}
	return ModuleResult{Status: StatusUnknown, Summary: "Unexpected randomize_va_space", Evidence: v}
}

// AuditdActive checks whether auditd is active.
type AuditdActive struct{}

func (AuditdActive) ID() string            { return "auditd_active" }
func (AuditdActive) TechniqueID() string   { return "T1562.001" }
func (AuditdActive) TechniqueName() string { return "Impair Defenses: Disable or Modify Tools" }
func (AuditdActive) Tactic() string        { return "defense_evasion" }
func (AuditdActive) Severity() string    { return SeverityMedium }

func (AuditdActive) Check(ctx context.Context, h Host, targetLinux bool) ModuleResult {
	if !targetLinux {
		return na(AuditdActive{}, "Target is not Linux")
	}
	_, _, code, err := h.Run(ctx, []string{"systemctl", "is-active", "auditd"})
	if err != nil {
		return ModuleResult{Status: StatusError, Summary: "systemctl failed", Evidence: err.Error()}
	}
	if code == 0 {
		return ModuleResult{Status: StatusNotVulnerable, Summary: "auditd is active", Evidence: "active", Remediation: "Keep audit rules aligned with organizational auditing standard.", ComplianceTags: []string{"stig", "cis"}}
	}
	return ModuleResult{Status: StatusVulnerable, Summary: "auditd is not active", Evidence: "inactive", Remediation: "Enable auditd and deploy audisp/audit rules per compliance baseline.", ComplianceTags: []string{"stig", "cis"}}
}

// CorePattern inspects kernel.core_pattern for risky user-mode helpers.
type CorePattern struct{}

func (CorePattern) ID() string            { return "core_pattern" }
func (CorePattern) TechniqueID() string   { return "T1562.001" }
func (CorePattern) TechniqueName() string { return "Impair Defenses: Disable or Modify Tools" }
func (CorePattern) Tactic() string        { return "defense_evasion" }
func (CorePattern) Severity() string      { return SeverityLow }

func (CorePattern) Check(ctx context.Context, h Host, targetLinux bool) ModuleResult {
	if !targetLinux {
		return na(CorePattern{}, "Target is not Linux")
	}
	out, _, code, err := h.Run(ctx, []string{"cat", "/proc/sys/kernel/core_pattern"})
	if err != nil || code != 0 {
		return ModuleResult{Status: StatusError, Summary: "cannot read core_pattern", Evidence: clip(out)}
	}
	p := strings.TrimSpace(out)
	if !strings.HasPrefix(p, "|") {
		return ModuleResult{Status: StatusNotVulnerable, Summary: "core_pattern does not pipe to a helper", Evidence: clip(p), ComplianceTags: []string{"baseline"}}
	}
	if strings.Contains(p, "systemd-coredump") {
		return ModuleResult{Status: StatusNotVulnerable, Summary: "core dumps handled by systemd-coredump", Evidence: clip(p), ComplianceTags: []string{"baseline"}}
	}
	return ModuleResult{Status: StatusVulnerable, Summary: "core_pattern pipes to a custom helper", Evidence: clip(p), Remediation: "Review core helper; prefer distribution-managed coredump handling.", ComplianceTags: []string{"cis"}}
}

// IPv4Forward checks IP forwarding sysctl.
type IPv4Forward struct{}

func (IPv4Forward) ID() string            { return "ipv4_forward" }
func (IPv4Forward) TechniqueID() string   { return "T1562.001" }
func (IPv4Forward) TechniqueName() string { return "Impair Defenses: Disable or Modify Tools" }
func (IPv4Forward) Tactic() string       { return "defense_evasion" }
func (IPv4Forward) Severity() string      { return SeverityMedium }

func (IPv4Forward) Check(ctx context.Context, h Host, targetLinux bool) ModuleResult {
	if !targetLinux {
		return na(IPv4Forward{}, "Target is not Linux")
	}
	out, _, code, err := h.Run(ctx, []string{"cat", "/proc/sys/net/ipv4/ip_forward"})
	if err != nil || code != 0 {
		return ModuleResult{Status: StatusError, Summary: "cannot read ip_forward", Evidence: clip(out)}
	}
	v := strings.TrimSpace(out)
	if v == "0" {
		return ModuleResult{Status: StatusNotVulnerable, Summary: "IPv4 forwarding disabled", Evidence: v, Remediation: "Keep forwarding off on non-router hosts.", ComplianceTags: []string{"cis", "stig"}}
	}
	if v == "1" {
		return ModuleResult{Status: StatusVulnerable, Summary: "IPv4 forwarding enabled", Evidence: v, Remediation: "Disable unless this host is a documented router.", ComplianceTags: []string{"cis", "stig"}}
	}
	return ModuleResult{Status: StatusUnknown, Summary: "Unexpected ip_forward value", Evidence: v}
}

// PtraceScope checks YAMA ptrace scope.
type PtraceScope struct{}

func (PtraceScope) ID() string            { return "ptrace_scope" }
func (PtraceScope) TechniqueID() string   { return "T1055" }
func (PtraceScope) TechniqueName() string { return "Process Injection" }
func (PtraceScope) Tactic() string       { return "defense_evasion" }
func (PtraceScope) Severity() string     { return SeverityMedium }

func (PtraceScope) Check(ctx context.Context, h Host, targetLinux bool) ModuleResult {
	if !targetLinux {
		return na(PtraceScope{}, "Target is not Linux")
	}
	out, _, code, err := h.Run(ctx, []string{"cat", "/proc/sys/kernel/yama/ptrace_scope"})
	if err != nil || code != 0 {
		return ModuleResult{Status: StatusNotApplicable, Summary: "YAMA ptrace_scope not available", Evidence: clip(out)}
	}
	v := strings.TrimSpace(out)
	if v == "1" || v == "2" || v == "3" {
		return ModuleResult{Status: StatusNotVulnerable, Summary: "ptrace restricted", Evidence: v, Remediation: "Keep ptrace_scope>=1 on servers.", ComplianceTags: []string{"cis"}}
	}
	if v == "0" {
		return ModuleResult{Status: StatusVulnerable, Summary: "ptrace_scope allows broad ptrace", Evidence: v, Remediation: "Set kernel.yama.ptrace_scope=1 (or stricter) unless debugging requires exception.", ComplianceTags: []string{"cis"}}
	}
	return ModuleResult{Status: StatusUnknown, Summary: "Unexpected ptrace_scope", Evidence: v}
}

// PasswdUIDZero detects multiple UID 0 accounts in /etc/passwd.
type PasswdUIDZero struct{}

func (PasswdUIDZero) ID() string            { return "passwd_uid_zero" }
func (PasswdUIDZero) TechniqueID() string   { return "T1078" }
func (PasswdUIDZero) TechniqueName() string { return "Valid Accounts" }
func (PasswdUIDZero) Tactic() string        { return "persistence" }
func (PasswdUIDZero) Severity() string     { return SeverityCritical }

func (PasswdUIDZero) Check(ctx context.Context, h Host, targetLinux bool) ModuleResult {
	if !targetLinux {
		return na(PasswdUIDZero{}, "Target is not Linux")
	}
	out, stderr, code, err := h.Run(ctx, []string{"awk", "-F:", "$3==0 {print}", "/etc/passwd"})
	if err != nil {
		return ModuleResult{Status: StatusError, Summary: "awk failed", Evidence: err.Error()}
	}
	if code != 0 {
		return ModuleResult{Status: StatusError, Summary: "cannot read passwd", Evidence: clip(stderr)}
	}
	lines := nonEmptyLines(out)
	if len(lines) <= 1 {
		return ModuleResult{Status: StatusNotVulnerable, Summary: "Single UID 0 account (expected root)", Evidence: clip(strings.Join(lines, "\n")), ComplianceTags: []string{"stig", "cis"}}
	}
	return ModuleResult{Status: StatusVulnerable, Summary: "Multiple UID 0 accounts", Evidence: clip(strings.Join(lines, "\n")), Remediation: "Remove duplicate UID 0 entries; use sudo for elevation.", ComplianceTags: []string{"stig", "cis"}}
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func na(m Module, reason string) ModuleResult {
	return ModuleResult{
		ModuleID:      m.ID(),
		TechniqueID:   m.TechniqueID(),
		TechniqueName: m.TechniqueName(),
		Tactic:        m.Tactic(),
		Severity:      SeverityInfo,
		Status:        StatusNotApplicable,
		Summary:       reason,
	}
}
