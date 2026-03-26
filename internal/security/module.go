package security

import "context"

// Module is a passive posture check aligned with ATT&CK metadata.
// targetLinux is true when the target OS was detected as Linux (uname -s), so modules can skip with not_applicable otherwise.
type Module interface {
	ID() string
	TechniqueID() string
	TechniqueName() string
	Tactic() string
	Severity() string
	Check(ctx context.Context, h Host, targetLinux bool) ModuleResult
}
