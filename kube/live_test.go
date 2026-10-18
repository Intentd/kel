package kube

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/intentd/kel/internal/testutil"
)

// Objects captured from a live cluster in testdata/live, each a workload and
// its pods in a state that is easy to read wrong. See testdata/live/README.md.

type liveSnapshot struct {
	Kubernetes string
	Workload   interface{}
	Pods       interface{}
}

// liveTypes are the typed client-go forms of the snapshots' workloads.
var liveTypes = map[interface{}]interface{}{
	"Deployment": appsv1.Deployment{}, "StatefulSet": appsv1.StatefulSet{}, "Job": batchv1.Job{},
}

// loadLive reads a snapshot in shape: "map" as a dynamic client returns it,
// or "typed" as client-go objects from a typed list call.
func loadLive(t *testing.T, name, shape string) liveSnapshot {
	t.Helper()
	var s struct {
		Kubernetes string                 `json:"kubernetes"`
		Workload   map[string]interface{} `json:"workload"`
		Pods       map[string]interface{} `json:"pods"`
	}
	testutil.ReadJSON(t, &s, "live", name+".json")
	if shape == "map" {
		return liveSnapshot{s.Kubernetes, s.Workload, s.Pods}
	}
	proto, ok := liveTypes[s.Workload["kind"]]
	if !ok {
		t.Fatalf("%s: no typed form for %v", name, s.Workload["kind"])
	}
	items, _ := s.Pods["items"].([]interface{})
	return liveSnapshot{s.Kubernetes, testutil.Typed(t, s.Workload, proto), testutil.TypedList(t, items, corev1.Pod{})}
}

