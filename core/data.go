package core

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

// dataFuncs — Data
//
// Domain: data
//
// Encode, hash and shape values: base64, a short SHA-256 for content-addressed
// names, slugs and truncation that keep a value valid as a Kubernetes name,
// and lookup tables written inline in an expression.
func dataFuncs() template.FuncMap {
	return template.FuncMap{
		"toBase64":   ToBase64,
		"fromBase64": FromBase64,
		"sha256sum":  Sha256Sum,
		"truncate":   Truncate,
		"slugify":    Slugify,
		"repoSlug":   RepoSlug,
		"lookup":     Lookup,
	}
}

// ToBase64 base64-encodes s, as a Secret's data values are stored.
//
// Keywords: data, base64, encode, secret, string
//
// Example:
//
//	{{ toBase64 "s3cr3t" }}  → "czNjcjN0"
func ToBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// FromBase64 decodes a base64 string, such as a Secret's data value. An error
// when s is not base64.
//
// Keywords: data, base64, decode, secret, string, read
//
// Example:
//
//	{{ fromBase64 .secret.data.password }}  → "s3cr3t"
func FromBase64(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("not base64: %w", err)
	}
	return string(b), nil
}

// Sha256Sum returns the first 8 hex characters of the SHA-256 hash of s: a
// short, stable fingerprint for content-addressed names and change
// detection.
//
// Keywords: data, hash, sha256, checksum, fingerprint, content, addressed, string
//
// Example:
//
//	{{ sha256sum "log.level=debug" }}  → "728d4917"
func Sha256Sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:4])
}

// Truncate cuts s to at most maxLen characters with no suffix, so the result
// stays a valid Kubernetes name or label value. For text people read, use
// ellipsis.
//
// Keywords: data, truncate, name, kubernetes, length, limit, string, dns
//
// Example:
//
//	{{ truncate "payments-api-production" 12 }}  → "payments-api"
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

var (
	slugNonAlnum  = regexp.MustCompile(`[^a-z0-9-]`)
	slugMultiDash = regexp.MustCompile(`-+`)
)

// Slugify converts a string to a Kubernetes-safe name segment: lowercase,
// non-alphanumeric characters replaced with dashes, consecutive dashes
// collapsed, leading and trailing dashes trimmed.
//
// Keywords: data, slug, slugify, name, safe, kubernetes, string, lowercase
//
// Example:
//
//	{{ slugify "My Team / Platform" }}  → "my-team-platform"
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = slugNonAlnum.ReplaceAllString(s, "-")
	s = slugMultiDash.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// RepoSlug extracts a Kubernetes-safe name from a repository reference — the
// last path segment (repo name), with a trailing `.git` stripped, then
// slugified. Works with git URLs, org/repo shorthand, or a bare repo name —
// the platform team's repository enum values don't need to be pre-cleaned.
//
// Keywords: data, repo, repository, slug, git, name, serve
//
// Example:
//
//	{{ repoSlug "myorg/payments-api" }}                     → "payments-api"
//	{{ repoSlug "git@github.com:myorg/payments-api.git" }}  → "payments-api"
func RepoSlug(s string) string {
	s = strings.TrimSuffix(strings.TrimRight(s, "/"), ".git")
	if i := strings.LastIndexAny(s, "/:"); i >= 0 {
		s = s[i+1:]
	}
	return Slugify(s)
}

// Lookup returns the value paired with a key in a flat list of alternating
// key/value pairs — the first match wins. Returns `""` if the key matches
// none, same as a missing field elsewhere. A trailing unpaired argument is
// ignored rather than erroring.
//
// Keywords: data, lookup, map, table, switch, derive
//
// Example:
//
//	{{ lookup "orders" "payments" "team-payments" "orders" "team-orders" }}  → "team-orders"
func Lookup(key string, pairs ...string) string {
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i] == key {
			return pairs[i+1]
		}
	}
	return ""
}
