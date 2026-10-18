package core

import "testing"

func TestFunctionIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"simple valid", "team@myorg.io", true},
		{"with plus tag", "team+alerts@myorg.io", true},
		{"with subdomain", "team@mail.myorg.io", true},
		{"empty", "", false},
		{"no at sign", "not-an-email", false},
		{"no domain", "team@", false},
		{"display name form rejected", "Team <team@myorg.io>", false},
		{"trailing space", "team@myorg.io ", true}, // ParseAddress trims
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidEmail(tt.email); got != tt.want {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestFunctionIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"https", "https://example.com/webhook", true},
		{"http", "http://example.com", true},
		{"with query", "https://example.com/path?x=1", true},
		{"empty", "", false},
		{"no scheme", "example.com", false},
		{"ftp unsupported", "ftp://example.com", false},
		{"no host", "https://", false},
		{"plain text", "not a url", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidURL(tt.url); got != tt.want {
				t.Errorf("IsValidURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestFunctionIsValidJSON(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"object", `{"app": "payments-api"}`, true},
		{"array", `[1, 2, 3]`, true},
		{"string", `"hello"`, true},
		{"number", `42`, true},
		{"empty", "", false},
		{"malformed", `{not json`, false},
		{"trailing comma", `{"a": 1,}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidJSON(tt.s); got != tt.want {
				t.Errorf("IsValidJSON(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestFunctionIsValidPort(t *testing.T) {
	tests := []struct {
		name string
		port string
		want bool
	}{
		{"typical", "8080", true},
		{"minimum", "1", true},
		{"maximum", "65535", true},
		{"zero invalid", "0", false},
		{"over max", "65536", false},
		{"negative", "-1", false},
		{"empty", "", false},
		{"non-numeric", "http", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidPort(tt.port); got != tt.want {
				t.Errorf("IsValidPort(%q) = %v, want %v", tt.port, got, tt.want)
			}
		})
	}
}
