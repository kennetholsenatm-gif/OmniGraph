package remotecmd

import "strings"

// ShellSingleQuote returns a POSIX-safe single-quoted string.
func ShellSingleQuote(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `'\''`) + `'`
}

// RemoteShC builds `/bin/sh -c 'exec ...'` for SSH session.Run.
func RemoteShC(argv []string) string {
	var inner strings.Builder
	inner.WriteString("exec")
	for _, a := range argv {
		inner.WriteByte(' ')
		inner.WriteString(ShellSingleQuote(a))
	}
	return "/bin/sh -c " + ShellSingleQuote(inner.String())
}
