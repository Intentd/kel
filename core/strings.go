package core

import (
	"strings"
	"text/template"
)

// stringFuncs — Strings
//
// Domain: strings
//
// The string operations expressions need most: case, trimming, prefixes and
// suffixes, splitting and joining, concatenation, case conversion between
// camelCase and kebab-case, and shortening text for display.
func stringFuncs() template.FuncMap {
	return template.FuncMap{
		"toLower":      ToLower,
		"toUpper":      ToUpper,
		"trimSpace":    TrimSpace,
		"trim":         Trim,
		"trimPrefix":   TrimPrefix,
		"trimSuffix":   TrimSuffix,
		"hasPrefix":    HasPrefix,
		"hasSuffix":    HasSuffix,
		"contains":     Contains,
		"replace":      Replace,
		"split":        Split,
		"join":         Join,
		"repeat":       Repeat,
		"camelToKebab": CamelToKebab,
		"ellipsis":     Ellipsis,
		"concat":       Concat,
	}
}

// Split splits a string by a separator into a slice. Combine with `index` to
// extract a specific element.
//
// Keywords: string, split, separator, slice, list, parse
//
// Example:
//
//	{{ index (split "latest,stable,v2" ",") 0 }}  → "latest"
func Split(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}

// CamelToKebab converts CamelCase or PascalCase to kebab-case. Useful when
// deriving Kubernetes resource names from Go type names.
//
// Keywords: string, case, camel, kebab, transform, naming, kubernetes
//
// Example:
//
//	{{ camelToKebab "WebsiteOperator" }}  → "website-operator"
func CamelToKebab(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Convert to lowercase
		lower := c
		if c >= 'A' && c <= 'Z' {
			lower = c - 'A' + 'a'
		}
		// Add hyphen before uppercase if:
		// - it's not the first character,
		// - and the previous character is lowercase (word boundary),
		// - or the previous character is uppercase but the next is lowercase (acronym boundary).
		if i > 0 && c >= 'A' && c <= 'Z' {
			prev := s[i-1]
			next := byte(0)
			if i+1 < len(s) {
				next = s[i+1]
			}
			// Insert hyphen when:
			//   prev is lowercase (new word)
			//   OR prev is uppercase and next is lowercase (end of acronym)
			if (prev >= 'a' && prev <= 'z') || (prev >= 'A' && prev <= 'Z' && next >= 'a' && next <= 'z') {
				b.WriteByte('-')
			}
		}
		b.WriteByte(lower)
	}
	return b.String()
}

// Ellipsis shortens s to at most n characters, ending in "..." when it was
// cut. For text people read (messages, descriptions); not for names or label
// values, which may not end in a dot: use truncate there.
//
// Keywords: string, ellipsis, shorten, limit, length, display, message
//
// Example:
//
//	{{ ellipsis .status.message 80 }}
func Ellipsis(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

// Concat joins any number of strings together with no separator. Useful for
// building domain names, resource name prefixes, or any value assembled from
// multiple parts.
//
// Keywords: string, concat, join, combine, build, prefix, suffix, append
//
// Example:
//
//	{{ concat "*." .spec.domain }}  → "*.api.example.com"
func Concat(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p)
	}
	return b.String()
}

// ToLower returns s in lower case.
//
// Keywords: string, case, lower
//
// Example:
//
//	{{ toLower "Payments-API" }}  → "payments-api"
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToUpper returns s in upper case.
//
// Keywords: string, case, upper
//
// Example:
//
//	{{ toUpper .spec.env }}  → "PROD"
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// TrimSpace returns s without leading and trailing white space.
//
// Keywords: string, trim, whitespace
//
// Example:
//
//	{{ trimSpace .spec.name }}
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// Trim returns s without the leading and trailing characters in cutset.
//
// Keywords: string, trim, cutset
//
// Example:
//
//	{{ trim .spec.path "/" }}  → "api"
func Trim(s, cutset string) string {
	return strings.Trim(s, cutset)
}

// TrimPrefix returns s without prefix, if it starts with it.
//
// Keywords: string, trim, prefix
//
// Example:
//
//	{{ trimPrefix .spec.image "docker.io/" }}
func TrimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

// TrimSuffix returns s without suffix, if it ends with it.
//
// Keywords: string, trim, suffix
//
// Example:
//
//	{{ trimSuffix .spec.host ".local" }}
func TrimSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

// HasPrefix reports whether s starts with prefix.
//
// Keywords: string, prefix, check
//
// Example:
//
//	{{ hasPrefix .spec.image "docker.io/" }}  → true
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix reports whether s ends with suffix.
//
// Keywords: string, suffix, check
//
// Example:
//
//	{{ hasSuffix .spec.host ".local" }}  → true
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Contains reports whether substr is within s.
//
// Keywords: string, contains, substring, check
//
// Example:
//
//	{{ contains .spec.image ":latest" }}  → true
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Replace returns s with every occurrence of old replaced by new.
//
// Keywords: string, replace, substitute
//
// Example:
//
//	{{ replace "payments-api" "-" "_" }}  → "payments_api"
func Replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// Join joins the elements of a list with sep between them.
//
// Keywords: string, join, list, separator
//
// Example:
//
//	{{ join (podNames .pods) ", " }}  → "web-0, web-1, web-2"
func Join(elems []string, sep string) string {
	return strings.Join(elems, sep)
}

// Repeat returns s repeated count times.
//
// Keywords: string, repeat
//
// Example:
//
//	{{ repeat "-" 3 }}  → "---"
func Repeat(s string, count int) string {
	return strings.Repeat(s, count)
}
