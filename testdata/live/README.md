# Live snapshots

Objects captured from a real cluster, each a workload and its pods in a state
that is easy to read wrong. `kube/live_test.go` checks KEL's answers on every
one, as maps (the shape a dynamic client returns) and as typed client-go
objects.

Each file holds `kubernetes` (the server version), `workload` and `pods` (a
PodList), as `kubectl get -o json` returns them with `managedFields` removed.

| Snapshot | State |
|---|---|
| `healthy`, `statefulset` | Everything ready |
| `bad-image` | Image tag doesn't exist: `ImagePullBackOff` |
| `crashloop-backoff` | Crash looping, waiting: `CrashLoopBackOff`, last `Error` |
| `crashloop-between-restarts` | Crash looping, between back-offs: `terminated/Error`, no `CrashLoopBackOff` anywhere |
| `oom-terminated` | Just OOM killed: `terminated/OOMKilled` |
| `oom-backoff` | OOM killed, waiting: `CrashLoopBackOff`, `OOMKilled` only in `lastState` |
| `init-crashloop` | Init container crash looping (`Init:CrashLoopBackOff`); `containerStatuses` says only `PodInitializing` |
| `init-error` | Init container between restarts (`Init:Error`) |
| `recovered` | Crashed 3 times, now running and ready |
| `never-ready` | Running, readiness probe never passes, progress deadline exceeded |
| `stuck-rollout` | New pods can't pull; old pods serve, so `Available` stays `True` |
| `job-failed`, `job-succeeded` | A Job past its backoff limit, and one that finished |

Captured on Kubernetes v1.37.0 (kind).

## Capturing again

On a cluster you can throw away, with `kubectl` pointing at it:

```bash
kubectl apply -f hack/live/workloads.yaml
kubectl -n kel-lab rollout status deploy/stuck-rollout
kubectl -n kel-lab set image deploy/stuck-rollout web=nginx:no-such-tag
```

- Wait about 5 minutes: progress deadlines pass and "recovered" recovers
```bash
./hack/live/capture.sh
go test ./kube -run TestLive
```

The script waits for each state, so it takes a few minutes. It needs `python3`.
To keep snapshots from another Kubernetes version, capture into a new directory
and add it to the test.
