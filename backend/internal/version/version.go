// Package version holds build-time metadata injected via -ldflags.
package version

// Version, Commit, and BuildDate are populated at build time by mage
// (see go.yaml -> build.versionPkg / crossBuild.versionPkg).
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)
