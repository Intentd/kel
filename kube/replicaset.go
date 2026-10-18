package kube

import "text/template"

// replicaSetFuncs — ReplicaSet
//
// Domain: replicaset
//
// Read a Deployment's ReplicaSets, passed as a list: how many, their names,
// and which are older than the Deployment's current revision; and any object's
// controller, from its ownerReferences.
func replicaSetFuncs() template.FuncMap {
	return template.FuncMap{
		// Any object's controller, from its own metadata.ownerReferences.
		"replicaSetOwnerName": ReplicaSetOwnerName,
		"replicaSetOwnerKind": ReplicaSetOwnerKind,
		// A Deployment's ReplicaSets, passed as a list.
		"deploymentReplicaSetCount": DeploymentReplicaSetCount,
		"deploymentReplicaSets":     DeploymentReplicaSetNames,
		"oldDeploymentReplicaSets":  DeploymentOldReplicaSetNames,
	}
}

// ReplicaSetOwnerName returns the name of the object's controller (the
// ownerReference with controller: true); for a ReplicaSet, its Deployment.
//
// Keywords: replicaset, owner, deployment, name, string, controller
//
// Example:
//
//	{{ replicaSetOwnerName .replicaset }}  → "web"
func ReplicaSetOwnerName(obj interface{}) string {
	return nestedString(controllerRef(obj), "name")
}

// ReplicaSetOwnerKind returns the kind of the object's controller; for a
// ReplicaSet, typically "Deployment".
//
// Keywords: replicaset, owner, deployment, kind, string, controller
//
// Example:
//
//	{{ replicaSetOwnerKind .replicaset }}  → "Deployment"
func ReplicaSetOwnerKind(obj interface{}) string {
	return nestedString(controllerRef(obj), "kind")
}

// DeploymentReplicaSetCount returns the number of ReplicaSets, including
// scaled-to-zero ones a Deployment keeps for rollback.
//
// Keywords: deployment, replicaset, count, rollback, int, history
//
// Example:
//
//	{{ deploymentReplicaSetCount .replicaSets }}  → 2
func DeploymentReplicaSetCount(replicaSets interface{}) int {
	return len(objectList(replicaSets))
}

// DeploymentReplicaSetNames returns the ReplicaSets' names.
//
// Keywords: deployment, replicaset, names, rollback, list, history, all
//
// Example:
//
//	{{ deploymentReplicaSets .replicaSets }}  → [web-6d4f web-5c8b]
func DeploymentReplicaSetNames(replicaSets interface{}) []string {
	var out []string
	for _, rs := range objectList(replicaSets) {
		if n := nestedString(rs, "metadata", "name"); n != "" {
			out = append(out, n)
		}
	}
	return out
}

// DeploymentOldReplicaSetNames returns the names of the deployment's
// ReplicaSets that are not its current revision: those whose
// deployment.kubernetes.io/revision annotation differs from the Deployment's.
// Empty until the Deployment has a revision.
//
// Keywords: deployment, replicaset, old, rollback, list, inactive, stale
//
// Example:
//
//	{{ oldDeploymentReplicaSets .deployment .replicaSets }}  → [web-5c8b]
func DeploymentOldReplicaSetNames(deployment, replicaSets interface{}) []string {
	current := revision(object(deployment))
	if current == "" {
		return nil
	}
	var out []string
	for _, rs := range objectList(replicaSets) {
		if revision(rs) != current {
			if n := nestedString(rs, "metadata", "name"); n != "" {
				out = append(out, n)
			}
		}
	}
	return out
}

// revision is an object's deployment.kubernetes.io/revision annotation.
func revision(obj map[string]interface{}) string {
	ann, _ := nested(obj, "metadata", "annotations").(map[string]interface{})
	v, _ := ann["deployment.kubernetes.io/revision"].(string)
	return v
}

// controllerRef returns the object's ownerReference with controller: true.
func controllerRef(obj interface{}) map[string]interface{} {
	refs, _ := nested(object(obj), "metadata", "ownerReferences").([]interface{})
	for _, r := range refs {
		ref := object(r)
		if c, _ := ref["controller"].(bool); c {
			return ref
		}
	}
	return nil
}
