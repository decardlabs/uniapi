package common

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"
)

// StartTime records the Unix timestamp when the application process started.
var StartTime = time.Now().Unix() // unit: second

// Version stores the display version reported to clients (e.g., build time and commit).
var Version = "3.1.0"

// BuildCommit captures the full Git commit hash used for the build when available.
var BuildCommit = ""

// BuildTime records the build timestamp in RFC3339 format when available.
var BuildTime = ""

func init() {
	Version, BuildCommit, BuildTime = computeVersionMetadata(Version, BuildCommit, BuildTime)
}

// computeVersionMetadata aggregates build metadata from the Go toolchain and manual overrides.
// baseVersion, baseCommit, and baseTime allow callers (or ldflags) to pre-populate values that
// should take precedence over the automatically detected metadata.
func computeVersionMetadata(baseVersion, baseCommit, baseTime string) (string, string, string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return finalizeVersionMetadata(baseVersion, baseCommit, baseTime, nil)
	}

	return finalizeVersionMetadata(baseVersion, baseCommit, baseTime, info)
}

// finalizeVersionMetadata produces the user-facing version string as well as the stored commit
// hash and build timestamp derived from the provided build information.
func finalizeVersionMetadata(baseVersion, baseCommit, baseTime string, info *debug.BuildInfo) (string, string, string) {
	version := baseVersion
	commit := baseCommit
	buildTime := baseTime
	modified := false

	if info != nil {
		if version == "" || version == "0.0.0" || version == "(devel)" {
			version = info.Main.Version
			if version == "" {
				version = "(devel)"
			}
		}

		if info.Main.Sum != "" && version != "" && version != "(devel)" {
			version = fmt.Sprintf("%s(%s)", version, info.Main.Sum)
		}

		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if commit == "" {
					commit = setting.Value
				}
			case "vcs.time":
				if buildTime == "" {
					buildTime = setting.Value
				}
			case "vcs.modified":
				modified = setting.Value == "true"
			}
		}
	}

	normalizedTime := normalizeBuildTime(buildTime)
	if normalizedTime != "" {
		buildTime = normalizedTime
	}

	commitDisplay := shortCommit(commit, modified)
	versionDisplay := renderVersionDisplay(version, buildTime, commitDisplay)

	return versionDisplay, commit, buildTime
}

// normalizeBuildTime coerces arbitrary build time strings into RFC3339 UTC format whenever
// possible. It returns an empty string if no valid time can be parsed.
func normalizeBuildTime(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	layouts := []string{time.RFC3339Nano, time.RFC3339}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, trimmed); err == nil {
			return ts.UTC().Format(time.RFC3339)
		}
	}

	return trimmed
}

// shortCommit normalizes the commit hash for display by trimming whitespace, shortening to seven
// characters, and appending a dirty indicator when the worktree contained local changes.
func shortCommit(commit string, modified bool) string {
	trimmed := strings.TrimSpace(commit)
	if trimmed == "" {
		return ""
	}

	if len(trimmed) > 7 {
		trimmed = trimmed[:7]
	}

	if modified {
		trimmed += "-dirty"
	}

	return trimmed
}

// renderVersionDisplay composes the final version string exposed to clients.
func renderVersionDisplay(version, buildTime, commit string) string {
	sanitizedVersion := sanitizeVersion(version)

	parts := make([]string, 0, 3)
	if sanitizedVersion != "" {
		parts = append(parts, sanitizedVersion)
	}
	if buildTime != "" {
		parts = append(parts, buildTime)
	}
	if commit != "" {
		parts = append(parts, fmt.Sprintf("(%s)", commit))
	}

	if len(parts) == 0 {
		return "(devel)"
	}

	return strings.Join(parts, " ")
}

// sanitizeVersion removes placeholder values that should not be shown to users.
func sanitizeVersion(version string) string {
	switch version {
	case "", "0.0.0", "(devel)":
		return ""
	default:
		return version
	}
}
