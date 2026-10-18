package ecosystem

import (
	"regexp"
	"text/template"
)

// validationFuncs — Validation
//
// Domain: validation
//
// Check references to artifacts the tools around Kubernetes consume: a git
// repository (Argo CD, Flux) and a container image reference (registries).
// Catching a malformed one early beats finding out when the tool fails to
// reconcile it.
func validationFuncs() template.FuncMap {
	return template.FuncMap{
		"isValidGitRepository": IsValidGitRepository,
		"isValidImageRef":      IsValidImageRef,
	}
}

// gitRepoPatterns matches the URL and SCP-like ("user@host:path") shapes a
// git remote commonly takes.
var gitRepoPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^(https?|git|ssh)://[^\s]+$`),
	regexp.MustCompile(`^[\w.-]+@[\w.-]+:[\w./~-]+(\.git)?$`),
}

// IsValidGitRepository reports whether s looks like a git repository URL —
// https://, git://, ssh://, or the SCP-like "git@host:org/repo.git" shorthand.
// This is a shape check, not a reachability check — it doesn't dial the host.
//
// Keywords: git, repository, url, validate, valid, boolean, format, repo
//
// Example:
//
//	{{ isValidGitRepository "https://github.com/myorg/payments"    }}  → true
//	{{ isValidGitRepository "git@github.com:myorg/payments.git"    }}  → true
//	{{ isValidGitRepository "not a repo"                            }}  → false
func IsValidGitRepository(s string) bool {
	if s == "" {
		return false
	}
	for _, re := range gitRepoPatterns {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

// imageRefPattern is a pragmatic container image reference check —
// [registry[:port]/]repo/path[:tag|@digest] — not a full implementation of
// the OCI distribution spec, but enough to catch the common mistakes: a
// stray space, an uppercase letter, a missing tag separator.
var imageRefPattern = regexp.MustCompile(
	`^(?:[a-z0-9]+(?:[.-][a-z0-9]+)*(?::[0-9]+)?/)?` +
		`[a-z0-9]+(?:[._-]+[a-z0-9]+)*(?:/[a-z0-9]+(?:[._-]+[a-z0-9]+)*)*` +
		`(?::[\w][\w.-]{0,127}|@[A-Za-z][A-Za-z0-9]*(?:[-_+.][A-Za-z][A-Za-z0-9]*)*:[0-9A-Fa-f]{32,})?$`,
)

// IsValidImageRef reports whether s looks like a valid container image
// reference: optional registry, lowercase repo path, optional :tag or @digest.
//
// Keywords: image, container, docker, registry, validate, valid, boolean, format, tag
//
// Example:
//
//	{{ isValidImageRef "myorg/app:v1.2.3"                    }}  → true
//	{{ isValidImageRef "registry.myorg.io:5000/team/app:latest" }}  → true
//	{{ isValidImageRef "MyOrg/App"                            }}  → false
func IsValidImageRef(s string) bool {
	if s == "" {
		return false
	}
	return imageRefPattern.MatchString(s)
}
