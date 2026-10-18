// pkg/note/kube_pvc_test.go
package kube

import (
	"testing"
)

func makePVC(phase, storageClass, capacity string, accessModes []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"spec": map[string]interface{}{
			"storageClassName": storageClass,
			"accessModes":      accessModes,
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{
					"storage": capacity,
				},
			},
		},
		"status": map[string]interface{}{
			"phase": phase,
		},
	}
}

func TestFunctionPVCBound(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want bool
	}{
		{"nil", nil, false},
		{"no status", map[string]interface{}{}, false},
		{"phase Pending", makePVC("Pending", "standard", "10Gi", nil), false},
		{"phase Bound", makePVC("Bound", "standard", "10Gi", nil), true},
		{"phase Lost", makePVC("Lost", "standard", "10Gi", nil), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PVCBound(tt.obj); got != tt.want {
				t.Errorf("PVCBound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionPVCPhase(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"nil", nil, ""},
		{"no status", map[string]interface{}{}, ""},
		{"Bound", makePVC("Bound", "standard", "10Gi", nil), "Bound"},
		{"Pending", makePVC("Pending", "standard", "10Gi", nil), "Pending"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PVCPhase(tt.obj); got != tt.want {
				t.Errorf("PVCPhase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionPVCCapacity(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"nil", nil, ""},
		{"no spec", map[string]interface{}{}, ""},
		{"10Gi", makePVC("Bound", "standard", "10Gi", nil), "10Gi"},
		{"100Gi", makePVC("Bound", "standard", "100Gi", nil), "100Gi"},
		{"no resources", map[string]interface{}{"spec": map[string]interface{}{}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PVCCapacity(tt.obj); got != tt.want {
				t.Errorf("PVCCapacity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionPVCStorageClass(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"nil", nil, ""},
		{"no spec", map[string]interface{}{}, ""},
		{"standard", makePVC("Bound", "standard", "10Gi", nil), "standard"},
		{"gp2", makePVC("Bound", "gp2", "10Gi", nil), "gp2"},
		{"empty storageClass", makePVC("Bound", "", "10Gi", nil), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PVCStorageClass(tt.obj); got != tt.want {
				t.Errorf("PVCStorageClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

func makePVCWithStorageClass(provisioner, reclaimPolicy string) map[string]interface{} {
	return map[string]interface{}{
		"_storageClass": map[string]interface{}{
			"provisioner": provisioner,
			"spec": map[string]interface{}{
				"reclaimPolicy": reclaimPolicy,
			},
		},
	}
}
