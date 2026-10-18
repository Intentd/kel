package kube

import (
	"strconv"
	"strings"
	"text/template"
)

// podFuncs — Pods
//
// Domain: pods
//
// Answer questions about a set of Pods, as the API returns them, passed as a
// list or a PodList: names, IPs, phases, nodes, readiness, restarts, and why
// containers are waiting or failed (CrashLoopBackOff, OOMKilled,
// ImagePullBackOff, …).
func podFuncs() template.FuncMap {
	return template.FuncMap{
		"podNames":       PodNames,
		"podIPs":         PodIPs,
		"podPhases":      PodPhases,
		"podNodes":       PodNodes,
		"podCount":       PodCount,
		"readyPodCount":  ReadyPodCount,
		"podMaxRestarts": PodMaxRestarts,
		"hasCrashingPod": HasCrashingPod,
		"podByOrdinal":   PodByOrdinal,
		// Container states: from status.containerStatuses of each pod.
		"podContainerReasons":             PodContainerReasons,
		"podContainerState":               PodContainerState,
		"podCrashLoopDetected":            PodCrashLoopDetected,
		"podImagePullBackOffDetected":     PodImagePullBackOffDetected,
		"podErrImagePullDetected":         PodErrImagePullDetected,
		"podErrorDetected":                PodErrorDetected,
		"podOOMKilledDetected":            PodOOMKilledDetected,
		"podRunContainerErrorDetected":    PodRunContainerErrorDetected,
		"podCreateContainerErrorDetected": PodCreateContainerErrorDetected,
		"podInvalidImageNameDetected":     PodInvalidImageNameDetected,
		"podPreStartHookErrorDetected":    PodPreStartHookErrorDetected,
		"podPostStartHookErrorDetected":   PodPostStartHookErrorDetected,
	}
}

// crashReasons are the container waiting/terminated reasons that mean a pod
// is not running as intended.
var crashReasons = []string{
	"CrashLoopBackOff", "OOMKilled", "Error",
	"ImagePullBackOff", "ErrImagePull", "InvalidImageName",
	"RunContainerError", "CreateContainerError",
	"PreStartHookError", "PostStartHookError",
}

// PodNames returns the pods' names, in order.
//
// Keywords: pods, names, list, deployment, statefulset, string
//
// Example:
//
//	{{ podNames .pods }}  → [web-0 web-1 web-2]
//	{{ join (podNames .pods) ", " }}  → "web-0, web-1, web-2"
func PodNames(pods interface{}) []string {
	return podStrings(pods, func(p map[string]interface{}) string { return nestedString(p, "metadata", "name") })
}

// PodIPs returns the pods' IP addresses; pods without one yet are skipped.
//
// Keywords: pods, ip, list, address, network, string
//
// Example:
//
//	{{ podIPs .pods }}  → [10.0.0.1 10.0.0.2]
func PodIPs(pods interface{}) []string {
	return podStrings(pods, func(p map[string]interface{}) string { return nestedString(p, "status", "podIP") })
}

// PodPhases returns each pod's phase, in order.
//
// Keywords: pods, phase, list, status, lifecycle, string
//
// Example:
//
//	{{ podPhases .pods }}  → [Running Running Pending]
func PodPhases(pods interface{}) []string {
	return podStrings(pods, func(p map[string]interface{}) string { return nestedString(p, "status", "phase") })
}

// PodNodes returns the nodes the pods are scheduled on; pending pods are
// skipped.
//
// Keywords: pods, node, list, scheduled, placement, string
//
// Example:
//
//	{{ podNodes .pods }}  → [node-1 node-2]
func PodNodes(pods interface{}) []string {
	return podStrings(pods, func(p map[string]interface{}) string { return nestedString(p, "spec", "nodeName") })
}

// PodCount returns the number of pods.
//
// Keywords: pods, count, int, total, size
//
// Example:
//
//	{{ podCount .pods }}  → 3
func PodCount(pods interface{}) int {
	return len(podList(pods))
}

// ReadyPodCount returns the number of pods whose Ready condition is True.
//
// Keywords: pods, ready, count, int, health
//
// Example:
//
//	{{ readyPodCount .pods }}  → 2
func ReadyPodCount(pods interface{}) int {
	n := 0
	for _, p := range podList(pods) {
		if podReady(p) {
			n++
		}
	}
	return n
}

