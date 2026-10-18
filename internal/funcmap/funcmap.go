// Package funcmap merges template function maps for KEL's packages: a name
// registered twice is a programming error, and replacing one must be said
// explicitly.
package funcmap

import (
	"fmt"
	"text/template"
)

// Register merges src into dst. A name already in dst is a programming error:
// two functions would answer to one name and the last registered would win
// silently. To replace a function on purpose, use Override.
func Register(dst, src template.FuncMap) {
	for k, v := range src {
		if _, taken := dst[k]; taken {
			panic(fmt.Sprintf("kel: function %q registered twice", k))
		}
		dst[k] = v
	}
}

// Override replaces functions already in dst with src's. Every name in src
// must already be registered: an override of nothing is a mistake too.
func Override(dst, src template.FuncMap) {
	for k, v := range src {
		if _, taken := dst[k]; !taken {
			panic(fmt.Sprintf("kel: override of %q, which is not registered", k))
		}
		dst[k] = v
	}
}
