package core

import (
	"encoding/json"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"text/template"
)

// validationFuncs — Validation
//
// Domain: validation
//
// Check that free text is the kind of value it claims to be: an email address,
// a URL, JSON, a port number. A form field is free text until something checks
// it; catching a malformed value early, with a clear message, beats finding
// out downstream.
func validationFuncs() template.FuncMap {
	return template.FuncMap{
		"isValidEmail": IsValidEmail,
		"isValidURL":   IsValidURL,
		"isValidJSON":  IsValidJSON,
		"isValidPort":  IsValidPort,
	}
}

// IsValidEmail reports whether s is a single, bare email address —
// "user@example.com", not a display-name form like "User <user@example.com>".
//
// Keywords: email, validate, valid, boolean, format, contact
//
// Example:
//
//	{{ isValidEmail "team@myorg.io" }}  → true
//	{{ isValidEmail "not-an-email"  }}  → false
func IsValidEmail(s string) bool {
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return false
	}
	// ParseAddress accepts "Display Name <addr>" too — reject anything
	// that isn't just the bare address the caller passed in.
	return addr.Address == strings.TrimSpace(s)
}

// IsValidURL reports whether s is an absolute http or https URL with a
// non-empty host.
//
// Keywords: url, http, https, validate, valid, boolean, format, webhook
//
// Example:
//
//	{{ isValidURL "https://example.com/webhook" }}  → true
//	{{ isValidURL "ftp://example.com"            }}  → false
//	{{ isValidURL "not a url"                     }}  → false
func IsValidURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// IsValidJSON reports whether s is syntactically valid JSON.
//
// Keywords: json, validate, valid, boolean, format, syntax
//
// Example:
//
//	{{ isValidJSON "{\"app\": \"payments-api\"}" }}  → true
//	{{ isValidJSON "{not json"                    }}  → false
func IsValidJSON(s string) bool {
	return json.Valid([]byte(s))
}

// IsValidPort reports whether s parses as an integer in the valid TCP/UDP port
// range, 1–65535.
//
// Keywords: port, network, validate, valid, boolean, format, number
//
// Example:
//
//	{{ isValidPort "8080"  }}  → true
//	{{ isValidPort "0"     }}  → false
//	{{ isValidPort "70000" }}  → false
func IsValidPort(s string) bool {
	n, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return n >= 1 && n <= 65535
}
