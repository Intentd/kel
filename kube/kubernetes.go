package kube

import (
	"fmt"
	"text/template"

	"github.com/intentd/kel/internal/anymap"
)

// kubernetesFuncs — Kubernetes
//
// Domain: kubernetes
//
// Answer the questions about any Kubernetes object that every controller asks:
// its conditions and their reasons, whether it is ready, synced, terminating
// or exists, and the fields of its spec and status, read safely.
func kubernetesFuncs() template.FuncMap {
	return template.FuncMap{
		// ── Metadata ──────────────────────────────────────────────────────────
		"meta":        Meta,
		"name":        Name,
		"namespace":   Namespace,
		"labels":      Labels,
		"annotations": Annotations,

		// ── Single-key label/annotation accessors ─────────────────────────────
		"getLabel":         GetLabel,
		"getLabelInt":      GetLabelInt,
		"hasLabel":         HasLabel,
		"getAnnotation":    GetAnnotation,
		"getAnnotationInt": GetAnnotationInt,
		"hasAnnotation":    HasAnnotation,
		"labelMatches":     LabelMatches,

		// ── Spec / Status / Phase ─────────────────────────────────────────────
		"spec":      Spec,
		"status":    Status,
		"getStatus": GetStatus,
		"hasStatus": HasStatus,
		"phase":     Phase,

		// ── Safe nested field lookup ──────────────────────────────────────────
		"get": Get,

		// ── Owner references ──────────────────────────────────────────────────
		"ownerKind": OwnerKind,
		"ownerName": OwnerName,

		// ── Conditions ────────────────────────────────────────────────────────
		"conditionsHealthy": ConditionsHealthy,
		"hasCondition":      HasCondition,
		"conditionReason":   ConditionReason,
		"conditionMessage":  ConditionMessage,

		// ── Existence and lifecycle ───────────────────────────────────────────
		"resourceExists": Exists,
		"isTerminating":  IsTerminating,

		// ── Generation sync ───────────────────────────────────────────────────
		"generation":         Generation,
		"observedGeneration": ObservedGeneration,
		"isSynced":           IsSynced,

		// ── Spec resource request accessors ──────────────────────────────────
		"resourceCPU":    ResourceCPU,
		"resourceMemory": ResourceMemory,
	}
}

//
// ────────────────────────────────────────────────────────────────
//   1. METADATA HELPERS
// ────────────────────────────────────────────────────────────────
//

// Meta returns the `metadata` map of a Kubernetes object. Returns an empty map
// if metadata is absent — never nil.
//
// Keywords: kubernetes, metadata, object, navigate, map
//
// Example:
//
//	{{ (meta .deployment).name }}  → "web"
func Meta(obj interface{}) map[string]interface{} {
	if m, ok := asObject(obj); ok {
		if meta, ok := m["metadata"].(map[string]interface{}); ok {
			return meta
		}
	}
	return map[string]interface{}{}
}

// Name returns metadata.name or an empty string.
//
// Keywords: name
//
// Example:
//
//	{{ name .deployment }}  → "web"
func Name(obj interface{}) string {
	meta := Meta(obj)
	if n, ok := meta["name"].(string); ok {
		return n
	}
	return ""
}

// Namespace returns metadata.namespace or an empty string.
//
// Keywords: namespace
//
// Example:
//
//	{{ namespace .deployment }}  → "payments"
func Namespace(obj interface{}) string {
	meta := Meta(obj)
	if ns, ok := meta["namespace"].(string); ok {
		return ns
	}
	return ""
}

// Labels returns the `metadata.labels` map. Returns an empty map if absent.
//
// Keywords: kubernetes, metadata, labels, map, selector, tags
//
// Example:
//
//	{{ mapGet (labels .deployment) "tier" }}  → "frontend"
func Labels(obj interface{}) Map {
	meta := Meta(obj)
	if lbls, ok := meta["labels"].(map[string]interface{}); ok {
		return lbls
	}
	return Map{}
}

// Annotations returns the `metadata.annotations` map. Returns an empty map if
// absent.
//
// Keywords: kubernetes, metadata, annotations, map, tags
//
// Example:
//
//	{{ mapGet (annotations .deployment) "example.com/owner" }}  → "team-payments"
func Annotations(obj interface{}) Map {
	meta := Meta(obj)
	if ann, ok := meta["annotations"].(map[string]interface{}); ok {
		return ann
	}
	return Map{}
}

