package kube

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
)

// A crash-looping pod as a typed client-go object; TypeMeta is empty, as it
// is on objects from a typed list call.
func typedCrashingPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "web-1", Namespace: "shop", Labels: map[string]string{"app": "web"}},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{
			Name: "web", Image: "nginx:1.27",
			Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("250m")}},
		}}},
		Status: corev1.PodStatus{
			Phase:      corev1.PodRunning,
			Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionFalse}},
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "web", RestartCount: 5,
				State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}},
			}},
		},
	}
}

func TestObjectShapes(t *testing.T) {
	typed := typedCrashingPod()
	u := unstructured.Unstructured{Object: object(typed)}

	singles := map[string]interface{}{
		"map":                        object(typed),
		"*corev1.Pod":                typed,
		"corev1.Pod":                 *typed,
		"*unstructured.Unstructured": &u,
		"unstructured.Unstructured":  u,
	}
	for name, v := range singles {
		assert.Equal(t, "web-1", Name(v), name)
		assert.Equal(t, "shop", Namespace(v), name)
		assert.Equal(t, "web", GetLabel(v, "app"), name)
		assert.Equal(t, "Pod", Get(v, "kind"), name+": kind comes from the Go type when TypeMeta is empty")
		assert.Equal(t, "Running", Phase(v), name)
		assert.True(t, Exists(v), name)
	}

	lists := map[string]interface{}{
		"[]any of maps":                  []interface{}{object(typed)},
		"[]*corev1.Pod":                  []*corev1.Pod{typed},
		"[]corev1.Pod":                   []corev1.Pod{*typed},
		"*corev1.PodList":                &corev1.PodList{Items: []corev1.Pod{*typed}},
		"corev1.PodList":                 corev1.PodList{Items: []corev1.Pod{*typed}},
		"[]unstructured.Unstructured":    []unstructured.Unstructured{u},
		"*unstructured.UnstructuredList": &unstructured.UnstructuredList{Items: []unstructured.Unstructured{u}},
		"[]any of typed":                 []interface{}{typed},
	}
	for name, v := range lists {
		assert.Equal(t, 1, PodCount(v), name)
		assert.Equal(t, []string{"web-1"}, PodNames(v), name)
		assert.True(t, HasCrashingPod(v), name)
		assert.True(t, PodCrashLoopDetected(v), name)
		assert.Equal(t, int64(5), PodMaxRestarts(v), name)
	}

	// Numbers come through as int64, as from a dynamic client, and quantities as strings.
	assert.Equal(t, int64(5), Get(typed, "status", "containerStatuses").([]interface{})[0].(map[string]interface{})["restartCount"])

	// Not objects: zero values, never a panic.
	for _, none := range []interface{}{nil, (*corev1.Pod)(nil), (*unstructured.Unstructured)(nil), "pod", 3, []string{"a"}, make(chan int)} {
		assert.Equal(t, "", Name(none))
		assert.Equal(t, 0, PodCount(none))
		assert.False(t, HasCrashingPod(none))
	}
}

func TestTypedDeploymentRollout(t *testing.T) {
	two := int32(2)
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "web", Generation: 2},
		Spec:       appsv1.DeploymentSpec{Replicas: &two},
		// New pods created, none available yet: the rollout is not complete.
		Status: appsv1.DeploymentStatus{ObservedGeneration: 2, Replicas: 2, UpdatedReplicas: 2, AvailableReplicas: 0},
	}
	assert.False(t, RolloutComplete(d), "a typed Deployment without TypeMeta gets the Deployment rule")
	d.Status.AvailableReplicas = 2
	assert.True(t, RolloutComplete(d))
	assert.Equal(t, 2, DesiredReplicas(d))
}

