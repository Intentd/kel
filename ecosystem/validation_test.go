package ecosystem

import "testing"

func TestFunctionIsValidGitRepository(t *testing.T) {
	tests := []struct {
		name string
		repo string
		want bool
	}{
		{"https URL", "https://github.com/myorg/payments", true},
		{"https URL with .git suffix", "https://github.com/myorg/payments.git", true},
		{"http URL", "http://internal-git.myorg.io/team/app", true},
		{"git protocol", "git://github.com/myorg/payments.git", true},
		{"ssh protocol", "ssh://git@github.com/myorg/payments.git", true},
		{"scp-like shorthand", "git@github.com:myorg/payments.git", true},
		{"scp-like without .git", "git@github.com:myorg/payments", true},
		{"empty", "", false},
		{"plain text", "not a repo", false},
		{"ftp scheme unsupported", "ftp://example.com/repo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidGitRepository(tt.repo); got != tt.want {
				t.Errorf("IsValidGitRepository(%q) = %v, want %v", tt.repo, got, tt.want)
			}
		})
	}
}

func TestFunctionIsValidImageRef(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want bool
	}{
		{"bare name", "nginx", true},
		{"org/repo", "myorg/app", true},
		{"with tag", "myorg/app:v1.2.3", true},
		{"with latest tag", "myorg/app:latest", true},
		{"registry with port", "registry.myorg.io:5000/team/app:latest", true},
		{"with digest", "myorg/app@sha256:" + fortyNineHexChars, true},
		{"empty", "", false},
		{"uppercase invalid", "MyOrg/App", false},
		{"contains space", "my org/app", false},
		{"double colon", "myorg/app::latest", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidImageRef(tt.ref); got != tt.want {
				t.Errorf("IsValidImageRef(%q) = %v, want %v", tt.ref, got, tt.want)
			}
		})
	}
}

// fortyNineHexChars is a syntactically valid sha256 hex digest (64 chars) —
// named for the regex's 32+ char minimum, but a real digest is always 64.
const fortyNineHexChars = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b85"
