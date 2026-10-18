package kube

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Fixtures in the shape the API returns.

type m = map[string]interface{}
type l = []interface{}

func pod(name, phase, node, ip string, ready bool, containers ...m) m {
	readyStatus := "False"
	if ready {
		readyStatus = "True"
	}
	cs := make(l, len(containers))
	for i, c := range containers {
		cs[i] = c
	}
	return m{
		"apiVersion": "v1", "kind": "Pod",
		"metadata": m{"name": name},
		"spec":     m{"nodeName": node},
		"status": m{
			"phase":             phase,
			"podIP":             ip,
			"conditions":        l{m{"type": "Ready", "status": readyStatus}},
			"containerStatuses": cs,
		},
	}
}

func running(name string, restarts int64) m {
	return m{"name": name, "restartCount": restarts, "state": m{"running": m{"startedAt": "2026-10-18T10:00:00Z"}}}
}

func waiting(name, reason string, restarts int64) m {
	return m{"name": name, "restartCount": restarts, "state": m{"waiting": m{"reason": reason}}}
}

func terminated(name, reason string, exit int64) m {
	return m{"name": name, "restartCount": int64(0), "state": m{"terminated": m{"reason": reason, "exitCode": exit}}}
}

func TestPods(t *testing.T) {
	pods := l{
		pod("web-0", "Running", "node-a", "10.0.0.1", true, running("app", 1)),
		pod("web-1", "Running", "node-b", "10.0.0.2", false, waiting("app", "CrashLoopBackOff", 5)),
		pod("web-2", "Pending", "", "", false, waiting("app", "ImagePullBackOff", 0)),
	}

	assert.Equal(t, []string{"web-0", "web-1", "web-2"}, PodNames(pods))
	assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, PodIPs(pods), "pods without an IP are skipped")
	assert.Equal(t, []string{"Running", "Running", "Pending"}, PodPhases(pods))
	assert.Equal(t, []string{"node-a", "node-b"}, PodNodes(pods), "pending pods are skipped")
	assert.Equal(t, 3, PodCount(pods))
	assert.Equal(t, 1, ReadyPodCount(pods))
	assert.Equal(t, int64(5), PodMaxRestarts(pods))
	assert.True(t, HasCrashingPod(pods))
	assert.True(t, PodCrashLoopDetected(pods))
	assert.True(t, PodImagePullBackOffDetected(pods))
	assert.False(t, PodOOMKilledDetected(pods))
	assert.Equal(t, []string{"CrashLoopBackOff", "ImagePullBackOff"}, PodContainerReasons(pods))
	assert.Equal(t, "Waiting", PodContainerState(pods, 1, "app"))
	assert.Equal(t, "", PodContainerState(pods, 7, "app"))
	assert.Equal(t, "10.0.0.2", nestedString(PodByOrdinal(pods, 1), "status", "podIP"))
	assert.Nil(t, PodByOrdinal(pods, 9))

	healthy := l{pod("ok-0", "Running", "n", "10.0.0.9", true, running("app", 0))}
	assert.False(t, HasCrashingPod(healthy))
	assert.False(t, HasCrashingPod(l{pod("r-0", "Running", "n", "", true, running("app", 3))}), "running again: past restarts are history")

	podList := m{"kind": "PodList", "items": pods}
	assert.Equal(t, 3, PodCount(podList), "a PodList works too")

	for _, none := range []interface{}{nil, l{}, "nope", m{}} {
		assert.Zero(t, PodCount(none))
		assert.Empty(t, PodNames(none))
		assert.False(t, HasCrashingPod(none))
	}
}

func TestJobPods(t *testing.T) {
	pods := l{
		pod("job-a", "Succeeded", "n", "", false, terminated("main", "Completed", 0)),
		pod("job-b", "Failed", "n", "", false, terminated("main", "Error", 2)),
		pod("job-c", "Running", "n", "", true, running("main", 0)),
	}
	assert.Equal(t, int64(0), JobFirstExitCode(pods))
	assert.Equal(t, int64(-1), JobFirstExitCode(l{pod("j", "Running", "n", "", true, running("main", 0))}))
	assert.Equal(t, []string{"job-c"}, JobActivePodNames(pods))
	assert.Equal(t, []string{"job-a"}, JobSucceededPodNames(pods))
	assert.Equal(t, []string{"job-b"}, JobFailedPodNames(pods))
}

