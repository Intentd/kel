// pkg/note/kubernetes_test.go
package kube

import "testing"

func TestFunctionGetLabel(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		key  string
		want string
	}{
		{"present", map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"app": "frontend"}}}, "app", "frontend"},
		{"missing key", map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"app": "frontend"}}}, "tier", ""},
		{"no labels", map[string]interface{}{"metadata": map[string]interface{}{}}, "app", ""},
		{"nil", nil, "app", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLabel(tt.obj, tt.key); got != tt.want {
				t.Errorf("GetLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFunctionGetLabelInt(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		key  string
		want int64
	}{
		{"numeric string", map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"replica-count": "3"}}}, "replica-count", 3},
		{"non-numeric", map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"replica-count": "abc"}}}, "replica-count", 0},
		{"missing key", map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{}}}, "replica-count", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLabelInt(tt.obj, tt.key); got != tt.want {
				t.Errorf("GetLabelInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFunctionHasLabel(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		key  string
		want bool
	}{
		{"present non-empty", map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"app": "frontend"}}}, "app", true},
		{"present empty", map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"app": ""}}}, "app", false},
		{"missing", map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{}}}, "app", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasLabel(tt.obj, tt.key); got != tt.want {
				t.Errorf("HasLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionGetAnnotation(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		key  string
		want string
	}{
		{"present", map[string]interface{}{"metadata": map[string]interface{}{"annotations": map[string]interface{}{"autoscale/enabled": "true"}}}, "autoscale/enabled", "true"},
		{"missing key", map[string]interface{}{"metadata": map[string]interface{}{"annotations": map[string]interface{}{}}}, "autoscale/enabled", ""},
		{"no annotations", map[string]interface{}{"metadata": map[string]interface{}{}}, "autoscale/enabled", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAnnotation(tt.obj, tt.key); got != tt.want {
				t.Errorf("GetAnnotation() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFunctionGetAnnotationInt(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		key  string
		want int64
	}{
		{"numeric string", map[string]interface{}{"metadata": map[string]interface{}{"annotations": map[string]interface{}{"autoscale/min-replicas": "2"}}}, "autoscale/min-replicas", 2},
		{"non-numeric", map[string]interface{}{"metadata": map[string]interface{}{"annotations": map[string]interface{}{"autoscale/min-replicas": "many"}}}, "autoscale/min-replicas", 0},
		{"missing key", map[string]interface{}{"metadata": map[string]interface{}{"annotations": map[string]interface{}{}}}, "autoscale/min-replicas", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAnnotationInt(tt.obj, tt.key); got != tt.want {
				t.Errorf("GetAnnotationInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFunctionHasAnnotation(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		key  string
		want bool
	}{
		{"present non-empty", map[string]interface{}{"metadata": map[string]interface{}{"annotations": map[string]interface{}{"autoscale/enabled": "true"}}}, "autoscale/enabled", true},
		{"present empty", map[string]interface{}{"metadata": map[string]interface{}{"annotations": map[string]interface{}{"autoscale/enabled": ""}}}, "autoscale/enabled", false},
		{"missing", map[string]interface{}{"metadata": map[string]interface{}{"annotations": map[string]interface{}{}}}, "autoscale/enabled", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasAnnotation(tt.obj, tt.key); got != tt.want {
				t.Errorf("HasAnnotation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionLabelMatches(t *testing.T) {
	obj := map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"app": "frontend", "env": "prod"}}}
	tests := []struct {
		name string
		obj  interface{}
		kvs  []string
		want bool
	}{
		{"full match", obj, []string{"app", "frontend", "env", "prod"}, true},
		{"partial subset matches", obj, []string{"app", "frontend"}, true},
		{"value mismatch", obj, []string{"app", "backend"}, false},
		{"key absent", obj, []string{"tier", "web"}, false},
		{"odd kvs", obj, []string{"app"}, false},
		{"no labels", map[string]interface{}{"metadata": map[string]interface{}{}}, []string{"app", "frontend"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LabelMatches(tt.obj, tt.kvs...); got != tt.want {
				t.Errorf("LabelMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionGetStatus(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		key  string
		want string
	}{
		{"string field", map[string]interface{}{"status": map[string]interface{}{"phase": "Running"}}, "phase", "Running"},
		{"numeric field", map[string]interface{}{"status": map[string]interface{}{"readyReplicas": float64(3)}}, "readyReplicas", "3"},
		{"structured field", map[string]interface{}{"status": map[string]interface{}{"loadBalancer": map[string]interface{}{"ingress": []interface{}{}}}}, "loadBalancer", ""},
		{"missing", map[string]interface{}{"status": map[string]interface{}{}}, "phase", ""},
		{"no status", map[string]interface{}{}, "phase", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStatus(tt.obj, tt.key); got != tt.want {
				t.Errorf("GetStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFunctionHasStatus(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		key  string
		want bool
	}{
		{"non-empty string", map[string]interface{}{"status": map[string]interface{}{"phase": "Running"}}, "phase", true},
		{"empty string", map[string]interface{}{"status": map[string]interface{}{"phase": ""}}, "phase", false},
		{"non-empty map", map[string]interface{}{"status": map[string]interface{}{"loadBalancer": map[string]interface{}{"ingress": []interface{}{"x"}}}}, "loadBalancer", true},
		{"empty map", map[string]interface{}{"status": map[string]interface{}{"loadBalancer": map[string]interface{}{}}}, "loadBalancer", false},
		{"numeric present", map[string]interface{}{"status": map[string]interface{}{"readyReplicas": float64(0)}}, "readyReplicas", true},
		{"missing", map[string]interface{}{"status": map[string]interface{}{}}, "phase", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasStatus(tt.obj, tt.key); got != tt.want {
				t.Errorf("HasStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
