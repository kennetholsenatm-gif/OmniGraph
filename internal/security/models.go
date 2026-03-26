package security

// Document is omnigraph/security/v1 JSON.
type Document struct {
	APIVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Metadata   Metadata     `json:"metadata"`
	Spec       DocumentSpec `json:"spec"`
}

// Metadata describes the scan run.
type Metadata struct {
	GeneratedAt string `json:"generatedAt"`
	Target      string `json:"target,omitempty"`
	AnsibleHost string `json:"ansibleHost,omitempty"`
	Transport   string `json:"transport,omitempty"`
	Profile     string `json:"profile,omitempty"`
	Disclaimer  string `json:"disclaimer,omitempty"`
}

// DocumentSpec holds aggregated results.
type DocumentSpec struct {
	Summary ScanSummary    `json:"summary"`
	Results []ModuleResult `json:"results"`
}

// ScanSummary counts outcome buckets.
type ScanSummary struct {
	ModulesRun    int `json:"modulesRun"`
	Vulnerable    int `json:"vulnerable"`
	NotVulnerable int `json:"notVulnerable"`
	Errors        int `json:"errors"`
	NotApplicable int `json:"notApplicable"`
}

// ModuleResult is one check outcome.
type ModuleResult struct {
	ModuleID       string   `json:"moduleId"`
	TechniqueID    string   `json:"techniqueId"`
	TechniqueName  string   `json:"techniqueName"`
	Tactic         string   `json:"tactic"`
	Severity       string   `json:"severity"`
	Status         string   `json:"status"`
	Summary        string   `json:"summary"`
	Evidence       string   `json:"evidence,omitempty"`
	Remediation    string   `json:"remediation,omitempty"`
	ComplianceTags []string `json:"complianceTags,omitempty"`
}

// Status constants for ModuleResult.Status.
const (
	StatusVulnerable    = "vulnerable"
	StatusNotVulnerable = "not_vulnerable"
	StatusError         = "error"
	StatusNotApplicable = "not_applicable"
	StatusUnknown       = "unknown"
)

// Severity constants.
const (
	SeverityInfo     = "info"
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

const apiVersion = "omnigraph/security/v1"
const kind = "SecurityScan"

const defaultDisclaimer = "Authorized security validation only. Unauthorized scanning is unlawful."