// Map is a map of an object's fields as the API returns them, such as its
// labels or annotations. It is a map[string]interface{}, so templates and the
// map functions read it as one.
type Map map[string]interface{}

// StringMap returns the map's string values by key, the type Go APIs take for
// labels and annotations, such as labels.Set. Values that are not strings are
// left out. Never nil.
//
//	sel.Matches(labels.Set(kube.Labels(pod).StringMap()))
func (m Map) StringMap() map[string]string {
	return anymap.Strings(m)
}

// ── Single-key label/annotation accessors ─────────────────────────────────────

// GetLabel returns the value of a single label key from `metadata.labels`.
// Returns `""` when the key is absent or the object has no labels.
//
// Keywords: kubernetes, label, metadata, get, string, access, selector
//
// Example:
//
//	{{ getLabel .deployment "app" }}                     → "web"
//	{{ getLabel .deployment "profile" | default "medium" }}  → "medium"
func GetLabel(obj interface{}, key string) string {
	v, _ := Labels(obj)[key].(string)
	return v
}

// GetLabelInt returns a label value parsed as `int64`. Returns `0` when the
// key is absent or the value is non-numeric.
//
// Keywords: kubernetes, label, metadata, int, numeric, replicas, autoscale
//
// Example:
//
//	{{ getLabelInt .deployment "example.com/shards" }}  → 4
func GetLabelInt(obj interface{}, key string) int64 {
	return toInt64(Labels(obj)[key])
}

// HasLabel returns `true` when the object has the given label key with a
// non-empty value.
//
// Keywords: kubernetes, label, metadata, boolean, check, exists, selector
//
// Example:
//
//	{{ hasLabel .deployment "app" }}  → true
func HasLabel(obj interface{}, key string) bool {
	return GetLabel(obj, key) != ""
}

// GetAnnotation returns the value of a single annotation key from
// `metadata.annotations`. Returns `""` when the key is absent or the object
// has no annotations.
//
// Keywords: kubernetes, annotation, metadata, get, string, access, config
//
// Example:
//
//	{{ getAnnotation .deployment "example.com/owner" }}  → "team-payments"
func GetAnnotation(obj interface{}, key string) string {
	v, _ := Annotations(obj)[key].(string)
	return v
}

// GetAnnotationInt returns an annotation value parsed as `int64`. Returns `0`
// when the key is absent or the value is non-numeric. Use with autoscale
// targets and thresholds stored in annotations.
//
// Keywords: kubernetes, annotation, metadata, int, numeric, autoscale, replicas, threshold
//
// Example:
//
//	{{ getAnnotationInt .deployment "deployment.kubernetes.io/revision" }}  → 2
func GetAnnotationInt(obj interface{}, key string) int64 {
	return toInt64(Annotations(obj)[key])
}

// HasAnnotation returns `true` when the object has the given annotation key
// with a non-empty value.
//
// Keywords: kubernetes, annotation, metadata, boolean, check, exists, gate
//
// Example:
//
//	{{ hasAnnotation .deployment "example.com/owner" }}  → true
func HasAnnotation(obj interface{}, key string) bool {
	return GetAnnotation(obj, key) != ""
}

// LabelMatches returns `true` when the object's labels contain every key/value
// pair given as arguments. Extra labels on the object are ignored. Useful for
// checking if a resource would be selected by a label selector.
//
// Keywords: kubernetes, label, selector, match, boolean, filter, scope
//
// Example:
//
//	{{ labelMatches .deployment "app" "web" "tier" "frontend" }}  → true
func LabelMatches(obj interface{}, kvs ...string) bool {
	if len(kvs)%2 != 0 {
		return false
	}
	lbls := Labels(obj)
	for i := 0; i < len(kvs); i += 2 {
		key, want := kvs[i], kvs[i+1]
		got, ok := lbls[key].(string)
		if !ok || got != want {
			return false
		}
	}
	return true
}

//
// ────────────────────────────────────────────────────────────────
//   2. SPEC / STATUS HELPERS
// ────────────────────────────────────────────────────────────────
//

// Spec returns the `spec` map of a Kubernetes object. Returns an empty map if
// absent.
//
// Keywords: kubernetes, spec, object, navigate, map
//
// Example:
//
//	{{ (spec .cronjob).schedule }}  → "0 * * * *"
func Spec(obj interface{}) map[string]interface{} {
	if m, ok := asObject(obj); ok {
		if spec, ok := m["spec"].(map[string]interface{}); ok {
			return spec
		}
	}
	return map[string]interface{}{}
}