// A map without kind, such as JSON from somewhere other than the API client,
// is read by its shape.
func TestRolloutCompleteWithoutKind(t *testing.T) {
	deployment := m{
		"metadata": m{"generation": int64(1)},
		"spec":     m{"replicas": int64(2), "strategy": m{"type": "RollingUpdate"}},
		"status":   m{"observedGeneration": int64(1), "replicas": int64(2), "updatedReplicas": int64(2), "availableReplicas": int64(0)},
	}
	assert.False(t, RolloutComplete(deployment), "a Deployment's strategy gives the Deployment rule")

	statefulSet := m{
		"spec":   m{"replicas": int64(2), "serviceName": "db"},
		"status": m{"readyReplicas": int64(2), "currentRevision": "db-1", "updateRevision": "db-2"},
	}
	assert.False(t, RolloutComplete(statefulSet), "revisions differ: the StatefulSet rule applies")

	daemonSet := m{"status": m{"desiredNumberScheduled": int64(3), "updatedNumberScheduled": int64(3), "numberAvailable": int64(2)}}
	assert.False(t, RolloutComplete(daemonSet), "one node not available: the DaemonSet rule applies")
}

func TestLabelsStringMap(t *testing.T) {
	pod := typedCrashingPod()
	got := Labels(pod).StringMap()
	assert.Equal(t, map[string]string{"app": "web"}, got)
	assert.True(t, labels.SelectorFromSet(labels.Set{"app": "web"}).Matches(labels.Set(got)))
	assert.Equal(t, map[string]string{}, Labels(nil).StringMap(), "never nil")
	assert.Equal(t, map[string]string{}, Annotations(pod).StringMap())

	// Still a map[string]interface{} underneath, for Go callers and templates.
	var plain map[string]interface{} = Labels(pod)
	assert.Equal(t, "web", plain["app"])
}

func TestConditionsHealthy(t *testing.T) {
	cond := func(typ, status string) m { return m{"type": typ, "status": status} }
	obj := func(gen, observed int64, conds ...m) m {
		cs := l{}
		for _, c := range conds {
			cs = append(cs, c)
		}
		return m{"metadata": m{"generation": gen}, "status": m{"observedGeneration": observed, "conditions": cs}}
	}

	assert.True(t, ConditionsHealthy(obj(2, 2, cond("Ready", "True"))), "Ready")
	assert.False(t, ConditionsHealthy(obj(2, 2, cond("Ready", "False"))), "not Ready")
	assert.False(t, ConditionsHealthy(obj(3, 2, cond("Ready", "True"))), "Ready is about an older spec")
	assert.False(t, ConditionsHealthy(obj(2, 2, cond("Ready", "True"), cond("Stalled", "True"))), "Stalled (kstatus)")
	assert.False(t, ConditionsHealthy(obj(2, 2, cond("Ready", "True"), cond("Reconciling", "True"))), "Reconciling (Flux)")
	assert.False(t, ConditionsHealthy(obj(2, 2, cond("Ready", "True"), cond("Synced", "False"))), "Synced False (Crossplane)")
	assert.True(t, ConditionsHealthy(obj(2, 2, cond("Available", "True"), cond("Progressing", "True"))), "Available without Ready")
	assert.False(t, ConditionsHealthy(obj(2, 2, cond("Available", "True"), cond("Progressing", "False"))), "progress deadline exceeded")
	assert.False(t, ConditionsHealthy(obj(2, 2)), "no conditions")
	assert.False(t, ConditionsHealthy(nil))

	// Ready's own observedGeneration (cert-manager, Flux) is checked too.
	staleReady := obj(4, 4, m{"type": "Ready", "status": "True", "observedGeneration": int64(3)})
	assert.False(t, ConditionsHealthy(staleReady))

	// No status.observedGeneration at all: not treated as stale.
	noObserved := m{"metadata": m{"generation": int64(5)}, "status": m{"conditions": l{cond("Ready", "True")}}}
	assert.True(t, ConditionsHealthy(noObserved))
}

// lastTerminated sets a container status's last termination.
func lastTerminated(cs m, reason string, exit int64) m {
	cs["lastState"] = m{"terminated": m{"reason": reason, "exitCode": exit}}
	return cs
}

