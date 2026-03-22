package version

// Version is set at build time via -ldflags (e.g. -X .../version.Version=1.0.0).
var Version = "0.0.0-dev"

// String returns the semantic version string for the control plane binary.
func String() string {
	return Version
}