// Status returns the `status` map of a Kubernetes object. Returns an empty map
// if absent.
//
// Keywords: kubernetes, status, object, navigate, map
//
// Example:
//
//	{{ (status .deployment).readyReplicas }}  → 3
func Status(obj interface{}) map[string]interface{} {
	if m, ok := asObject(obj); ok {
		if status, ok := m["status"].(map[string]interface{}); ok {
			return status
		}
	}
	return map[string]interface{}{}
}

// GetStatus returns the string value of a specific key from `status`. Numeric
// and boolean values are converted to their string form. Returns `""` when the
// field is absent, nil, or a structured value (map, slice) — use `get` or
// `status` for those.
//
// Keywords: kubernetes, status, get, string, field, access
//
// Example:
//
//	{{ getStatus .pod "phase" }}  → "Running"
func GetStatus(obj interface{}, key string) string {
	v := Status(obj)[key]
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case map[string]interface{}, []interface{}:
		return ""
	default:
		return fmt.Sprint(t)
	}
}

// HasStatus returns `true` when the given status field exists and is
// non-empty. Works for both scalar and structured (map, slice) fields.
//
// Keywords: kubernetes, status, boolean, check, exists, present, field
//
// Example:
//
//	{{ hasStatus .service "loadBalancer" }}  → true
func HasStatus(obj interface{}, key string) bool {
	v, ok := Status(obj)[key]
	if !ok || v == nil {
		return false
	}
	switch t := v.(type) {
	case string:
		return t != ""
	case map[string]interface{}:
		return len(t) > 0
	case []interface{}:
		return len(t) > 0
	default:
		return true
	}
}

//
// ────────────────────────────────────────────────────────────────
//   3. SAFE NESTED FIELD LOOKUP
// ────────────────────────────────────────────────────────────────
//

// Get navigates a nested path through a Kubernetes object using variadic
// string segments. Returns `nil` if any segment is missing. Nil-safe at every
// level.
//
// Keywords: kubernetes, navigate, nested, path, safe, access, deep
//
// Example:
//
//	{{ get .cronjob "status" "lastScheduleTime" }}  → "2026-05-19T10:00:00Z"
func Get(obj interface{}, path ...string) interface{} {
	current := obj
	if o := object(obj); o != nil && len(path) > 0 {
		current = o
	}
	for _, p := range path {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = m[p]
		if current == nil {
			return nil
		}
	}
	return current
}

//
// ────────────────────────────────────────────────────────────────
//   4. OWNERREFERENCE HELPERS
// ────────────────────────────────────────────────────────────────
//

// OwnerKind returns the `kind` of the first `ownerReference`.
//
// Keywords: kubernetes, owner, reference, kind, controller
//
// Example:
//
//	{{ ownerKind .replicaset }}  → "Deployment"
func OwnerKind(obj interface{}) string {
	meta := Meta(obj)
	owners, ok := meta["ownerReferences"].([]interface{})
	if !ok || len(owners) == 0 {
		return ""
	}
	if first, ok := owners[0].(map[string]interface{}); ok {
		if kind, ok := first["kind"].(string); ok {
			return kind
		}
	}
	return ""
}

// OwnerName returns the `name` of the first `ownerReference`.
//
// Keywords: kubernetes, owner, reference, name, controller
//
// Example:
//
//	{{ ownerName .replicaset }}  → "web"
func OwnerName(obj interface{}) string {
	meta := Meta(obj)
	owners, ok := meta["ownerReferences"].([]interface{})
	if !ok || len(owners) == 0 {
		return ""
	}
	if first, ok := owners[0].(map[string]interface{}); ok {
		if name, ok := first["name"].(string); ok {
			return name
		}
	}
	return ""
}

//
// ────────────────────────────────────────────────────────────────
//   5. CONDITION HELPERS
// ────────────────────────────────────────────────────────────────
//

