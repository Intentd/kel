package core

import (
	"strings"
	"text/template"
)

// domainFuncs — Domain
//
// Domain: domain
//
// Take domain names and URLs apart: the host of a URL, and a domain without
// its subdomain.
func domainFuncs() template.FuncMap {
	return template.FuncMap{
		"domainHost": DomainHost,
		"domainBare": DomainBare,
	}
}

// DomainHost strips surrounding whitespace, `https://` or `http://` prefix,
// and trailing `/`. Returns the bare hostname. Returns `""` for nil or
// non-string input. Idempotent — calling it on an already-clean hostname is a
// no-op.
//
// Keywords: domain, string, url, hostname, protocol, normalize, strip
//
// Example:
//
//	{{ domainHost "https://acme.example.com/" }}  → "acme.example.com"
func DomainHost(v interface{}) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimSuffix(s, "/")
	return s
}

// DomainBare is like `domainHost`, and also strips a leading `www.`
// subdomain. Returns `""` for nil or non-string input. Idempotent.
//
// Keywords: domain, string, www, subdomain, hostname, normalize, strip
//
// Example:
//
//	{{ domainBare "https://www.acme.com/" }}  → "acme.com"
func DomainBare(v interface{}) string {
	s := DomainHost(v)
	if s == "" {
		return ""
	}
	return strings.TrimPrefix(s, "www.")
}
