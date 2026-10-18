package kube

import "text/template"

// replicaFuncs — Replica
//
// Domain: replica
//
// Read the replica counts of a workload (desired, ready, available and
// updated) and whether all replicas are ready or a rollout has completed.
func replicaFuncs() template.FuncMap {
	return template.FuncMap{
		"allReplicasReady":  AllReplicasReady,
		"desiredReplicas":   DesiredReplicas,
		"readyReplicas":     ReadyReplicas,
		"availableReplicas": AvailableReplicas,
		"updatedReplicas":   UpdatedReplicas,
		"rolloutComplete":   RolloutComplete,
	}
}

// AllReplicasReady returns `true` when `status.readyReplicas ==
// spec.replicas`. The canonical rollout-complete gate. Returns `true` when
// scaled to zero (desired=0 and ready=0).
//
// Keywords: replica, rollout, ready, deployment, statefulset, gate, complete
//
// Example:
//
//	{{ allReplicasReady .deployment }}  → true
func AllReplicasReady(obj interface{}) bool {
	return ReadyReplicas(obj) == DesiredReplicas(obj)
}

// RolloutComplete reports whether a workload's rollout has finished, by the
// rules kubectl rollout status uses for its kind. In every case the status
// must describe the current spec (status.observedGeneration at least
// metadata.generation), so a change the controller has not seen yet is not
// complete.
//
//   - Deployment: every replica runs the new template, no old replica is
//     left, and the new replicas are available.
//   - StatefulSet: every replica is ready and, with a partition, the replicas
//     from the partition up are updated; without one, the current revision
//     is the update revision.
//   - DaemonSet: every scheduled node runs the new template and is available.
//   - Any other kind: updated replicas equal desired replicas.
//
// Keywords: replica, rollout, complete, deployment, statefulset, daemonset, generation
//
// Example:
//
//	{{ rolloutComplete .deployment }}  → true
func RolloutComplete(obj interface{}) bool {
	o := object(obj)
	if o == nil {
		return false
	}
	gen, observed := toInt64(nested(o, "metadata", "generation")), toInt64(nested(o, "status", "observedGeneration"))
	if gen > 0 && observed < gen {
		return false
	}
	switch nestedString(o, "kind") {
	case "Deployment":
		desired, updated := DesiredReplicas(o), UpdatedReplicas(o)
		total := int(toInt64(nested(o, "status", "replicas")))
		return updated >= desired && total <= updated && AvailableReplicas(o) >= updated
	case "StatefulSet":
		desired := DesiredReplicas(o)
		if ReadyReplicas(o) < desired {
			return false
		}
		if partition := StatefulSetPartition(o); partition > 0 {
			return UpdatedReplicas(o) >= desired-partition
		}
		return StatefulSetCurrentRevision(o) == StatefulSetUpdateRevision(o)
	case "DaemonSet":
		want := toInt64(nested(o, "status", "desiredNumberScheduled"))
		return toInt64(nested(o, "status", "updatedNumberScheduled")) >= want &&
			toInt64(nested(o, "status", "numberAvailable")) >= want
	default:
		return UpdatedReplicas(o) == DesiredReplicas(o)
	}
}

// DesiredReplicas returns `spec.replicas` as `int`. Returns `1` when not set
// (Kubernetes default).
//
// Keywords: replica, spec, desired, count, deployment, int, scale
//
// Example:
//
//	{{ desiredReplicas .deployment }}  → 3
func DesiredReplicas(obj interface{}) int {
	spec := Get(obj, "spec")
	m, ok := spec.(map[string]interface{})
	if !ok {
		return 1
	}
	v, exists := m["replicas"]
	if !exists || v == nil {
		return 1
	}
	return int(toInt64(v))
}

// ReadyReplicas returns `status.readyReplicas` as `int`. Returns `0` when
// absent.
//
// Keywords: replica, status, ready, count, deployment, int
//
// Example:
//
//	{{ readyReplicas .deployment }}  → 3
func ReadyReplicas(obj interface{}) int {
	status := Status(obj)
	return int(toInt64(status["readyReplicas"]))
}

// AvailableReplicas returns `status.availableReplicas` as `int`. Returns `0`
// when absent.
//
// Keywords: replica, status, available, count, deployment, int, minreadyseconds
//
// Example:
//
//	{{ availableReplicas .deployment }}  → 3
func AvailableReplicas(obj interface{}) int {
	status := Status(obj)
	return int(toInt64(status["availableReplicas"]))
}

// UpdatedReplicas returns `status.updatedReplicas` as `int` — pods running the
// current pod template. When `updatedReplicas == desiredReplicas`, the rollout
// is complete.
//
// Keywords: replica, status, updated, count, deployment, int, rollout
//
// Example:
//
//	{{ updatedReplicas .deployment }}  → 3
func UpdatedReplicas(obj interface{}) int {
	status := Status(obj)
	return int(toInt64(status["updatedReplicas"]))
}
