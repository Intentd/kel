package bad

import (
	"strings"
	"text/template"
)

// badFuncs has no title.
func badFuncs() template.FuncMap {
	return template.FuncMap{
		"upper":   strings.ToUpper,
		"nodoc":   NoDoc,
		"noKw":    NoKeywords,
		"related": Related,
	}
}

func NoDoc() string { return "" }

// NoKeywords has an example but no keywords.
//
// Example:
//
//	{{ noKw }}
func NoKeywords() string { return "" }

// Related points at a missing function.
//
// Keywords: x
//
// Example:
//
//	{{ related }}
//
// See also: missing
func Related() string { return "" }
