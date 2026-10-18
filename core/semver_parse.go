package core

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// SemverValid reports whether v is a valid semantic version; a leading "v"
// is accepted.
//
// Keywords: semver, version, valid, validate, boolean, parse, check
//
// Example:
//
//	{{ semverValid .spec.version }}  → true
func SemverValid(v string) bool {
	_, err := semver.NewVersion(v)
	return err == nil
}

// SemverMajor returns the major component of version v; an error when v is
// not a valid version.
//
// Keywords: semver, version, major, component, parse
//
// Example:
//
//	{{ semverMajor "1.2.3" }}  → 1
//
// See also: semverMinor, semverPatch
func SemverMajor(v string) (uint64, error) {
	sv, err := semver.NewVersion(v)
	if err != nil {
		return 0, fmt.Errorf("invalid semver %q: %w", v, err)
	}
	return sv.Major(), nil
}

// SemverMinor returns the minor component of version v; an error when v is
// not a valid version.
//
// Keywords: semver, version, minor, component, parse
//
// Example:
//
//	{{ semverMinor "1.2.3" }}  → 2
//
// See also: semverMajor, semverPatch
func SemverMinor(v string) (uint64, error) {
	sv, err := semver.NewVersion(v)
	if err != nil {
		return 0, fmt.Errorf("invalid semver %q: %w", v, err)
	}
	return sv.Minor(), nil
}

// SemverPatch returns the patch component of version v; an error when v is
// not a valid version.
//
// Keywords: semver, version, patch, component, parse
//
// Example:
//
//	{{ semverPatch "1.2.3" }}  → 3
//
// See also: semverMajor, semverMinor
func SemverPatch(v string) (uint64, error) {
	sv, err := semver.NewVersion(v)
	if err != nil {
		return 0, fmt.Errorf("invalid semver %q: %w", v, err)
	}
	return sv.Patch(), nil
}

// SemverCompare compares versions a and b: -1 when a is older, 0 when they
// are equal, 1 when a is newer. An error when either is not a valid version.
//
// Keywords: semver, version, compare, order, upgrade, greater, less
//
// Example:
//
//	{{ semverCompare .spec.version .status.runningVersion }}  → 1
//
// See also: semverConstraint
func SemverCompare(a, b string) (int, error) {
	sa, err := semver.NewVersion(a)
	if err != nil {
		return 0, fmt.Errorf("invalid semver %q: %w", a, err)
	}
	sb, err := semver.NewVersion(b)
	if err != nil {
		return 0, fmt.Errorf("invalid semver %q: %w", b, err)
	}
	return sa.Compare(sb), nil
}

// SemverConstraint reports whether version v satisfies constraint, written
// as ">=1.2.0,<2.0.0", "^2.0", "~1.2" or "1.x". An error when either does
// not parse.
//
// Keywords: semver, version, constraint, range, compatible, boolean, check
//
// Example:
//
//	{{ semverConstraint .spec.version "^1.0" }}  → true
//
// See also: semverCompare
func SemverConstraint(v, constraint string) (bool, error) {
	sv, err := semver.NewVersion(v)
	if err != nil {
		return false, fmt.Errorf("invalid semver %q: %w", v, err)
	}
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false, fmt.Errorf("invalid semver constraint %q: %w", constraint, err)
	}
	return c.Check(sv), nil
}

// SemverBump increments component ("major", "minor" or "patch") of version
// v, resetting the lower components and dropping prerelease and build
// metadata.
//
// Keywords: semver, version, bump, increment, next, release
//
// Example:
//
//	{{ semverBump "1.2.3" "minor" }}  → 1.3.0
func SemverBump(v, component string) (string, error) {
	sv, err := semver.NewVersion(strings.TrimPrefix(v, "v"))
	if err != nil {
		return "", fmt.Errorf("invalid semver %q: %w", v, err)
	}
	major, minor, patch := sv.Major(), sv.Minor(), sv.Patch()
	switch strings.ToLower(component) {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", fmt.Errorf("unknown semver component %q: must be major, minor, or patch", component)
	}
	return semver.New(major, minor, patch, "", "").String(), nil
}