// PodMaxRestarts returns the highest restart count of any pod (the sum over
// its containers). 0 when there are no pods.
//
// Keywords: pods, restart, count, crash, int, max
//
// Example:
//
//	{{ podMaxRestarts .pods }}  → 1
func PodMaxRestarts(pods interface{}) int64 {
	var max int64
	for _, p := range podList(pods) {
		if r := podRestarts(p); r > max {
			max = r
		}
	}
	return max
}

// HasCrashingPod reports whether any pod is unhealthy: restarted more than
// twice, or a container waiting or terminated with a failure reason
// (CrashLoopBackOff, OOMKilled, ImagePullBackOff, …).
//
// Keywords: pods, crash, restart, boolean, loop, health
//
// Example:
//
//	{{ hasCrashingPod .pods }}  → true
func HasCrashingPod(pods interface{}) bool {
	list := podList(pods)
	for _, reason := range crashReasons {
		if hasContainerReason(list, reason) {
			return true
		}
	}
	for _, p := range list {
		if podRestarts(p) > 2 {
			return true
		}
	}
	return false
}

// PodByOrdinal returns the pod whose name ends in -<ordinal>, as StatefulSet
// pods are named (web-0, web-1, …); nil when none does.
//
// Keywords: pods, statefulset, ordinal, index, primary, member
//
// Example:
//
//	{{ (podByOrdinal .pods 0).status.podIP }}  → "10.0.0.1"
func PodByOrdinal(pods interface{}, ordinal int) map[string]interface{} {
	for _, p := range podList(pods) {
		if podOrdinal(nestedString(p, "metadata", "name")) == ordinal {
			return p
		}
	}
	return nil
}

// PodContainerReasons returns the distinct reasons containers are waiting or
// terminated with, across all pods, in first-seen order.
//
// Keywords: pods, container, reasons, error, list, string, status
//
// Example:
//
//	{{ podContainerReasons .pods }}  → [ImagePullBackOff]
func PodContainerReasons(pods interface{}) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range podList(pods) {
		for _, cs := range containerStatuses(p) {
			if _, r := containerStateReason(cs); r != "" && !seen[r] {
				seen[r] = true
				out = append(out, r)
			}
		}
	}
	return out
}

// PodContainerState returns the state (Running, Waiting, Terminated) of a
// named container in the pod at ordinal; "" when either is not found.
//
// Keywords: pods, container, state, statefulset, ordinal, string, running, waiting
//
// Example:
//
//	{{ podContainerState .pods 0 "app" }}  → "Running"
func PodContainerState(pods interface{}, ordinal int, container string) string {
	pod := PodByOrdinal(pods, ordinal)
	for _, cs := range containerStatuses(pod) {
		if n, _ := cs["name"].(string); n == container {
			s, _ := containerStateReason(cs)
			return s
		}
	}
	return ""
}

// PodCrashLoopDetected reports whether any container in the pods is waiting or terminated with reason CrashLoopBackOff.
//
// Keywords: pods, container, crashloopbackoff, failure, health
//
// Example:
//
//	{{ podCrashLoopDetected .pods }}  → false
func PodCrashLoopDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "CrashLoopBackOff")
}

// PodImagePullBackOffDetected reports whether any container in the pods is waiting or terminated with reason ImagePullBackOff.
//
// Keywords: pods, container, imagepullbackoff, failure, health
//
// Example:
//
//	{{ podImagePullBackOffDetected .pods }}  → true
func PodImagePullBackOffDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "ImagePullBackOff")
}

// PodErrImagePullDetected reports whether any container in the pods is waiting or terminated with reason ErrImagePull.
//
// Keywords: pods, container, errimagepull, failure, health
//
// Example:
//
//	{{ podErrImagePullDetected .pods }}  → false
func PodErrImagePullDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "ErrImagePull")
}

// PodErrorDetected reports whether any container in the pods is waiting or terminated with reason Error.
//
// Keywords: pods, container, error, failure, health
//
// Example:
//
//	{{ podErrorDetected .pods }}  → false
func PodErrorDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "Error")
}

// PodOOMKilledDetected reports whether any container in the pods is waiting or terminated with reason OOMKilled.
//
// Keywords: pods, container, oomkilled, failure, health
//
// Example:
//
//	{{ podOOMKilledDetected .pods }}  → false
func PodOOMKilledDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "OOMKilled")
}

