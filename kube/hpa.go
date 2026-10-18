package kube

import "text/template"

// hpaFuncs — HPA
//
// Domain: hpa
//
// Read a HorizontalPodAutoscaler: current, desired, minimum and maximum
// replicas, whether it is scaling or at its limit, its conditions, its scale
// target and the types of metric it reports.
func hpaFuncs() template.FuncMap {
	return template.FuncMap{
		"hpaCurrentReplicas": HPACurrentReplicas,
		"hpaDesiredReplicas": HPADesiredReplicas,
		"hpaMinReplicas":     HPAMinReplicas,
		"hpaMaxReplicas":     HPAMaxReplicas,
		"hpaScaling":         HPAScaling,
		"hpaAtMax":           HPAAtMax,
		// From status.conditions.
		"hpaScalingActive":   HPAScalingActive,
		"hpaAbleToScale":     HPAAbleToScale,
		"hpaScalingLimited":  HPAScalingLimited,
		"hpaScaleTargetName": HPAScaleTargetName,
		"hpaScaleTargetKind": HPAScaleTargetKind,
		"hpaMetricTypes":     HPAMetricTypes,
	}
}

// HPACurrentReplicas returns `status.currentReplicas` as `int64`.
//
// Keywords: hpa, autoscaler, replicas, current, int, scale
//
// Example:
//
//	{{ hpaCurrentReplicas .hpa }}  → 3
func HPACurrentReplicas(obj interface{}) int64 {
	status := Status(obj)
	return toInt64(status["currentReplicas"])
}

// HPADesiredReplicas returns `status.desiredReplicas` as `int64`.
//
// Keywords: hpa, autoscaler, replicas, desired, int, scale, target
//
// Example:
//
//	{{ hpaDesiredReplicas .hpa }}  → 5
func HPADesiredReplicas(obj interface{}) int64 {
	status := Status(obj)
	return toInt64(status["desiredReplicas"])
}

// HPAMinReplicas returns `spec.minReplicas` as `int64`, defaulting to `1` when
// not set (matching the Kubernetes default).
//
// Keywords: hpa, autoscaler, replicas, minimum, int, bound, floor
//
// Example:
//
//	{{ hpaMinReplicas .hpa }}  → 2
func HPAMinReplicas(obj interface{}) int64 {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return 1
	}
	spec, _ := m["spec"].(map[string]interface{})
	if spec == nil {
		return 1
	}
	v := toInt64(spec["minReplicas"])
	if v == 0 {
		return 1
	}
	return v
}

// HPAMaxReplicas returns `spec.maxReplicas` as `int64`.
//
// Keywords: hpa, autoscaler, replicas, maximum, int, bound, ceiling, capacity
//
// Example:
//
//	{{ hpaMaxReplicas .hpa }}  → 10
func HPAMaxReplicas(obj interface{}) int64 {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return 0
	}
	spec, _ := m["spec"].(map[string]interface{})
	if spec == nil {
		return 0
	}
	return toInt64(spec["maxReplicas"])
}

// HPAScaling returns `true` when the HPA has a valid metric
// (`ScalingActive=True`) **and** `currentReplicas != desiredReplicas` — a real
// scale-out or scale-in is in progress. Returns `false` when the metric source
// is unknown, which would otherwise cause a misleading `true` due to
// `desiredReplicas=0`.
//
// Keywords: hpa, autoscaler, scaling, active, boolean, in-progress, replicas
//
// Example:
//
//	{{ hpaScaling .hpa }}  → true
func HPAScaling(obj interface{}) bool {
	if !HPAScalingActive(obj) {
		return false
	}
	return HPACurrentReplicas(obj) != HPADesiredReplicas(obj)
}

// HPAAtMax returns `true` when `currentReplicas >= maxReplicas` — the HPA has
// hit its ceiling.
//
// Keywords: hpa, autoscaler, replicas, capacity, boolean, ceiling, max
//
// Example:
//
//	{{ hpaAtMax .hpa }}  → false
func HPAAtMax(obj interface{}) bool {
	max := HPAMaxReplicas(obj)
	if max == 0 {
		return false
	}
	return HPACurrentReplicas(obj) >= max
}

// HPAScalingActive returns `true` when the `ScalingActive` condition is `True`
// — the HPA can read its metric and is permitted to scale. `False` when the
// metric source is unknown (`<unknown>/80%`) or the HPA is otherwise disabled.
//
// Keywords: hpa, autoscaler, condition, metric, boolean, active, scalingactive
//
// Example:
//
//	{{ hpaScalingActive .hpa }}  → true
func HPAScalingActive(obj interface{}) bool {
	return HasCondition(obj, "ScalingActive")
}

// HPAAbleToScale returns `true` when the `AbleToScale` condition is `True` —
// the scale target exists and is reachable.
//
// Keywords: hpa, autoscaler, condition, target, boolean, abletoscale, reachable
//
// Example:
//
//	{{ hpaAbleToScale .hpa }}  → true
func HPAAbleToScale(obj interface{}) bool {
	return HasCondition(obj, "AbleToScale")
}

// HPAScalingLimited returns `true` when the `ScalingLimited` condition is
// `True` — the desired replica count was clamped by the min or max bound.
//
// Keywords: hpa, autoscaler, condition, clamped, boolean, scalinglimited, bound
//
// Example:
//
//	{{ hpaScalingLimited .hpa }}  → false
func HPAScalingLimited(obj interface{}) bool {
	return HasCondition(obj, "ScalingLimited")
}

// ── Scale target and metrics, from the HPA itself ─────────────────────────

// HPAScaleTargetName returns the name of the workload the HPA scales
// (spec.scaleTargetRef.name).
//
// Keywords: hpa, autoscaler, target, name, string, deployment
//
// Example:
//
//	{{ hpaScaleTargetName .hpa }}  → "web"
func HPAScaleTargetName(hpa interface{}) string {
	return nestedString(object(hpa), "spec", "scaleTargetRef", "name")
}

// HPAScaleTargetKind returns the kind of the workload the HPA scales
// (spec.scaleTargetRef.kind).
//
// Keywords: hpa, autoscaler, target, kind, string, deployment
//
// Example:
//
//	{{ hpaScaleTargetKind .hpa }}  → "Deployment"
func HPAScaleTargetKind(hpa interface{}) string {
	return nestedString(object(hpa), "spec", "scaleTargetRef", "kind")
}

// HPAMetricTypes returns the type of each metric the HPA currently reports
// (status.currentMetrics): Resource, Pods, Object, External,
// ContainerResource.
//
// Keywords: hpa, autoscaler, metrics, types, string, resource, external
//
// Example:
//
//	{{ hpaMetricTypes .hpa }}  → [Resource External]
func HPAMetricTypes(hpa interface{}) []string {
	metrics, _ := nested(object(hpa), "status", "currentMetrics").([]interface{})
	var out []string
	for _, m := range metrics {
		if t := nestedString(object(m), "type"); t != "" {
			out = append(out, t)
		}
	}
	return out
}