// ConditionsHealthy reports whether an object's standard conditions say it is
// healthy, for any kind that reports them: Flux, cert-manager, Crossplane,
// Argo and most operators. It is true when the controller has seen the latest
// spec and either Ready, or Available when there is no Ready, is "True".
// It is false when there are no such conditions, Stalled or Reconciling is
// "True", Progressing or Synced is "False", or status.observedGeneration (or
// Ready's own observedGeneration) is behind metadata.generation. For a
// Deployment or StatefulSet use rolloutComplete as well: Available stays
// "True" while a rollout is stuck on old pods.
//
// Keywords: kubernetes, condition, ready, healthy, status, crd, flux, cert-manager, crossplane, operator
//
// Example:
//
//	{{ conditionsHealthy .helmRelease }}  → true
//	{{ conditionsHealthy .deployment }}  → true
func ConditionsHealthy(obj interface{}) bool {
	conds := map[string]map[string]interface{}{}
	list, _ := Status(obj)["conditions"].([]interface{})
	for _, c := range list {
		if cm, ok := c.(map[string]interface{}); ok {
			if t, _ := cm["type"].(string); t != "" {
				conds[t] = cm
			}
		}
	}
	is := func(condType, want string) bool {
		s, _ := conds[condType]["status"].(string)
		return s == want
	}
	gen := Generation(obj)
	if observed, ok := Status(obj)["observedGeneration"]; ok && gen > 0 && toInt64(observed) < gen {
		return false
	}
	if is("Stalled", "True") || is("Reconciling", "True") || is("Progressing", "False") || is("Synced", "False") {
		return false
	}
	if ready, ok := conds["Ready"]; ok {
		if observed, ok := ready["observedGeneration"]; ok && gen > 0 && toInt64(observed) < gen {
			return false
		}
		return is("Ready", "True")
	}
	return is("Available", "True")
}

// HasCondition returns `true` if `status.conditions` contains a condition of
// the given type with `status: "True"`.
//
// Keywords: kubernetes, condition, status, check, boolean, ready, available
//
// Example:
//
//	{{ hasCondition .deployment "Available" }}  → true
func HasCondition(obj interface{}, condType string) bool {
	status := Status(obj)
	conds, ok := status["conditions"].([]interface{})
	if !ok {
		return false
	}
	for _, c := range conds {
		if cm, ok := c.(map[string]interface{}); ok {
			if t, ok := cm["type"].(string); ok && t == condType {
				if s, ok := cm["status"].(string); ok && s == "True" {
					return true
				}
			}
		}
	}
	return false
}

// Exists reports whether obj is a Kubernetes object: a map with a
// metadata.name. A child that was never created, a nil, or an empty map is
// not.
//
// Keywords: kubernetes, exists, check, boolean, nil, created, present
//
// Example:
//
//	{{ resourceExists .deployment }}  → true
//	{{ resourceExists .missing }}     → false
func Exists(obj interface{}) bool {
	return nestedString(object(obj), "metadata", "name") != ""
}

//   "phase":              Phase,
//   "conditionReason":    ConditionReason,
//   "conditionMessage":   ConditionMessage,
//   "isTerminating":      IsTerminating,
//   "generation":         Generation,
//   "observedGeneration": ObservedGeneration,
//   "isSynced":           IsSynced,

// ── Spec resource request accessors ──────────────────────────────────────────

// ResourceCPU returns `spec.resources.requests.cpu` from any object. Returns
// `""` when spec, resources, requests, or cpu is absent — no nil pointer panic
// at any level.
//
// Keywords: spec, resources, requests, cpu, safe, default, normalize
//
// Example:
//
//	{{ resourceCPU . }}  → "250m"
//	{{ resourceCPU .pod | default "100m" }}  → "100m"
func ResourceCPU(obj interface{}) string {
	m, ok := asObject(obj)
	if !ok {
		return ""
	}
	spec, _ := m["spec"].(map[string]interface{})
	resources, _ := spec["resources"].(map[string]interface{})
	requests, _ := resources["requests"].(map[string]interface{})
	v, _ := requests["cpu"].(string)
	return v
}

// ResourceMemory returns `spec.resources.requests.memory` from any object.
// Returns `""` when any level is absent.
//
// Keywords: spec, resources, requests, memory, safe, default, normalize
//
// Example:
//
//	{{ resourceMemory . }}  → "256Mi"
func ResourceMemory(obj interface{}) string {
	m, ok := asObject(obj)
	if !ok {
		return ""
	}
	spec, _ := m["spec"].(map[string]interface{})
	resources, _ := spec["resources"].(map[string]interface{})
	requests, _ := resources["requests"].(map[string]interface{})
	v, _ := requests["memory"].(string)
	return v
}

// ── Phase ─────────────────────────────────────────────────────────────────────

