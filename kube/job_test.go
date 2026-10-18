// pkg/note/kube_job_test.go
package kube

import (
	"testing"
)

func TestFunctionJobSucceeded(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want bool
	}{
		{"nil", nil, false},
		{"no status", map[string]interface{}{}, false},
		{"status without succeeded", map[string]interface{}{
			"status": map[string]interface{}{},
		}, false},
		{"succeeded = 0", map[string]interface{}{
			"status": map[string]interface{}{
				"succeeded": int64(0),
			},
		}, false},
		{"succeeded = 1", map[string]interface{}{
			"status": map[string]interface{}{
				"succeeded": int64(1),
			},
		}, true},
		{"succeeded = 2", map[string]interface{}{
			"status": map[string]interface{}{
				"succeeded": int64(2),
			},
		}, true},
		{"succeeded as float64", map[string]interface{}{
			"status": map[string]interface{}{
				"succeeded": float64(1),
			},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JobSucceeded(tt.obj); got != tt.want {
				t.Errorf("JobSucceeded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionJobFailed(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want bool
	}{
		{"nil", nil, false},
		{"no status", map[string]interface{}{}, false},
		{"status without failed", map[string]interface{}{
			"status": map[string]interface{}{},
		}, false},
		{"failed = 0", map[string]interface{}{
			"status": map[string]interface{}{
				"failed": int64(0),
			},
		}, false},
		{"failed = 1", map[string]interface{}{
			"status": map[string]interface{}{
				"failed": int64(1),
			},
		}, true},
		{"failed = 2", map[string]interface{}{
			"status": map[string]interface{}{
				"failed": int64(2),
			},
		}, true},
		{"failed as float64", map[string]interface{}{
			"status": map[string]interface{}{
				"failed": float64(1),
			},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JobFailed(tt.obj); got != tt.want {
				t.Errorf("JobFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionJobActive(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want bool
	}{
		{"nil", nil, false},
		{"no status", map[string]interface{}{}, false},
		{"status without active", map[string]interface{}{
			"status": map[string]interface{}{},
		}, false},
		{"active = 0", map[string]interface{}{
			"status": map[string]interface{}{
				"active": int64(0),
			},
		}, false},
		{"active = 1", map[string]interface{}{
			"status": map[string]interface{}{
				"active": int64(1),
			},
		}, true},
		{"active = 2", map[string]interface{}{
			"status": map[string]interface{}{
				"active": int64(2),
			},
		}, true},
		{"active as float64", map[string]interface{}{
			"status": map[string]interface{}{
				"active": float64(1),
			},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JobActive(tt.obj); got != tt.want {
				t.Errorf("JobActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionCronJobActiveCount(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want int
	}{
		{"nil", nil, 0},
		{"no status", map[string]interface{}{}, 0},
		{"status without active", map[string]interface{}{
			"status": map[string]interface{}{},
		}, 0},
		{"two active", map[string]interface{}{
			"status": map[string]interface{}{
				"active": []interface{}{
					map[string]interface{}{"name": "job-1"},
					map[string]interface{}{"name": "job-2"},
				},
			},
		}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CronJobActiveCount(tt.obj); got != tt.want {
				t.Errorf("CronJobActiveCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionCronJobLastScheduleTime(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"nil", nil, ""},
		{"no status", map[string]interface{}{}, ""},
		{"no lastScheduleTime", map[string]interface{}{
			"status": map[string]interface{}{},
		}, ""},
		{"has lastScheduleTime", map[string]interface{}{
			"status": map[string]interface{}{
				"lastScheduleTime": "2026-05-19T10:00:00Z",
			},
		}, "2026-05-19T10:00:00Z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CronJobLastScheduleTime(tt.obj); got != tt.want {
				t.Errorf("CronJobLastScheduleTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionCronJobLastSuccessTime(t *testing.T) {
	tests := []struct {
		name string
		obj  interface{}
		want string
	}{
		{"nil", nil, ""},
		{"no status", map[string]interface{}{}, ""},
		{"no lastSuccessfulTime", map[string]interface{}{
			"status": map[string]interface{}{},
		}, ""},
		{"has lastSuccessfulTime", map[string]interface{}{
			"status": map[string]interface{}{
				"lastSuccessfulTime": "2026-05-19T10:00:00Z",
			},
		}, "2026-05-19T10:00:00Z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CronJobLastSuccessTime(tt.obj); got != tt.want {
				t.Errorf("CronJobLastSuccessTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