func TestWarnings(t *testing.T) {
	events := l{
		m{"type": "Normal", "reason": "Scheduled", "message": "assigned"},
		m{"type": "Warning", "reason": "BackOff", "message": "Back-off restarting failed container"},
		m{"type": "Warning", "reason": "Unhealthy", "note": "Liveness probe failed"}, // events.k8s.io/v1
	}
	assert.True(t, HasWarnings(events))
	assert.Equal(t, 2, WarningCount(events), "Normal events do not count")
	assert.Equal(t, "Back-off restarting failed container", FirstWarning(events))
	assert.Equal(t, []string{"Back-off restarting failed container", "Liveness probe failed"}, WarningMessages(events))
	assert.Equal(t, []string{"BackOff", "Unhealthy"}, WarningReasons(events))
	assert.False(t, HasWarnings(l{m{"type": "Normal"}}))
	assert.Equal(t, "", FirstWarning(nil))
}

func TestNodes(t *testing.T) {
	node := m{
		"metadata": m{"name": "node-a", "labels": m{
			"topology.kubernetes.io/zone":      "us-east-2a",
			"topology.kubernetes.io/region":    "us-east-2",
			"node.kubernetes.io/instance-type": "t3.medium",
		}},
		"spec": m{"taints": l{m{"key": "node.kubernetes.io/not-ready", "effect": "NoSchedule"}}},
	}
	assert.Equal(t, "us-east-2a", NodeZone(node))
	assert.Equal(t, "us-east-2", NodeRegion(node))
	assert.Equal(t, "t3.medium", NodeInstanceType(node))
	assert.Equal(t, []string{"node.kubernetes.io/not-ready"}, NodeTaints(node))
	assert.Equal(t, "node-a", PodNodeName(pod("p", "Running", "node-a", "", true)))
	assert.Equal(t, "", NodeZone(nil))
}

func TestReplicaSets(t *testing.T) {
	rs := func(name, rev string) m {
		return m{"metadata": m{
			"name":        name,
			"annotations": m{"deployment.kubernetes.io/revision": rev},
			"ownerReferences": l{
				m{"kind": "Deployment", "name": "web", "controller": true},
			},
		}}
	}
	deployment := m{"metadata": m{"name": "web", "annotations": m{"deployment.kubernetes.io/revision": "3"}}}
	sets := l{rs("web-a", "2"), rs("web-b", "3"), rs("web-c", "1")}

	assert.Equal(t, "web", ReplicaSetOwnerName(rs("web-a", "2")))
	assert.Equal(t, "Deployment", ReplicaSetOwnerKind(rs("web-a", "2")))
	assert.Equal(t, "", ReplicaSetOwnerName(m{"metadata": m{}}))
	assert.Equal(t, 3, DeploymentReplicaSetCount(sets))
	assert.Equal(t, []string{"web-a", "web-b", "web-c"}, DeploymentReplicaSetNames(sets))
	assert.Equal(t, []string{"web-a", "web-c"}, DeploymentOldReplicaSetNames(deployment, sets), "all but the current revision")
	assert.Empty(t, DeploymentOldReplicaSetNames(m{"metadata": m{}}, sets), "no revision yet")
}

func TestVolumes(t *testing.T) {
	pvc := m{"spec": m{"accessModes": l{"ReadWriteOnce", "ReadOnlyMany"}}}
	pv := m{
		"metadata": m{"annotations": m{"pv.kubernetes.io/provisioned-by": "ebs.csi.aws.com"}},
		"spec":     m{"volumeMode": "Filesystem"},
	}
	sc := m{"provisioner": "ebs.csi.aws.com", "reclaimPolicy": "Retain"}
	assert.Equal(t, []string{"ReadWriteOnce", "ReadOnlyMany"}, PVCAccessModes(pvc))
	assert.Equal(t, "ebs.csi.aws.com", PVProvisioner(pv))
	assert.Equal(t, "Filesystem", PVVolumeMode(pv))
	assert.Equal(t, "ebs.csi.aws.com", StorageClassProvisioner(sc))
	assert.Equal(t, "Retain", StorageClassReclaimPolicy(sc), "reclaimPolicy is top-level on a StorageClass")
}