func TestCrashLoopBetweenBackOffs(t *testing.T) {
	// Between back-offs a crash-looping container shows terminated/Error, not CrashLoopBackOff.
	c := terminated("app", "Error", 1)
	c["restartCount"] = int64(4)
	looping := l{pod("web-0", "Running", "n", "", false, c)}
	assert.True(t, PodCrashLoopDetected(looping))

	// A single failure that won't be restarted (restartPolicy: Never) is not a loop.
	once := l{pod("job-0", "Failed", "n", "", false, terminated("run", "Error", 1))}
	assert.False(t, PodCrashLoopDetected(once))
	assert.True(t, PodErrorDetected(once))
}

func TestLastTermination(t *testing.T) {
	// OOMKilled container in back-off: CrashLoopBackOff now, OOMKilled last.
	oom := l{pod("web-0", "Running", "n", "", false, lastTerminated(waiting("app", "CrashLoopBackOff", 11), "OOMKilled", 137))}
	assert.True(t, PodOOMKilledDetected(oom))
	assert.True(t, PodCrashLoopDetected(oom))
	assert.Equal(t, []string{"CrashLoopBackOff", "OOMKilled"}, PodContainerReasons(oom))

	// Running again after an old OOM kill: history, not a current problem.
	recovered := l{pod("web-0", "Running", "n", "10.0.0.1", true, lastTerminated(running("app", 1), "OOMKilled", 137))}
	assert.False(t, PodOOMKilledDetected(recovered))
	assert.Empty(t, PodContainerReasons(recovered))
}

func TestHasCrashingPodIsAboutNow(t *testing.T) {
	type tc struct {
		name string
		pod  m
		want bool
	}
	failedRun := terminated("app", "Error", 1)
	failedRun["restartCount"] = int64(4)
	for _, c := range []tc{
		{"running, ready, 7 restarts from last week", pod("p", "Running", "n", "", true, running("app", 7)), false},
		{"waiting in CrashLoopBackOff", pod("p", "Running", "n", "", false, waiting("app", "CrashLoopBackOff", 5)), true},
		{"terminated Error between restarts", pod("p", "Running", "n", "", false, failedRun), true},
		{"running again after an OOM kill", pod("p", "Running", "n", "", true, lastTerminated(running("app", 1), "OOMKilled", 137)), false},
		{"waiting after an OOM kill", pod("p", "Running", "n", "", false, lastTerminated(waiting("app", "ContainerCreating", 1), "OOMKilled", 137)), true},
		{"init container Init:Error", withInit(pod("p", "Pending", "n", "", false, waiting("app", "PodInitializing", 0)), terminated("migrate", "Error", 1)), true},
		{"job pod, restartPolicy Never, exited 1", pod("p", "Failed", "n", "", false, terminated("run", "Error", 1)), true},
		{"waiting to start for the first time", pod("p", "Pending", "n", "", false, waiting("app", "ContainerCreating", 0)), false},
	} {
		assert.Equal(t, c.want, HasCrashingPod(l{c.pod}), c.name)
	}
}

func TestPodRestartedWithin(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	restartedAt := func(at string) l {
		cs := running("app", 1)
		cs["lastState"] = m{"terminated": m{"reason": "Error", "exitCode": int64(1), "finishedAt": at}}
		return l{pod("p", "Running", "n", "", true, cs)}
	}
	got, err := podRestartedWithin(restartedAt("2026-10-20T11:55:00Z"), "10m", now)
	assert.NoError(t, err)
	assert.True(t, got, "5 minutes ago")

	got, _ = podRestartedWithin(restartedAt("2026-10-20T11:00:00Z"), "10m", now)
	assert.False(t, got, "an hour ago")

	got, _ = podRestartedWithin(restartedAt("2026-10-19T13:00:00Z"), "1d", now)
	assert.True(t, got, "KEL's day unit works")

	got, _ = podRestartedWithin(l{pod("p", "Running", "n", "", true, running("app", 0))}, "10m", now)
	assert.False(t, got, "never restarted")

	for _, bad := range []string{"", "soon", "never", "-5m"} {
		_, err := podRestartedWithin(restartedAt("2026-10-20T11:55:00Z"), bad, now)
		assert.Error(t, err, bad)
	}
}
