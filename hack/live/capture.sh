#!/bin/sh
# Capture every snapshot in testdata/live from the kel-lab namespace.
# Apply workloads.yaml first and wait a few minutes; see testdata/live/README.md.
set -e
cd "$(dirname "$0")/../.."
out=${OUT:-testdata/live}
c() { python3 -I hack/live/capture.py "$out" "$@"; }

# States that last: capture as they are.
c healthy deploy/healthy app=healthy
c statefulset sts/db app=db
c bad-image deploy/bad-image app=bad-image
c never-ready deploy/never-ready app=never-ready
c stuck-rollout deploy/stuck-rollout app=stuck-rollout
c job-failed job/job-fail job-name=job-fail
c job-succeeded job/job-ok job-name=job-ok
c recovered deploy/recovered app=recovered '"running" in cs["state"] and cs["restartCount"]>=3'

# States that come and go: wait for each.
c crashloop-backoff deploy/crashloop app=crashloop '"waiting" in cs.get("state",{}) and cs.get("lastState",{}).get("terminated")'
c init-crashloop deploy/init-crash app=init-crash '"waiting" in ics.get("state",{}) and ics["restartCount"]>0'
c oom-terminated deploy/oom app=oom '"terminated" in cs.get("state",{}) and cs["state"]["terminated"]["reason"]=="OOMKilled" and cs["restartCount"]>0'

# The short-lived states come soonest on a new pod, while back-offs are short.
kubectl -n kel-lab delete pod -l 'app in (crashloop,oom,init-crash)' --wait=true
c crashloop-between-restarts deploy/crashloop app=crashloop '"terminated" in cs.get("state",{}) and cs["restartCount"]>0' &
c oom-backoff deploy/oom app=oom '"waiting" in cs.get("state",{}) and cs.get("lastState",{}).get("terminated",{}).get("reason")=="OOMKilled"' &
c init-error deploy/init-crash app=init-crash '"terminated" in ics.get("state",{}) and ics["restartCount"]>0' &
wait