func TestEndpoints(t *testing.T) {
	slice := m{
		"kind":  "EndpointSlice",
		"ports": l{m{"port": int64(8080)}},
		"endpoints": l{
			m{"addresses": l{"10.0.0.1"}, "conditions": m{"ready": true}},
			m{"addresses": l{"10.0.0.2"}, "conditions": m{"ready": false}},
		},
	}
	assert.True(t, HasEndpoints(slice))
	assert.Equal(t, []string{"10.0.0.1:8080"}, ServiceEndpoints(slice), "ready only")
	assert.Equal(t, 2, ServiceEndpointCount(l{slice}), "a list of slices works too")
	assert.Equal(t, "10.0.0.1:8080", ServiceFirstEndpoint(slice))
	assert.False(t, HasEndpoints(nil))
}

func TestHPAAndIngress(t *testing.T) {
	hpa := m{
		"spec":   m{"scaleTargetRef": m{"kind": "Deployment", "name": "web"}},
		"status": m{"currentMetrics": l{m{"type": "Resource"}, m{"type": "External"}}},
	}
	assert.Equal(t, "web", HPAScaleTargetName(hpa))
	assert.Equal(t, "Deployment", HPAScaleTargetKind(hpa))
	assert.Equal(t, []string{"Resource", "External"}, HPAMetricTypes(hpa))

	ing := m{
		"spec": m{
			"rules": l{m{"host": "api.example.com"}, m{}, m{"host": "www.example.com"}},
			"tls":   l{m{"hosts": l{"api.example.com", "api.example.com"}}},
		},
		"status": m{"loadBalancer": m{"ingress": l{m{"ip": "1.2.3.4"}, m{"hostname": "lb.example.com"}}}},
	}
	assert.Equal(t, []string{"api.example.com", "www.example.com"}, IngressRules(ing))
	assert.Equal(t, []string{"api.example.com"}, IngressTLSHosts(ing))
	assert.Equal(t, []string{"1.2.3.4", "lb.example.com"}, IngressLoadBalancerIPs(ing))
}

func TestContainerPortsAndExists(t *testing.T) {
	deployment := m{
		"metadata": m{"name": "web"},
		"spec": m{"template": m{"spec": m{"containers": l{
			m{"name": "app", "ports": l{m{"containerPort": int64(8080)}, m{"containerPort": int64(9090)}}},
		}}}},
	}
	assert.Equal(t, []int64{8080, 9090}, ContainerPorts(deployment, 0))
	assert.Nil(t, ContainerPorts(deployment, 3))
	assert.True(t, Exists(deployment))
	assert.False(t, Exists(m{"status": m{}}), "a placeholder without a name is not an object")
	assert.False(t, Exists(nil))
}

func TestRolloutComplete(t *testing.T) {
	workload := func(kind string, gen, observed int64, spec, status m) m {
		return m{"kind": kind, "metadata": m{"generation": gen}, "spec": spec, "status": merge(status, m{"observedGeneration": observed})}
	}
	deploy := func(gen, observed int64, replicas, updated, available, total int64) m {
		return workload("Deployment", gen, observed, m{"replicas": replicas},
			m{"updatedReplicas": updated, "availableReplicas": available, "replicas": total})
	}
	assert.True(t, RolloutComplete(deploy(2, 2, 3, 3, 3, 3)))
	assert.False(t, RolloutComplete(deploy(3, 2, 3, 3, 3, 3)), "the controller has not seen the change")
	assert.False(t, RolloutComplete(deploy(2, 2, 3, 3, 3, 4)), "an old replica is still terminating")
	assert.False(t, RolloutComplete(deploy(2, 2, 3, 3, 2, 3)), "a new replica is not available")

	sts := func(partition int64, ready, updated int64, cur, upd string) m {
		spec := m{"replicas": int64(3)}
		if partition > 0 {
			spec["updateStrategy"] = m{"rollingUpdate": m{"partition": partition}}
		}
		return workload("StatefulSet", 1, 1, spec, m{"readyReplicas": ready, "updatedReplicas": updated, "currentRevision": cur, "updateRevision": upd})
	}
	assert.True(t, RolloutComplete(sts(0, 3, 3, "r2", "r2")))
	assert.False(t, RolloutComplete(sts(0, 3, 3, "r1", "r2")), "revisions differ")
	assert.True(t, RolloutComplete(sts(2, 3, 1, "r1", "r2")), "pods from the partition up are updated")
	assert.False(t, RolloutComplete(sts(0, 2, 3, "r2", "r2")), "a replica is not ready")
	assert.Equal(t, 2, StatefulSetPartition(sts(2, 3, 1, "r1", "r2")))
	assert.Equal(t, 0, StatefulSetPartition(sts(0, 3, 3, "r", "r")))

	ds := workload("DaemonSet", 1, 1, m{}, m{"desiredNumberScheduled": int64(4), "updatedNumberScheduled": int64(4), "numberAvailable": int64(3)})
	assert.False(t, RolloutComplete(ds), "a node is not available")
	assert.False(t, RolloutComplete(nil))
}