// PodRunContainerErrorDetected reports whether any container in the pods is waiting or terminated with reason RunContainerError.
//
// Keywords: pods, container, runcontainererror, failure, health
//
// Example:
//
//	{{ podRunContainerErrorDetected .pods }}  → false
func PodRunContainerErrorDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "RunContainerError")
}

// PodCreateContainerErrorDetected reports whether any container in the pods is waiting or terminated with reason CreateContainerError.
//
// Keywords: pods, container, createcontainererror, failure, health
//
// Example:
//
//	{{ podCreateContainerErrorDetected .pods }}  → false
func PodCreateContainerErrorDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "CreateContainerError")
}

// PodInvalidImageNameDetected reports whether any container in the pods is waiting or terminated with reason InvalidImageName.
//
// Keywords: pods, container, invalidimagename, failure, health
//
// Example:
//
//	{{ podInvalidImageNameDetected .pods }}  → false
func PodInvalidImageNameDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "InvalidImageName")
}

// PodPreStartHookErrorDetected reports whether any container in the pods is waiting or terminated with reason PreStartHookError.
//
// Keywords: pods, container, prestarthookerror, failure, health
//
// Example:
//
//	{{ podPreStartHookErrorDetected .pods }}  → false
func PodPreStartHookErrorDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "PreStartHookError")
}

// PodPostStartHookErrorDetected reports whether any container in the pods is waiting or terminated with reason PostStartHookError.
//
// Keywords: pods, container, poststarthookerror, failure, health
//
// Example:
//
//	{{ podPostStartHookErrorDetected .pods }}  → false
func PodPostStartHookErrorDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "PostStartHookError")
}

// ── native Pod helpers ────────────────────────────────────────────────────

// podList returns the pods in v: a list of Pod objects, or a PodList (an
// object with items). Anything else is no pods.
func podList(v interface{}) []map[string]interface{} {
	return objectList(v)
}

// podStrings collects f(pod) for each pod, skipping empty values.
func podStrings(pods interface{}, f func(map[string]interface{}) string) []string {
	var out []string
	for _, p := range podList(pods) {
		if s := f(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func podReady(pod map[string]interface{}) bool {
	conds, _ := nested(pod, "status", "conditions").([]interface{})
	for _, c := range conds {
		cm, _ := c.(map[string]interface{})
		if t, _ := cm["type"].(string); t == "Ready" {
			s, _ := cm["status"].(string)
			return s == "True"
		}
	}
	return false
}

func podRestarts(pod map[string]interface{}) int64 {
	var total int64
	for _, cs := range containerStatuses(pod) {
		total += toInt64(cs["restartCount"])
	}
	return total
}

// podOrdinal is the number after the last dash in a pod name, -1 when there
// is none.
func podOrdinal(name string) int {
	i := strings.LastIndex(name, "-")
	if i < 0 {
		return -1
	}
	n, err := strconv.Atoi(name[i+1:])
	if err != nil {
		return -1
	}
	return n
}

func containerStatuses(pod map[string]interface{}) []map[string]interface{} {
	list, _ := nested(pod, "status", "containerStatuses").([]interface{})
	out := make([]map[string]interface{}, 0, len(list))
	for _, c := range list {
		if cm, ok := c.(map[string]interface{}); ok {
			out = append(out, cm)
		}
	}
	return out
}

// containerStateReason returns a container status's current state (Running,
// Waiting, Terminated) and, when waiting or terminated, its reason.
func containerStateReason(cs map[string]interface{}) (state, reason string) {
	st, _ := cs["state"].(map[string]interface{})
	if _, ok := st["running"]; ok {
		return "Running", ""
	}
	if w, ok := st["waiting"].(map[string]interface{}); ok {
		r, _ := w["reason"].(string)
		return "Waiting", r
	}
	if t, ok := st["terminated"].(map[string]interface{}); ok {
		r, _ := t["reason"].(string)
		return "Terminated", r
	}
	return "", ""
}

func hasContainerReason(pods []map[string]interface{}, reason string) bool {
	for _, p := range pods {
		for _, cs := range containerStatuses(p) {
			if _, r := containerStateReason(cs); r == reason {
				return true
			}
		}
	}
	return false
}
