// pkg/note/kube_statefulset_test.go
package kube

import "testing"

func makeStatefulSetWithRevisions(current, update string) map[string]interface{} {
	return map[string]interface{}{
		"status": map[string]interface{}{
			"currentRevision": current,
			"updateRevision":  update,
		},
	}
}

func makeStatefulSetWithPVCs(count int) map[string]interface{} {
	pvcs := make([]interface{}, count)
	for i := range pvcs {
		pvcs[i] = map[string]interface{}{"metadata": map[string]interface{}{"name": "pvc"}}
	}
	return map[string]interface{}{"_pvcs": pvcs}
}

func TestFunctionStatefulSetCurrentRevision(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"nil", nil, ""},
		{"no status", map[string]interface{}{}, ""},
		{"has revision", makeStatefulSetWithRevisions("my-sts-abc123", "my-sts-abc123"), "my-sts-abc123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatefulSetCurrentRevision(tt.obj); got != tt.want {
				t.Errorf("StatefulSetCurrentRevision() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFunctionStatefulSetUpdateRevision(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"nil", nil, ""},
		{"no status", map[string]interface{}{}, ""},
		{"pending update", makeStatefulSetWithRevisions("my-sts-old", "my-sts-new"), "my-sts-new"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatefulSetUpdateRevision(tt.obj); got != tt.want {
				t.Errorf("StatefulSetUpdateRevision() = %q, want %q", got, tt.want)
			}
		})
	}
}
