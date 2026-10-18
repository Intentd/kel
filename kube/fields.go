package kube

import "text/template"

// fieldFuncs — Fields
//
// Domain: fields
//
// Read the fields every Kubernetes object has: name, namespace, UID, resource
// version, creation timestamp, labels and annotations.
func fieldFuncs() template.FuncMap {
	return template.FuncMap{
		"resourceName":      ResourceName,
		"resourceNamespace": ResourceNamespace,
		"resourceUID":       ResourceUID,
		"resourceVersion":   ResourceVersion,
		"creationTimestamp": CreationTimestamp,
	}
}

// ResourceName returns `metadata.name`. Returns `""` when absent.
//
// Keywords: kubernetes, metadata, name, field, identity, resource
//
// Example:
//
//	{{ resourceName .deployment }}  → "web"
func ResourceName(obj interface{}) string {
	return metaStringField(obj, "name")
}

// ResourceNamespace returns `metadata.namespace`. Returns `""` for
// cluster-scoped resources.
//
// Keywords: kubernetes, metadata, namespace, field, scope, resource
//
// Example:
//
//	{{ resourceNamespace .deployment }}  → "payments"
func ResourceNamespace(obj interface{}) string {
	return metaStringField(obj, "namespace")
}

// ResourceUID returns `metadata.uid` — the unique identifier assigned by the
// API server. Stable across renames; changes on delete-and-recreate.
//
// Keywords: kubernetes, metadata, uid, identifier, unique, resource
//
// Example:
//
//	{{ resourceUID .deployment }}  → "4b3f8d21-8e3a-4f8c-b9d2-1a2b3c4d5e6f"
func ResourceUID(obj interface{}) string {
	return metaStringField(obj, "uid")
}

// ResourceVersion returns `metadata.resourceVersion` — the etcd revision
// string. Increments on every write. Useful for detecting whether an object
// has changed since last observation.
//
// Keywords: kubernetes, metadata, version, revision, etcd, change
//
// Example:
//
//	{{ resourceVersion .deployment }}  → "14872"
func ResourceVersion(obj interface{}) string {
	return metaStringField(obj, "resourceVersion")
}

// CreationTimestamp returns `metadata.creationTimestamp` as an RFC3339 string.
//
// Keywords: kubernetes, metadata, timestamp, created, time, age, rfc3339
//
// Example:
//
//	{{ creationTimestamp .deployment }}  → "2026-01-15T10:30:00Z"
//	{{ timeSince (creationTimestamp .deployment) }}
func CreationTimestamp(obj interface{}) string {
	return metaStringField(obj, "creationTimestamp")
}

// metaStringField extracts a string field from metadata.
func metaStringField(obj interface{}, field string) string {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return ""
	}
	meta, ok := m["metadata"].(map[string]interface{})
	if !ok {
		return ""
	}
	v, _ := meta[field].(string)
	return v
}
