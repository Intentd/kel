package kube

import "text/template"

// statefulSetFuncs — StatefulSet
//
// Domain: statefulset
//
// Read what is particular to a StatefulSet: its current and update revisions,
// and the partition of a staged rolling update. Replica counts, rollout
// completion and pods have their own functions.
func statefulSetFuncs() template.FuncMap {
	return template.FuncMap{
		"statefulSetCurrentRevision": StatefulSetCurrentRevision,
		"statefulSetUpdateRevision":  StatefulSetUpdateRevision,
		"statefulSetPartition":       StatefulSetPartition,
	}
}

// StatefulSetCurrentRevision returns `status.currentRevision` — the pod
// template hash of currently running pods, or `""` before the first rollout.
//
// Keywords: statefulset, revision, rollout, hash, string, current, template
//
// Example:
//
//	{{ statefulSetCurrentRevision .statefulset }}  → "db-6d8f4b9c5"
func StatefulSetCurrentRevision(obj interface{}) string {
	status := Status(obj)
	v, _ := status["currentRevision"].(string)
	return v
}

// StatefulSetUpdateRevision returns `status.updateRevision` — the hash of the
// pending update; equal to `currentRevision` when the rollout is complete.
//
// Keywords: statefulset, revision, rollout, update, string, hash, pending
//
// Example:
//
//	{{ statefulSetUpdateRevision .statefulset }}  → "db-7a2c1d4f8"
func StatefulSetUpdateRevision(obj interface{}) string {
	status := Status(obj)
	v, _ := status["updateRevision"].(string)
	return v
}

// StatefulSetPartition returns the partition of a StatefulSet's rolling
// update (spec.updateStrategy.rollingUpdate.partition): pods with an ordinal
// at or above it are updated, the rest keep the old revision. 0 when unset.
//
// Keywords: statefulset, partition, rollout, canary, staged, update
//
// Example:
//
//	{{ statefulSetPartition .statefulset }}  → 2
func StatefulSetPartition(obj interface{}) int {
	return int(toInt64(nested(object(obj), "spec", "updateStrategy", "rollingUpdate", "partition")))
}
