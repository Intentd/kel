// pkg/note/kube_node_test.go
package kube

import (
	"testing"
)

func makeNode(conditions map[string]string, allocatable map[string]string, taints []string) map[string]interface{} {
	condList := make([]interface{}, 0, len(conditions))
	for k, v := range conditions {
		condList = append(condList, map[string]interface{}{"type": k, "status": v})
	}
	alloc := map[string]interface{}{}
	for k, v := range allocatable {
		alloc[k] = v
	}
	taintList := make([]interface{}, 0, len(taints))
	for _, k := range taints {
		taintList = append(taintList, map[string]interface{}{"key": k, "effect": "NoSchedule"})
	}
	return map[string]interface{}{
		"spec": map[string]interface{}{
			"taints": taintList,
		},
		"status": map[string]interface{}{
			"conditions":  condList,
			"allocatable": alloc,
		},
	}
}

func TestFunctionNodeReady(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want bool
	}{
		{"nil", nil, false},
		{"no status", map[string]interface{}{}, false},
		{"Ready=True", makeNode(map[string]string{"Ready": "True"}, nil, nil), true},
		{"Ready=False", makeNode(map[string]string{"Ready": "False"}, nil, nil), false},
		{"Ready absent", makeNode(map[string]string{"MemoryPressure": "False"}, nil, nil), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NodeReady(tt.obj); got != tt.want {
				t.Errorf("NodeReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionNodeAllocatableCPU(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"nil", nil, ""},
		{"no status", map[string]interface{}{}, ""},
		{"with cpu", makeNode(nil, map[string]string{"cpu": "3920m"}, nil), "3920m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NodeAllocatableCPU(tt.obj); got != tt.want {
				t.Errorf("NodeAllocatableCPU() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionNodeAllocatableMemory(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"nil", nil, ""},
		{"no status", map[string]interface{}{}, ""},
		{"with memory", makeNode(nil, map[string]string{"memory": "15032020Ki"}, nil), "15032020Ki"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NodeAllocatableMemory(tt.obj); got != tt.want {
				t.Errorf("NodeAllocatableMemory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionNodeCondition(t *testing.T) {
	tests := []struct {
		name     string
		obj      interface{}
		condType string
		want     string
	}{
		{"nil", nil, "Ready", ""},
		{"condition not found", makeNode(map[string]string{"Ready": "True"}, nil, nil), "MemoryPressure", ""},
		{"Ready=True", makeNode(map[string]string{"Ready": "True"}, nil, nil), "Ready", "True"},
		{"MemoryPressure=False", makeNode(map[string]string{"Ready": "True", "MemoryPressure": "False"}, nil, nil), "MemoryPressure", "False"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NodeCondition(tt.obj, tt.condType); got != tt.want {
				t.Errorf("NodeCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func makePodWithNode(nodeName, zone, region, instanceType string) map[string]interface{} {
	return map[string]interface{}{
		"_node": map[string]interface{}{
			"name":         nodeName,
			"zone":         zone,
			"region":       region,
			"instanceType": instanceType,
		},
	}
}
