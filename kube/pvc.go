package kube

import (
	"text/template"
)

// pvcFuncs — PVC
//
// Domain: pvc
//
// Read a PersistentVolumeClaim (whether it is bound, its phase, capacity,
// storage class and access modes), the PersistentVolume bound to it, and its
// StorageClass.
func pvcFuncs() template.FuncMap {
	return template.FuncMap{
		"pvcBound":                  PVCBound,
		"pvcPhase":                  PVCPhase,
		"pvcCapacity":               PVCCapacity,
		"pvcStorageClass":           PVCStorageClass,
		"pvcAccessModes":            PVCAccessModes,
		"pvProvisioner":             PVProvisioner,
		"pvVolumeMode":              PVVolumeMode,
		"storageClassProvisioner":   StorageClassProvisioner,
		"storageClassReclaimPolicy": StorageClassReclaimPolicy,
	}
}

// PVCBound returns `true` when `status.phase == "Bound"`.
//
// Keywords: pvc, storage, bound, phase, boolean, ready, volume
//
// Example:
//
//	{{ pvcBound .pvc }}  → true
func PVCBound(obj interface{}) bool {
	return PVCPhase(obj) == "Bound"
}

// PVCPhase returns `status.phase` — `"Bound"`, `"Pending"`, `"Released"`, or
// `"Lost"`.
//
// Keywords: pvc, storage, phase, status, string, lifecycle, volume
//
// Example:
//
//	{{ pvcPhase .pvc }}  → "Bound"
func PVCPhase(obj interface{}) string {
	status := Status(obj)
	v, _ := status["phase"].(string)
	return v
}

// PVCCapacity returns `spec.resources.requests.storage`.
//
// Keywords: pvc, storage, capacity, size, string, volume, request
//
// Example:
//
//	{{ pvcCapacity .pvc }}  → "10Gi"
func PVCCapacity(obj interface{}) string {
	spec := Spec(obj)
	if spec == nil {
		return ""
	}
	resources, _ := spec["resources"].(map[string]interface{})
	if resources == nil {
		return ""
	}
	requests, _ := resources["requests"].(map[string]interface{})
	if requests == nil {
		return ""
	}
	v, _ := requests["storage"].(string)
	return v
}

// PVCStorageClass returns `spec.storageClassName`, or `""` when not set
// (cluster default applies).
//
// Keywords: pvc, storage, class, string, provisioner, volume
//
// Example:
//
//	{{ pvcStorageClass .pvc }}  → "standard"
func PVCStorageClass(obj interface{}) string {
	spec := Spec(obj)
	if spec == nil {
		return ""
	}
	v, _ := spec["storageClassName"].(string)
	return v
}

// PVCAccessModes returns the PVC's spec.accessModes.
//
// Keywords: pvc, storage, access, modes, list, string, readwriteonce, readonlymany
//
// Example:
//
//	{{ pvcAccessModes .pvc }}  → [ReadWriteOnce]
func PVCAccessModes(pvc interface{}) []string {
	modes, _ := nested(object(pvc), "spec", "accessModes").([]interface{})
	var out []string
	for _, m := range modes {
		if s, _ := m.(string); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// ── The bound PersistentVolume ────────────────────────────────────────────

// PVProvisioner returns the provisioner that created a PersistentVolume
// (annotation pv.kubernetes.io/provisioned-by).
//
// Keywords: provisioner, pv
//
// Example:
//
//	{{ pvProvisioner .pv }}  → "ebs.csi.aws.com"
func PVProvisioner(pv interface{}) string {
	ann, _ := nested(object(pv), "metadata", "annotations").(map[string]interface{})
	v, _ := ann["pv.kubernetes.io/provisioned-by"].(string)
	return v
}

// PVVolumeMode returns a PersistentVolume's volumeMode: Filesystem or Block.
//
// Keywords: mode, pv, volume
//
// Example:
//
//	{{ pvVolumeMode .pv }}  → "Filesystem"
func PVVolumeMode(pv interface{}) string {
	return nestedString(object(pv), "spec", "volumeMode")
}

// ── StorageClass ──────────────────────────────────────────────────────────

// StorageClassProvisioner returns a StorageClass's provisioner.
//
// Keywords: class, provisioner, storage
//
// Example:
//
//	{{ storageClassProvisioner .storageClass }}  → "ebs.csi.aws.com"
func StorageClassProvisioner(sc interface{}) string {
	return nestedString(object(sc), "provisioner")
}

// StorageClassReclaimPolicy returns a StorageClass's reclaimPolicy: Delete or
// Retain.
//
// Keywords: class, policy, reclaim, storage
//
// Example:
//
//	{{ storageClassReclaimPolicy .storageClass }}  → "Delete"
func StorageClassReclaimPolicy(sc interface{}) string {
	return nestedString(object(sc), "reclaimPolicy")
}
