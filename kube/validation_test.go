// pkg/note/kube_validation_test.go
package kube

import "testing"

func TestFunctionIsValidLabelValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"empty is valid", "", true},
		{"simple alphanumeric", "payments", true},
		{"with dash, underscore, dot", "team-payments_v1.0", true},
		{"max length 63", stringOfLen(63), true},
		{"over max length 64", stringOfLen(64), false},
		{"contains space", "team payments", false},
		{"leading dash", "-payments", false},
		{"trailing dash", "payments-", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidLabelValue(tt.value); got != tt.want {
				t.Errorf("IsValidLabelValue(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestFunctionIsValidLabelKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"empty is invalid", "", false},
		{"bare name", "tier", true},
		{"prefixed name", "platform.myorg.io/tier", true},
		{"contains space", "bad key", false},
		{"multiple slashes", "a/b/c", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidLabelKey(tt.key); got != tt.want {
				t.Errorf("IsValidLabelKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestFunctionIsValidAnnotationKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"empty is invalid", "", false},
		{"bare name", "jira-ticket", true},
		{"prefixed name", "platform.myorg.io/jira-ticket", true},
		{"contains space", "bad key", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidAnnotationKey(tt.key); got != tt.want {
				t.Errorf("IsValidAnnotationKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestFunctionIsDNS1123Subdomain(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"empty is invalid", "", false},
		{"bare name", "my-app", true},
		{"dotted subdomain", "my-app.example.com", true},
		{"uppercase invalid", "My-App", false},
		{"underscore invalid", "my_app", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDNS1123Subdomain(tt.value); got != tt.want {
				t.Errorf("IsDNS1123Subdomain(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func stringOfLen(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