// Phase returns `status.phase` as a string. Returns `""` when absent.
//
// Keywords: kubernetes, status, phase, string, pod, lifecycle
//
// Example:
//
//	{{ phase .pod }}  → "Running"
func Phase(obj interface{}) string {
	status := Status(obj)
	v, _ := status["phase"].(string)
	return v
}

// ── Condition detail ──────────────────────────────────────────────────────────

// ConditionReason returns the `reason` field of a named condition. Returns
// `""` when absent.
//
// Keywords: kubernetes, condition, reason, status, string, message
//
// Example:
//
//	{{ conditionReason .deployment "Available" }}  → "MinimumReplicasAvailable"
func ConditionReason(obj interface{}, condType string) string {
	return conditionField(obj, condType, "reason")
}

// ConditionMessage returns the `message` field of a named condition. Returns
// `""` when absent.
//
// Keywords: kubernetes, condition, message, status, string, detail
//
// Example:
//
//	{{ conditionMessage .deployment "Available" }}  → "Deployment has minimum availability."
func ConditionMessage(obj interface{}, condType string) string {
	return conditionField(obj, condType, "message")
}

// conditionField is the shared implementation for ConditionReason
// and ConditionMessage.
func conditionField(obj interface{}, condType, field string) string {
	status := Status(obj)
	conds, ok := status["conditions"].([]interface{})
	if !ok {
		return ""
	}
	for _, c := range conds {
		cm, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		t, _ := cm["type"].(string)
		if t != condType {
			continue
		}
		v, _ := cm[field].(string)
		return v
	}
	return ""
}

// ── Lifecycle ─────────────────────────────────────────────────────────────────

// IsTerminating returns `true` when `metadata.deletionTimestamp` is set — the
// object exists but is being deleted.
//
// Keywords: kubernetes, terminating, deleting, boolean, lifecycle, deletion
//
// Example:
//
//	{{ isTerminating .deployment }}  → false
func IsTerminating(obj interface{}) bool {
	if obj == nil {
		return false
	}
	m, ok := asObject(obj)
	if !ok {
		return false
	}
	meta, ok := m["metadata"].(map[string]interface{})
	if !ok {
		return false
	}
	// deletionTimestamp is set when the object is terminating.
	// It is either a non-empty string or a time.Time via JSON unmarshal.
	ts := meta["deletionTimestamp"]
	if ts == nil {
		return false
	}
	// String representation: non-empty means terminating
	if s, ok := ts.(string); ok {
		return s != "" && s != "null"
	}
	// Any other non-nil value is also terminating
	return true
}

// ── Generation tracking ───────────────────────────────────────────────────────

// Generation returns `metadata.generation` as int64. Increments on every spec
// change. Returns `0` when absent.
//
// Keywords: kubernetes, generation, metadata, revision, int
//
// Example:
//
//	{{ generation .deployment }}  → 3
func Generation(obj interface{}) int64 {
	if obj == nil {
		return 0
	}
	m, ok := asObject(obj)
	if !ok {
		return 0
	}
	meta, ok := m["metadata"].(map[string]interface{})
	if !ok {
		return 0
	}
	return toInt64(meta["generation"])
}

// ObservedGeneration returns `status.observedGeneration` as int64. This is the
// generation the controller last acted on.
//
// Keywords: kubernetes, generation, status, observed, revision, int
//
// Example:
//
//	{{ observedGeneration .deployment }}  → 3
func ObservedGeneration(obj interface{}) int64 {
	status := Status(obj)
	return toInt64(status["observedGeneration"])
}

// IsSynced returns `true` when `metadata.generation ==
// status.observedGeneration`, meaning the controller has fully processed the
// current spec. Returns `true` for resources without generation tracking (both
// 0).
//
// Keywords: kubernetes, synced, generation, rollout, boolean, ready, reconciled
//
// Example:
//
//	{{ isSynced .deployment }}  → true
func IsSynced(obj interface{}) bool {
	gen := Generation(obj)
	obsGen := ObservedGeneration(obj)
	// Both 0 means the resource has no generation tracking (statusless types).
	// Treat as synced to avoid blocking.
	if gen == 0 && obsGen == 0 {
		return true
	}
	return gen == obsGen
}

// toInt64 converts JSON number types to int64 safely.
// JSON unmarshals numbers as float64 by default.
func toInt64(v interface{}) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	case string:
		var n int64
		fmt.Sscanf(t, "%d", &n)
		return n
	default:
		return 0
	}
}