func merge(a, b m) m {
	out := m{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func TestCronJobNext(t *testing.T) {
	now := time.Date(2026, 5, 19, 10, 30, 0, 0, time.UTC)
	cj := func(spec map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{"kind": "CronJob", "spec": spec}
	}
	tests := []struct {
		name    string
		cronJob interface{}
		want    string
		wantErr bool
	}{
		{"hourly", cj(map[string]interface{}{"schedule": "0 * * * *"}), "2026-05-19T11:00:00Z", false},
		{"macro", cj(map[string]interface{}{"schedule": "@daily"}), "2026-05-20T00:00:00Z", false},
		{"time zone", cj(map[string]interface{}{"schedule": "0 9 * * *", "timeZone": "America/New_York"}), "2026-05-19T13:00:00Z", false},
		{"suspended", cj(map[string]interface{}{"schedule": "0 * * * *", "suspend": true}), "", false},
		{"no schedule", cj(map[string]interface{}{}), "", false},
		{"nil", nil, "", false},
		{"bad schedule", cj(map[string]interface{}{"schedule": "61 * * * *"}), "", true},
		{"bad time zone", cj(map[string]interface{}{"schedule": "0 * * * *", "timeZone": "Mars/Olympus"}), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cronJobNext(tt.cronJob, now)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Errorf("cronJobNext() = %q, %v; want %q, error %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

// withInit adds init container statuses to a pod fixture.
func withInit(p m, inits ...m) m {
	cs := make(l, len(inits))
	for i, c := range inits {
		cs[i] = c
	}
	p["status"].(m)["initContainerStatuses"] = cs
	return p
}

func TestInitContainers(t *testing.T) {
	initializing := waiting("app", "PodInitializing", 0)

	// An init container that keeps failing: the pod never starts (kubectl: Init:CrashLoopBackOff).
	crashing := l{withInit(pod("web-0", "Pending", "n", "", false, initializing), waiting("migrate", "CrashLoopBackOff", 4))}
	assert.True(t, HasCrashingPod(crashing))
	assert.True(t, PodCrashLoopDetected(crashing))
	assert.Equal(t, int64(4), PodMaxRestarts(crashing))
	assert.Equal(t, []string{"CrashLoopBackOff", "PodInitializing"}, PodContainerReasons(crashing), "init containers first, as they run")
	assert.Equal(t, "Waiting", PodContainerState(crashing, 0, "migrate"))

	// Between restarts the init container shows terminated/Error (kubectl: Init:Error).
	failed := l{withInit(pod("web-0", "Pending", "n", "", false, initializing), terminated("migrate", "Error", 1))}
	assert.True(t, HasCrashingPod(failed))
	assert.True(t, PodErrorDetected(failed))
	assert.Equal(t, int64(1), JobFirstExitCode(failed))

	// Init containers that finished are not a problem and add no reasons.
	done := l{withInit(pod("web-0", "Running", "n", "10.0.0.1", true, running("app", 0)), terminated("migrate", "Completed", 0))}
	assert.False(t, HasCrashingPod(done))
	assert.Empty(t, PodContainerReasons(done))
	assert.Equal(t, "Terminated", PodContainerState(done, 0, "migrate"), "a finished init container can still be looked up by name")

	// A job whose init container succeeded reports the main container's exit code, not the init's 0.
	job := l{withInit(pod("job-x", "Failed", "n", "", false, terminated("run", "Error", 3)), terminated("migrate", "Completed", 0))}
	assert.Equal(t, int64(3), JobFirstExitCode(job))

	// A native sidecar (init container with restartPolicy: Always) that crashes.
	sidecar := l{withInit(pod("web-0", "Running", "n", "10.0.0.1", false, running("app", 0)), waiting("proxy", "CrashLoopBackOff", 6))}
	assert.True(t, PodCrashLoopDetected(sidecar))
	assert.Equal(t, int64(6), PodMaxRestarts(sidecar))
}
