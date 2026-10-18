// pkg/note/kube_hpa_test.go
package kube

import (
	"testing"
)

func makeHPA(current, desired, min, max int64) map[string]interface{} {
	return map[string]interface{}{
		"spec": map[string]interface{}{
			"minReplicas": min,
			"maxReplicas": max,
		},
		"status": map[string]interface{}{
			"currentReplicas": current,
			"desiredReplicas": desired,
		},
	}
}

func makeHPAScaling(current, desired, min, max int64) map[string]interface{} {
	hpa := makeHPA(current, desired, min, max)
	status := hpa["status"].(map[string]interface{})
	status["conditions"] = []interface{}{
		map[string]interface{}{
			"type":   "ScalingActive",
			"status": "True",
		},
	}
	return hpa
}

func TestFunctionHPACurrentReplicas(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want int64
	}{
		{"nil", nil, 0},
		{"no status", map[string]interface{}{}, 0},
		{"three replicas", makeHPA(3, 3, 2, 10), 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HPACurrentReplicas(tt.obj); got != tt.want {
				t.Errorf("HPACurrentReplicas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionHPADesiredReplicas(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want int64
	}{
		{"nil", nil, 0},
		{"no status", map[string]interface{}{}, 0},
		{"five desired", makeHPA(3, 5, 2, 10), 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HPADesiredReplicas(tt.obj); got != tt.want {
				t.Errorf("HPADesiredReplicas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionHPAMinReplicas(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want int64
	}{
		{"nil", nil, 1},
		{"no spec", map[string]interface{}{}, 1},
		{"min 2", makeHPA(3, 3, 2, 10), 2},
		{"min 0 defaults to 1", makeHPA(3, 3, 0, 10), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HPAMinReplicas(tt.obj); got != tt.want {
				t.Errorf("HPAMinReplicas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionHPAMaxReplicas(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want int64
	}{
		{"nil", nil, 0},
		{"no spec", map[string]interface{}{}, 0},
		{"max 10", makeHPA(3, 3, 2, 10), 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HPAMaxReplicas(tt.obj); got != tt.want {
				t.Errorf("HPAMaxReplicas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionHPAScaling(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want bool
	}{
		{"nil", nil, false},
		{"stable current==desired", makeHPAScaling(3, 3, 2, 10), false},
		{"scaling up current<desired", makeHPAScaling(3, 5, 2, 10), true},
		{"scaling down current>desired", makeHPAScaling(5, 3, 2, 10), true},
		{"metric unknown desiredReplicas=0", makeHPA(1, 0, 1, 10), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HPAScaling(tt.obj); got != tt.want {
				t.Errorf("HPAScaling() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionHPAAtMax(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want bool
	}{
		{"nil", nil, false},
		{"below max", makeHPA(3, 3, 2, 10), false},
		{"at max", makeHPA(10, 10, 2, 10), true},
		{"above max (edge)", makeHPA(11, 11, 2, 10), true},
		{"max 0 (not set)", makeHPA(3, 3, 2, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HPAAtMax(tt.obj); got != tt.want {
				t.Errorf("HPAAtMax() = %v, want %v", got, tt.want)
			}
		})
	}
}