func TestLiveSnapshots(t *testing.T) {
	cases := []struct {
		name  string
		check func(t *testing.T, s liveSnapshot)
	}{
		{"healthy", func(t *testing.T, s liveSnapshot) {
			assert.True(t, RolloutComplete(s.Workload))
			assert.True(t, ConditionsHealthy(s.Workload))
			assert.Equal(t, 2, ReadyPodCount(s.Pods))
			assert.False(t, HasCrashingPod(s.Pods))
			assert.Empty(t, PodContainerReasons(s.Pods))
		}},
		{"statefulset", func(t *testing.T, s liveSnapshot) {
			assert.True(t, RolloutComplete(s.Workload))
			assert.True(t, AllReplicasReady(s.Workload))
			assert.False(t, HasCrashingPod(s.Pods))
		}},
		{"bad-image", func(t *testing.T, s liveSnapshot) {
			assert.False(t, RolloutComplete(s.Workload))
			assert.False(t, ConditionsHealthy(s.Workload))
			assert.True(t, PodImagePullBackOffDetected(s.Pods))
			assert.True(t, HasCrashingPod(s.Pods))
			assert.False(t, PodCrashLoopDetected(s.Pods), "never started, so not a crash loop")
		}},
		{"crashloop-backoff", func(t *testing.T, s liveSnapshot) {
			assert.True(t, PodCrashLoopDetected(s.Pods))
			assert.True(t, HasCrashingPod(s.Pods))
			assert.Equal(t, []string{"CrashLoopBackOff", "Error"}, PodContainerReasons(s.Pods))
			assert.False(t, PodOOMKilledDetected(s.Pods))
		}},
		{"crashloop-between-restarts", func(t *testing.T, s liveSnapshot) {
			// The state says terminated/Error, not CrashLoopBackOff.
			assert.True(t, PodCrashLoopDetected(s.Pods))
			assert.True(t, HasCrashingPod(s.Pods))
			assert.Equal(t, []string{"Error"}, PodContainerReasons(s.Pods))
		}},
		{"oom-terminated", func(t *testing.T, s liveSnapshot) {
			assert.True(t, PodOOMKilledDetected(s.Pods))
			assert.True(t, PodCrashLoopDetected(s.Pods), "exit 137 after restarting before")
			assert.True(t, HasCrashingPod(s.Pods))
		}},
		{"oom-backoff", func(t *testing.T, s liveSnapshot) {
			// The state says CrashLoopBackOff; OOMKilled is in lastState.
			assert.True(t, PodOOMKilledDetected(s.Pods))
			assert.True(t, PodCrashLoopDetected(s.Pods))
			assert.Equal(t, []string{"CrashLoopBackOff", "OOMKilled"}, PodContainerReasons(s.Pods))
		}},
		{"init-crashloop", func(t *testing.T, s liveSnapshot) {
			// Only initContainerStatuses shows the failure (kubectl: Init:CrashLoopBackOff).
			assert.True(t, HasCrashingPod(s.Pods))
			assert.True(t, PodCrashLoopDetected(s.Pods))
			assert.Equal(t, []string{"CrashLoopBackOff", "Error", "PodInitializing"}, PodContainerReasons(s.Pods))
			assert.False(t, RolloutComplete(s.Workload))
		}},
		{"init-error", func(t *testing.T, s liveSnapshot) {
			// kubectl: Init:Error, between the init container's restarts.
			assert.True(t, HasCrashingPod(s.Pods))
			assert.True(t, PodCrashLoopDetected(s.Pods))
			assert.True(t, PodErrorDetected(s.Pods))
		}},
		{"recovered", func(t *testing.T, s liveSnapshot) {
			// Crashed three times, then ran fine: the restarts are history.
			assert.Equal(t, int64(3), PodMaxRestarts(s.Pods))
			assert.False(t, HasCrashingPod(s.Pods))
			assert.Empty(t, PodContainerReasons(s.Pods))
			assert.True(t, RolloutComplete(s.Workload))
			assert.True(t, ConditionsHealthy(s.Workload))
		}},
		{"never-ready", func(t *testing.T, s liveSnapshot) {
			// Running, never Ready, nothing crashing.
			assert.False(t, RolloutComplete(s.Workload))
			assert.False(t, ConditionsHealthy(s.Workload))
			assert.Equal(t, "ProgressDeadlineExceeded", ConditionReason(s.Workload, "Progressing"))
			assert.Equal(t, 0, ReadyPodCount(s.Pods))
			assert.False(t, HasCrashingPod(s.Pods))
		}},
		{"stuck-rollout", func(t *testing.T, s liveSnapshot) {
			// Old pods still serve, so Available stays True; the new one can't pull.
			assert.True(t, HasCondition(s.Workload, "Available"))
			assert.True(t, AllReplicasReady(s.Workload))
			assert.False(t, RolloutComplete(s.Workload))
			assert.False(t, ConditionsHealthy(s.Workload))
			assert.Equal(t, "ProgressDeadlineExceeded", ConditionReason(s.Workload, "Progressing"))
			assert.True(t, PodImagePullBackOffDetected(s.Pods))
		}},
		{"job-failed", func(t *testing.T, s liveSnapshot) {
			assert.True(t, JobFailed(s.Workload))
			assert.Equal(t, int64(3), JobFirstExitCode(s.Pods))
			assert.True(t, HasCrashingPod(s.Pods))
			assert.False(t, PodCrashLoopDetected(s.Pods), "restartPolicy Never: failed, not looping")
		}},
		{"job-succeeded", func(t *testing.T, s liveSnapshot) {
			assert.True(t, JobSucceeded(s.Workload))
			assert.Equal(t, int64(0), JobFirstExitCode(s.Pods))
			assert.False(t, HasCrashingPod(s.Pods))
		}},
	}
	for _, shape := range []string{"map", "typed"} {
		for _, c := range cases {
			t.Run(shape+"/"+c.name, func(t *testing.T) {
				c.check(t, loadLive(t, c.name, shape))
			})
		}
	}
}

// Every snapshot in testdata/live has a case above.
func TestLiveSnapshotsCovered(t *testing.T) {
	files, _ := filepath.Glob(testutil.Testdata("live", "*.json"))
	assert.Len(t, files, 14)
}
