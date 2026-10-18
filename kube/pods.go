package kube

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/intentd/kel/core"
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
		"podNames":           PodNames,
		"podIPs":             PodIPs,
		"podPhases":          PodPhases,
		"podNodes":           PodNodes,
		"podCount":           PodCount,
		"readyPodCount":      ReadyPodCount,
		"podMaxRestarts":     PodMaxRestarts,
		"hasCrashingPod":     HasCrashingPod,
		"podRestartedWithin": PodRestartedWithin,
		"podByOrdinal":       PodByOrdinal,
		// Container states: from status.initContainerStatuses and
		// status.containerStatuses of each pod.
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

// HasCrashingPod reports whether any container is failing now: waiting or
// terminated with a failure reason (CrashLoopBackOff, OOMKilled,
// ImagePullBackOff, …), or not running after restarting before. Init
// containers count. A container that is running again is not crashing,
// however often it restarted: restart counts only reset when the pod is
// replaced. For one that runs and crashes in turn, use podRestartedWithin.
//
// Keywords: pods, crash, restart, boolean, loop, health, failing
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
		for _, cs := range containerStatuses(p) {
			if state, _ := containerStateReason(cs); state != "Running" && toInt64(cs["restartCount"]) > 0 {
				return true
			}
		}
	}
	return false
}

// PodRestartedWithin reports whether any container in the pods last
// terminated within window ("10m", "1h", "1d"), going by its last
// termination's finishedAt. It catches a container that runs for a while and
// crashes in turn, which looks healthy whenever it is running. An error when
// window is not a positive duration.
//
// Keywords: pods, restart, recent, flapping, crash, window, time, health
//
// Example:
//
//	{{ podRestartedWithin .pods "10m" }}  → false
func PodRestartedWithin(pods interface{}, window string) (bool, error) {
	return podRestartedWithin(pods, window, time.Now())
}

func podRestartedWithin(pods interface{}, window string, now time.Time) (bool, error) {
	d, err := core.ParseTimeDuration(window)
	if err != nil {
		return false, err
	}
	if d <= 0 {
		return false, fmt.Errorf("podRestartedWithin: window %q is not a positive duration", window)
	}
	for _, p := range podList(pods) {
		for _, cs := range allContainerStatuses(p) {
			finished, err := time.Parse(time.RFC3339, nestedString(cs, "lastState", "terminated", "finishedAt"))
			if err == nil && now.Sub(finished) <= d {
				return true, nil
			}
		}
	}
	return false, nil
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
			for _, r := range containerReasons(cs) {
				if !seen[r] {
					seen[r] = true
					out = append(out, r)
				}
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
	for _, cs := range allContainerStatuses(pod) {
		if n, _ := cs["name"].(string); n == container {
			s, _ := containerStateReason(cs)
			return s
		}
	}
	return ""
}

// PodCrashLoopDetected reports whether any container in the pods is crash looping: waiting with reason
// CrashLoopBackOff, or terminated with a non-zero exit code after restarting before, as it shows between back-offs.
//
// Keywords: pods, container, crashloopbackoff, failure, health
//
// Example:
//
//	{{ podCrashLoopDetected .pods }}  → false
func PodCrashLoopDetected(pods interface{}) bool {
	list := podList(pods)
	if hasContainerReason(list, "CrashLoopBackOff") {
		return true
	}
	// Between back-offs a crash-looping container shows terminated with a
	// non-zero exit code, not CrashLoopBackOff. Having restarted before
	// tells it apart from a single failure (restartPolicy: Never).
	for _, p := range list {
		for _, cs := range containerStatuses(p) {
			if t, ok := nested(cs, "state", "terminated").(map[string]interface{}); ok &&
				toInt64(t["exitCode"]) != 0 && toInt64(cs["restartCount"]) > 0 {
				return true
			}
		}
	}
	return false
}

// PodImagePullBackOffDetected reports whether any container in the pods is waiting or terminated with reason, or last terminated with reason while not running, ImagePullBackOff.
//
// Keywords: pods, container, imagepullbackoff, failure, health
//
// Example:
//
//	{{ podImagePullBackOffDetected .pods }}  → true
func PodImagePullBackOffDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "ImagePullBackOff")
}

