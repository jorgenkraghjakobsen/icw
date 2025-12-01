package version

import (
	"fmt"
	"runtime"
)

// These variables are set at build time via -ldflags
var (
	// Version is the semantic version (e.g., "2.0.0")
	Version = "dev"

	// Commit is the git commit hash
	Commit = "unknown"

	// BuildDate is the build timestamp
	BuildDate = "unknown"

	// GoVersion is the Go version used to build
	GoVersion = runtime.Version()
)

// Info returns formatted version information
func Info() string {
	return fmt.Sprintf("icw version %s (%s)\nBuilt: %s\nCommit: %s\nGo: %s",
		Version, runtime.GOOS+"/"+runtime.GOARCH, BuildDate, Commit, GoVersion)
}

// Short returns a short version string
func Short() string {
	if Commit != "unknown" && len(Commit) > 7 {
		return fmt.Sprintf("%s (%s)", Version, Commit[:7])
	}
	return Version
}

// Full returns the full semantic version
func Full() string {
	return Version
}
