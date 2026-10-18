// pkg/note/kube_replica_test.go
package kube

import (
	"testing"
)

func TestFunctionDesiredReplicas(t *testing.T) {
	tests := []struct {
		name   string
		obj    interface{}
		wanted int
	}{
		{
			name: "spec.replicas = 3",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"replicas": 3,
				},
			},
			wanted: 3,
		},
		{
			name: "spec.replicas missing (default 1)",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{},
			},
			wanted: 1,
		},
		{
			name: "spec.replicas = 0 (scale-to-zero)",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"replicas": 0,
				},
			},
			wanted: 0,
		},
		{
			name:   "nil obj",
			obj:    nil,
			wanted: 1, // fallback
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DesiredReplicas(tt.obj)
			if got != tt.wanted {
				t.Errorf("DesiredReplicas() = %v, want %v", got, tt.wanted)
			}
		})
	}
}

func TestFunctionReadyReplicas(t *testing.T) {
	tests := []struct {
		name   string
		obj    interface{}
		wanted int
	}{
		{
			name: "status.readyReplicas = 2",
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"readyReplicas": 2,
				},
			},
			wanted: 2,
		},
		{
			name: "status missing readyReplicas",
			obj: map[string]interface{}{
				"status": map[string]interface{}{},
			},
			wanted: 0,
		},
		{
			name:   "status missing",
			obj:    map[string]interface{}{},
			wanted: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReadyReplicas(tt.obj)
			if got != tt.wanted {
				t.Errorf("ReadyReplicas() = %v, want %v", got, tt.wanted)
			}
		})
	}
}

func TestFunctionAvailableReplicas(t *testing.T) {
	tests := []struct {
		name   string
		obj    interface{}
		wanted int
	}{
		{
			name: "status.availableReplicas = 2",
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"availableReplicas": 2,
				},
			},
			wanted: 2,
		},
		{
			name: "status missing availableReplicas",
			obj: map[string]interface{}{
				"status": map[string]interface{}{},
			},
			wanted: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AvailableReplicas(tt.obj)
			if got != tt.wanted {
				t.Errorf("AvailableReplicas() = %v, want %v", got, tt.wanted)
			}
		})
	}
}

func TestFunctionUpdatedReplicas(t *testing.T) {
	tests := []struct {
		name   string
		obj    interface{}
		wanted int
	}{
		{
			name: "status.updatedReplicas = 3",
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"updatedReplicas": 3,
				},
			},
			wanted: 3,
		},
		{
			name: "status missing updatedReplicas",
			obj: map[string]interface{}{
				"status": map[string]interface{}{},
			},
			wanted: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UpdatedReplicas(tt.obj)
			if got != tt.wanted {
				t.Errorf("UpdatedReplicas() = %v, want %v", got, tt.wanted)
			}
		})
	}
}

func TestFunctionAllReplicasReady(t *testing.T) {
	tests := []struct {
		name   string
		obj    interface{}
		wanted bool
	}{
		{
			name: "ready == desired -> true",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"replicas": 3,
				},
				"status": map[string]interface{}{
					"readyReplicas": 3,
				},
			},
			wanted: true,
		},
		{
			name: "ready < desired -> false",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"replicas": 3,
				},
				"status": map[string]interface{}{
					"readyReplicas": 2,
				},
			},
			wanted: false,
		},
		{
			name: "zero replicas and zero ready -> true (scale-to-zero)",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"replicas": 0,
				},
				"status": map[string]interface{}{
					"readyReplicas": 0,
				},
			},
			wanted: true,
		},
		{
			name: "desired missing (default 1) but ready=0 -> false",
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"readyReplicas": 0,
				},
			},
			wanted: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AllReplicasReady(tt.obj)
			if got != tt.wanted {
				t.Errorf("AllReplicasReady() = %v, want %v", got, tt.wanted)
			}
		})
	}
}

func TestFunctionRolloutComplete(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want bool
	}{
		{
			name: "rollout complete — updated == desired",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"replicas": int64(3),
				},
				"status": map[string]interface{}{
					"updatedReplicas": int64(3),
				},
			},
			want: true,
		},
		{
			name: "rollout incomplete — updated < desired",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"replicas": int64(3),
				},
				"status": map[string]interface{}{
					"updatedReplicas": int64(2),
				},
			},
			want: false,
		},
		{
			name: "missing spec.replicas (defaults to 1)",
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"updatedReplicas": int64(1),
				},
			},
			want: true,
		},
		{
			name: "missing status.updatedReplicas (defaults to 0)",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"replicas": int64(3),
				},
			},
			want: false,
		},
		{
			name: "scale to zero — desired=0, updated=0",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"replicas": int64(0),
				},
				"status": map[string]interface{}{
					"updatedReplicas": int64(0),
				},
			},
			want: true,
		},
		{
			name: "nil object",
			obj:  nil,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RolloutComplete(tt.obj); got != tt.want {
				t.Errorf("RolloutComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}
