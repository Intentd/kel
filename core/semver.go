package core

import "text/template"

// semverFuncs — Semver
//
// Domain: semver
//
// Work with semantic versions ("1.2.3", "v1.2.3-rc.1"): read their parts,
// compare them, bump them, and check them against constraints such as ">=
// 1.20". A version that does not parse is an error.
func semverFuncs() template.FuncMap {
	return template.FuncMap{
		"semverMajor":      SemverMajor,
		"semverMinor":      SemverMinor,
		"semverPatch":      SemverPatch,
		"semverValid":      SemverValid,
		"semverCompare":    SemverCompare,
		"semverBump":       SemverBump,
		"semverConstraint": SemverConstraint,
	}
}
