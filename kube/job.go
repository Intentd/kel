package kube

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/robfig/cron/v3"
)

// jobFuncs — Job
//
// Domain: job
//
// Read Jobs and CronJobs: whether a Job succeeded, failed or is running, and
// when a CronJob last ran, last succeeded and runs next. The functions that
// look inside a Job's pods take the pods as a list.
func jobFuncs() template.FuncMap {
	return template.FuncMap{
		"jobSucceeded":         JobSucceeded,
		"jobFailed":            JobFailed,
		"jobActive":            JobActive,
		"jobFirstExitCode":     JobFirstExitCode,
		"jobActivePodNames":    JobActivePodNames,
		"jobSucceededPodNames": JobSucceededPodNames,
		"jobFailedPodNames":    JobFailedPodNames,
		// CronJobs.
		"cronJobActiveCount":      CronJobActiveCount,
		"cronJobLastScheduleTime": CronJobLastScheduleTime,
		"cronJobNextScheduleTime": CronJobNextScheduleTime,
		"cronJobLastSuccessTime":  CronJobLastSuccessTime,
	}
}

// JobSucceeded returns `true` when `status.succeeded > 0` — at least one pod
// has completed successfully.
//
// Keywords: job, batch, succeeded, complete, boolean, migration, init
//
// Example:
//
//	{{ jobSucceeded .job }}  → true
func JobSucceeded(obj interface{}) bool {
	status := Status(obj)
	return toInt64(status["succeeded"]) > 0
}

// JobFailed returns `true` when `status.failed > 0` — at least one pod has
// failed (and not been retried to success).
//
// Keywords: job, batch, failed, error, boolean, retry, fault
//
// Example:
//
//	{{ jobFailed .job }}  → false
func JobFailed(obj interface{}) bool {
	status := Status(obj)
	return toInt64(status["failed"]) > 0
}

// JobActive returns `true` when `status.active > 0` — at least one pod is
// currently running.
//
// Keywords: job, batch, active, running, boolean, in-progress
//
// Example:
//
//	{{ jobActive .job }}  → false
func JobActive(obj interface{}) bool {
	status := Status(obj)
	return toInt64(status["active"]) > 0
}

// ── Job pods ──────────────────────────────────────────────────────────────

// JobFirstExitCode returns the exit code of the first terminated container
// among the job's pods; -1 when none has terminated yet.
//
// Keywords: job, batch, exit, code, pod, terminated
//
// Example:
//
//	{{ jobFirstExitCode .jobPods }}  → 1
func JobFirstExitCode(pods interface{}) int64 {
	for _, p := range podList(pods) {
		for _, cs := range containerStatuses(p) {
			if t, ok := nested(cs, "state", "terminated").(map[string]interface{}); ok {
				return toInt64(t["exitCode"])
			}
		}
	}
	return -1
}

// JobActivePodNames returns the names of the job's pods that are not done
// (neither Succeeded nor Failed).
//
// Keywords: job, batch, active, pods, names, running
//
// Example:
//
//	{{ jobActivePodNames .jobPods }}  → [migrate-r2d5m]
func JobActivePodNames(pods interface{}) []string {
	return podNamesInPhase(pods, func(phase string) bool { return phase != "Succeeded" && phase != "Failed" })
}

// JobSucceededPodNames returns the names of the job's pods that succeeded.
//
// Keywords: job, batch, succeeded, pods, names, complete
//
// Example:
//
//	{{ jobSucceededPodNames .jobPods }}  → [migrate-q9w4z]
func JobSucceededPodNames(pods interface{}) []string {
	return podNamesInPhase(pods, func(phase string) bool { return phase == "Succeeded" })
}

// JobFailedPodNames returns the names of the job's pods that failed.
//
// Keywords: job, batch, failed, pods, names, error
//
// Example:
//
//	{{ jobFailedPodNames .jobPods }}  → [migrate-x7k2p]
func JobFailedPodNames(pods interface{}) []string {
	return podNamesInPhase(pods, func(phase string) bool { return phase == "Failed" })
}

func podNamesInPhase(pods interface{}, match func(string) bool) []string {
	var names []string
	for _, p := range podList(pods) {
		if match(nestedString(p, "status", "phase")) {
			if n := nestedString(p, "metadata", "name"); n != "" {
				names = append(names, n)
			}
		}
	}
	return names
}

// CronJobActiveCount returns the number of currently active Job runs (length
// of `status.active`).
//
// Keywords: cronjob, cron, active, count, int, scheduled
//
// Example:
//
//	{{ cronJobActiveCount .cronjob }}  → 1
func CronJobActiveCount(obj interface{}) int {
	status := Status(obj)
	active, _ := status["active"].([]interface{})
	return len(active)
}

// CronJobLastScheduleTime returns the last time the CronJob was scheduled
// (`status.lastScheduleTime`). Returns `""` when not yet scheduled.
//
// Keywords: cronjob, cron, schedule, time, last, timestamp
//
// Example:
//
//	{{ cronJobLastScheduleTime .cronjob }}  → "2026-05-19T10:00:00Z"
func CronJobLastScheduleTime(obj interface{}) string {
	status := Status(obj)
	v, _ := status["lastScheduleTime"].(string)
	return v
}

// CronJobNextScheduleTime returns when the CronJob runs next, computed from
// its spec the way the CronJob controller schedules it: spec.schedule in
// spec.timeZone (the controller's zone, UTC here, when unset). "" when the
// CronJob is suspended or has no schedule; an error when the schedule or time
// zone does not parse.
//
// Keywords: cronjob, cron, schedule, time, next, timestamp, timezone, suspend
//
// Example:
//
//	{{ cronJobNextScheduleTime .cronjob }}  → "2026-05-19T11:00:00Z"
//
// See also: nextCron, cronJobLastScheduleTime
func CronJobNextScheduleTime(cronJob interface{}) (string, error) {
	return cronJobNext(cronJob, time.Now())
}

// cronJobNext is CronJobNextScheduleTime at the given time.
func cronJobNext(cronJob interface{}, now time.Time) (string, error) {
	cj := object(cronJob)
	if suspend, _ := nested(cj, "spec", "suspend").(bool); suspend {
		return "", nil
	}
	expr := strings.TrimSpace(nestedString(cj, "spec", "schedule"))
	if expr == "" {
		return "", nil
	}
	if tz := nestedString(cj, "spec", "timeZone"); tz != "" && !strings.HasPrefix(expr, "CRON_TZ=") && !strings.HasPrefix(expr, "TZ=") {
		expr = "CRON_TZ=" + tz + " " + expr
	}
	schedule, err := cron.ParseStandard(expr)
	if err != nil {
		return "", fmt.Errorf("cronjob schedule %q: %w", expr, err)
	}
	return schedule.Next(now).UTC().Format(time.RFC3339), nil
}

// CronJobLastSuccessTime returns the last time the CronJob completed
// successfully (`status.lastSuccessfulTime`). Returns `""` when it has never
// succeeded.
//
// Keywords: cronjob, cron, success, time, last, timestamp
//
// Example:
//
//	{{ cronJobLastSuccessTime .cronjob }}  → "2026-05-19T10:00:00Z"
func CronJobLastSuccessTime(obj interface{}) string {
	status := Status(obj)
	v, _ := status["lastSuccessfulTime"].(string)
	return v
}
