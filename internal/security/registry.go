package security

// All registers built-in modules (passive / read-only oriented).
var All = []Module{
	&T1082SystemInfo{},
	&SELinuxMode{},
	&FirewalldActive{},
	&SSHDPermitRoot{},
	&SysctlASLR{},
	&AuditdActive{},
	&CorePattern{},
	&IPv4Forward{},
	&PtraceScope{},
	&PasswdUIDZero{},
}
