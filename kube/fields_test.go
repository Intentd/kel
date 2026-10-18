// pkg/note/kube_fields_test.go
package kube

import (
	"testing"
)

func TestFunctionResourceName(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"valid", map[string]interface{}{"metadata": map[string]interface{}{"name": "my-app"}}, "my-app"},
		{"missing name", map[string]interface{}{"metadata": map[string]interface{}{}}, ""},
		{"missing metadata", map[string]interface{}{}, ""},
		{"not a map", "string", ""},
		{"nil", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResourceName(tt.obj); got != tt.want {
				t.Errorf("ResourceName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFunctionResourceNamespace(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"valid", map[string]interface{}{"metadata": map[string]interface{}{"namespace": "default"}}, "default"},
		{"missing namespace", map[string]interface{}{"metadata": map[string]interface{}{}}, ""},
		{"missing metadata", map[string]interface{}{}, ""},
		{"not a map", "string", ""},
		{"nil", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResourceNamespace(tt.obj); got != tt.want {
				t.Errorf("ResourceNamespace() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFunctionResourceUID(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"valid", map[string]interface{}{"metadata": map[string]interface{}{"uid": "abc-123"}}, "abc-123"},
		{"missing uid", map[string]interface{}{"metadata": map[string]interface{}{}}, ""},
		{"missing metadata", map[string]interface{}{}, ""},
		{"not a map", "string", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResourceUID(tt.obj); got != tt.want {
				t.Errorf("ResourceUID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFunctionResourceVersion(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"valid", map[string]interface{}{"metadata": map[string]interface{}{"resourceVersion": "12345"}}, "12345"},
		{"missing", map[string]interface{}{"metadata": map[string]interface{}{}}, ""},
		{"missing metadata", map[string]interface{}{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResourceVersion(tt.obj); got != tt.want {
				t.Errorf("ResourceVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFunctionCreationTimestamp(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"valid", map[string]interface{}{"metadata": map[string]interface{}{"creationTimestamp": "2024-01-01T00:00:00Z"}}, "2024-01-01T00:00:00Z"},
		{"missing", map[string]interface{}{"metadata": map[string]interface{}{}}, ""},
		{"missing metadata", map[string]interface{}{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreationTimestamp(tt.obj); got != tt.want {
				t.Errorf("CreationTimestamp() = %q, want %q", got, tt.want)
			}
		})
	}
}