// PodErrImagePullDetected reports whether any container in the pods is waiting or terminated with reason, or last terminated with reason while not running, ErrImagePull.
//
// Keywords: pods, container, errimagepull, failure, health
//
// Example:
//
//	{{ podErrImagePullDetected .pods }}  → false
func PodErrImagePullDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "ErrImagePull")
}

// PodErrorDetected reports whether any container in the pods is waiting or terminated with reason, or last terminated with reason while not running, Error.
//
// Keywords: pods, container, error, failure, health
//
// Example:
//
//	{{ podErrorDetected .pods }}  → false
func PodErrorDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "Error")
}

// PodOOMKilledDetected reports whether any container in the pods is waiting or terminated with reason, or last terminated with reason while not running, OOMKilled.
//
// Keywords: pods, container, oomkilled, failure, health
//
// Example:
//
//	{{ podOOMKilledDetected .pods }}  → false
func PodOOMKilledDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "OOMKilled")
}

// PodRunContainerErrorDetected reports whether any container in the pods is waiting or terminated with reason, or last terminated with reason while not running, RunContainerError.
//
// Keywords: pods, container, runcontainererror, failure, health
//
// Example:
//
//	{{ podRunContainerErrorDetected .pods }}  → false
func PodRunContainerErrorDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "RunContainerError")
}

// PodCreateContainerErrorDetected reports whether any container in the pods is waiting or terminated with reason, or last terminated with reason while not running, CreateContainerError.
//
// Keywords: pods, container, createcontainererror, failure, health
//
// Example:
//
//	{{ podCreateContainerErrorDetected .pods }}  → false
func PodCreateContainerErrorDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "CreateContainerError")
}

// PodInvalidImageNameDetected reports whether any container in the pods is waiting or terminated with reason, or last terminated with reason while not running, InvalidImageName.
//
// Keywords: pods, container, invalidimagename, failure, health
//
// Example:
//
//	{{ podInvalidImageNameDetected .pods }}  → false
func PodInvalidImageNameDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "InvalidImageName")
}

// PodPreStartHookErrorDetected reports whether any container in the pods is waiting or terminated with reason, or last terminated with reason while not running, PreStartHookError.
//
// Keywords: pods, container, prestarthookerror, failure, health
//
// Example:
//
//	{{ podPreStartHookErrorDetected .pods }}  → false
func PodPreStartHookErrorDetected(pods interface{}) bool {
	return hasContainerReason(podList(pods), "PreStartHookError")
}

// PodPostStartHookErrorDetected reports whether any container in the pods is waiting or terminated with reason, or last terminated with reason while not running, PostStartHookError.
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

// containerStatuses returns the statuses that say how a pod is doing: its init
// containers (including native sidecars) that have not finished successfully,
// then its containers. An init container that exited 0 has done its job, so
// it adds no reasons, restarts or exit codes.
func containerStatuses(pod map[string]interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	for _, cs := range statusList(pod, "initContainerStatuses") {
		if t, ok := nested(cs, "state", "terminated").(map[string]interface{}); ok && toInt64(t["exitCode"]) == 0 {
			continue
		}
		out = append(out, cs)
	}
	return append(out, statusList(pod, "containerStatuses")...)
}

// allContainerStatuses returns every init container and container status,
// for looking a container up by name.
func allContainerStatuses(pod map[string]interface{}) []map[string]interface{} {
	return append(statusList(pod, "initContainerStatuses"), statusList(pod, "containerStatuses")...)
}

func statusList(pod map[string]interface{}, field string) []map[string]interface{} {
	list, _ := nested(pod, "status", field).([]interface{})
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
			for _, r := range containerReasons(cs) {
				if r == reason {
					return true
				}
			}
		}
	}
	return false
}

// containerReasons returns why a container is not running: its current
// state's reason and, while it is not running, the reason it last terminated.
// An OOMKilled container in back-off shows CrashLoopBackOff now and OOMKilled
// last. A running container's past terminations are history and are left out.
func containerReasons(cs map[string]interface{}) []string {
	state, reason := containerStateReason(cs)
	if state == "Running" {
		return nil
	}
	var out []string
	if reason != "" {
		out = append(out, reason)
	}
	if last := nestedString(cs, "lastState", "terminated", "reason"); last != "" && last != reason {
		out = append(out, last)
	}
	return out
}
