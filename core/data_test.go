// pkg/note/data_test.go
package core

import (
	"encoding/base64"
	"testing"
)

func TestFunctionToBase64(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"empty", "", ""},
		{"simple", "hello", base64.StdEncoding.EncodeToString([]byte("hello"))},
		{"with special chars", "a/b+c", base64.StdEncoding.EncodeToString([]byte("a/b+c"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToBase64(tt.s); got != tt.want {
				t.Errorf("toBase64(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestFunctionFromBase64(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"empty", "", ""},
		{"valid", base64.StdEncoding.EncodeToString([]byte("hello")), "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := FromBase64(tt.s); err != nil || got != tt.want {
				t.Errorf("fromBase64(%q) = %q, %v, want %q", tt.s, got, err, tt.want)
			}
		})
	}
	if _, err := FromBase64("!!!"); err == nil {
		t.Error("fromBase64(!!!): want an error for input that is not base64")
	}
}

func TestFunctionToJSON(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want string
	}{
		{"map", map[string]int{"a": 1}, `{"a":1}`},
		{"slice", []string{"x", "y"}, `["x","y"]`},
		{"string", "hello", `"hello"`},
		{"nil", nil, "null"},
		{"struct", struct{ A int }{42}, `{"A":42}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := ToJSON(tt.v); got != tt.want {
				t.Errorf("toJSON(%v) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestFunctionSHA256Sum(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		wantLen int
	}{
		{"empty", "", 8},
		{"hello", "hello", 8},
		{"long", "this is a longer string that is hashed", 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sha256Sum(tt.s)
			if len(got) != tt.wantLen {
				t.Errorf("sha256sum(%q) length = %d, want %d", tt.s, len(got), tt.wantLen)
			}
			// deterministic
			if got != Sha256Sum(tt.s) {
				t.Errorf("sha256sum not deterministic: %q != %q", got, Sha256Sum(tt.s))
			}
		})
	}
}

func TestFunctionTruncate(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{"shorter", "hello", 10, "hello"},
		{"exact", "world", 5, "world"},
		{"longer", "hello world", 5, "hello"},
		{"zero length", "test", 0, ""},
		{"empty", "", 10, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.s, tt.maxLen); got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFunctionSlugify(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"empty", "", ""},
		{"simple", "My App", "my-app"},
		{"with slash", "My App / Service", "my-app-service"},
		{"multiple spaces", "a   b", "a-b"},
		{"leading dash", "-hello-", "hello"},
		{"special chars", "foo@bar$baz", "foo-bar-baz"},
		{"underscores", "hello_world", "hello-world"},
		{"numbers", "app123", "app123"},
		{"already slug", "my-app", "my-app"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slugify(tt.s); got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}
