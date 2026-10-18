<div align="center">
  <img src="./assets/kel.svg" alt="KEL, the Kubernetes Expression Library" height="64" />

  <p><strong>Kubernetes knowledge, written once.</strong></p>

  <p>
    <a href="https://github.com/intentd/kel/actions/workflows/ci.yml"><img src="https://github.com/intentd/kel/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
    <a href="https://github.com/intentd/kel/tags"><img src="https://img.shields.io/github/v/tag/intentd/kel?label=version" alt="Version" /></a>
    <a href="https://pkg.go.dev/github.com/intentd/kel"><img src="https://pkg.go.dev/badge/github.com/intentd/kel.svg" alt="Go Reference" /></a>
    <img src="https://img.shields.io/badge/Go-1.27+-00ADD8.svg" alt="Go 1.27+" />
    <img src="https://img.shields.io/badge/Kubernetes-1.37-326CE5.svg" alt="Tested on Kubernetes 1.37" />
    <img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License" />
  </p>
</div>

Functions that know Kubernetes, for Go templates and any engine that calls Go functions.

```yaml
{{ if hasCrashingPod .pods }}
crashing: {{ join (podContainerReasons .pods) ", " }}
{{ end }}
```

KEL is a library, not a language. There is no parser, grammar or runtime to learn: Go's `text/template` already composes functions with pipes. KEL adds functions that understand Kubernetes objects (Deployments, Pods, conditions, Events, …), and general helpers for the values around them: strings, numbers, time, cron, semver, Prometheus results and more.

## Why

Every team that runs Kubernetes ends up writing the same functions. `hasCrashingPod`. `rolloutComplete`. `allReplicasReady`. `isTerminating`. The names change a little and the logic doesn't. Each copy lives in one controller, and when that project ends the knowledge goes with it.

The shape of the API hardly moves: a Deployment's status looks the same in every cluster. What takes knowing is how to read it. A crash-looping container doesn't always say `CrashLoopBackOff`. A Deployment can report `Available` while its rollout is stuck. A pod that restarted last week still shows the restarts today. Each team learns these one incident at a time, and writes the lesson into its own copy.

KEL collects that knowledge in one place: each lesson written once, as a function, with a test that keeps it true. Use it wherever the question comes up: a controller, a CLI, a test assertion, an alert template.

## What KEL knows

A few of the things KEL's functions account for, so their callers don't have to:

- **Crash loops don't always say so.** Between back-offs a crash-looping container shows `terminated` with reason `Error`; only while it waits does it show `CrashLoopBackOff`. `podCrashLoopDetected` sees both.
- **The reason moves.** Once an OOM-killed container is waiting to restart, its state says `CrashLoopBackOff` and `OOMKilled` is in its last termination. `podOOMKilledDetected` reads both, but only while the container isn't running: a past kill on a container that recovered is history.
- **Restart counts never reset** while the pod lives. `hasCrashingPod` asks whether a container is failing now; `podRestartedWithin .pods "10m"` asks whether one restarted recently.
- **Init containers are reported separately**, in `initContainerStatuses`. A pod stuck in `Init:Error` has nothing wrong in its `containerStatuses`. The pod functions read both, and skip init containers that finished.
- **Status can be about an older spec.** Until `status.observedGeneration` reaches `metadata.generation`, the controller hasn't acted on the latest change. `rolloutComplete` and `conditionsHealthy` treat that as not done.
- **`Available` isn't finished.** A Deployment stays `Available` on its old pods while the new ones fail. `rolloutComplete` follows the rules `kubectl rollout status` uses; `conditionsHealthy` also reads `Progressing`.
- **Most operators agree on conditions.** `Ready`, `Stalled`, `Reconciling` and `Synced` mean the same thing across many projects' resources, so `conditionsHealthy` answers for kinds KEL has never seen.
- **Typed objects arrive without their kind.** Objects from a typed client-go list call have an empty `TypeMeta`. KEL takes the kind from the Go type.

## Install

```bash
go get github.com/intentd/kel
```

## Use it in a template

A complete program is in [examples/main.go](examples/main.go): a Deployment mid-rollout with a crash-looping pod, asked about from a template and from Go. Run it with `go run ./examples`.

`kel.FuncMap()` returns every function as a `template.FuncMap`:

```go
data := map[string]any{
	"deployment": deployment, // *appsv1.Deployment, *unstructured.Unstructured or a map
	"pods":       pods,       // []corev1.Pod, *corev1.PodList, a list of maps, …
}

tmpl := template.Must(template.New("status").Funcs(kel.FuncMap()).Parse(
	`{{ if rolloutComplete .deployment }}rolled out{{ else }}rolling out{{ end }}`,
))
err := tmpl.Execute(os.Stdout, data)
```

Pass objects in whatever shape you have them:

- typed client-go objects: `*corev1.Pod`, `[]corev1.Pod`, `*corev1.PodList`
- unstructured objects: `*unstructured.Unstructured`, `[]unstructured.Unstructured`, `*unstructured.UnstructuredList`
- maps, as a dynamic client returns them or as decoded from JSON or YAML

Every function gives the same answer for each. Functions do no I/O: a function about pods takes the pods as an argument, so you decide what to fetch and pass it in. A function that takes a list wants a list; one object is not a list of one.

## Use it from Go

Every function is also an exported Go function, so the same logic works without templates:

```go
import "github.com/intentd/kel/kube"

pods, err := clientset.CoreV1().Pods(ns).List(ctx, opts)
if !kube.RolloutComplete(deployment) && kube.HasCrashingPod(pods) {
	log.Printf("rollout stuck: %v", kube.PodContainerReasons(pods))
}

// Labels and annotations as the map[string]string Go APIs take:
selector.Matches(labels.Set(kube.Labels(pod).StringMap()))
```

## What's inside

The functions live in three packages, grouped by what they read:

| Package | Reads | Domains |
|---|---|---|
| `core` | plain values | strings, math, conditional, types, collections, safe access, data, time, cron, semver, random, net, domain, validation |
| `kube` | Kubernetes objects | kubernetes, fields, replica, pods, warnings, container, job, service, ingress, hpa, pvc, node, statefulset, replicaset, quantity, Kubernetes validation |
| `ecosystem` | what tools around Kubernetes produce | Prometheus query results, Git repository URLs, image references |

A few examples:

| Expression | Answers |
|---|---|
| `{{ rolloutComplete .deployment }}` | Has the rollout finished, by the rules `kubectl rollout status` uses for Deployments, StatefulSets and DaemonSets? |
| `{{ conditionsHealthy .obj }}` | Do the object's standard conditions say it's healthy, for any kind that reports them? |
| `{{ conditionReason .obj "Ready" }}` | Why is the Ready condition in its current state? |
| `{{ hasCrashingPod .pods }}` | Is any container failing right now? |
| `{{ podOOMKilledDetected .pods }}` | Is any container failing because it runs out of memory? |
| `{{ warningReasons .events }}` | Which Warning events were recorded? |
| `{{ timeSince (creationTimestamp .obj) }}` | How many seconds old is the object? |
| `{{ semverConstraint .spec.version "^1.4" }}` | Does the version satisfy the range? |
| `{{ promAboveThreshold .response 1000 }}` | Is the Prometheus query result above 1000? |

## How functions behave

The functions follow the same rules, so you can predict one you haven't used:

- **Missing data gives zero values.** An absent field, an empty list or a nil object returns `""`, `0`, `false` or an empty list, not a panic.
- **Questions return booleans.** `is…`, `has…`, `all…`, `any…` and `…Detected` functions return `bool`.
- **Lists return lists.** `podNames`, `warningReasons`, `ingressRules` and other plural names return `[]string` (or `[]int64` for ports). Use `join` to turn one into text.
- **Bad input is an error.** A function that cannot answer, such as `semverMajor` on `"not-a-version"`, returns an error. In a template the error stops execution and is reported, instead of rendering a wrong value.
- **Answers are about now.** Health functions describe the current state, not the object's history, and don't read the clock. The few that do, such as `podRestartedWithin` and `timeSince`, say so.

## Add your own functions

`Register` adds functions beside KEL's and panics if a name is already taken. `Override` replaces one of KEL's on purpose and panics if the name doesn't exist, so a typo can't add a new function by mistake.

```go
funcs := kel.FuncMap() // a copy; changing it does not affect other callers
kel.Register(funcs, template.FuncMap{"teamOwner": teamOwner})
```

## Find a function

Each function's documentation comes from its Go doc comment, and the registry is generated from those comments. `kel.Builtin()` returns it, which is useful for CLI help, editor completion or your own reference pages:

```go
reg := kel.Builtin()
f, ok := reg.Get("rolloutComplete") // domain, description, example, keywords
matches := reg.Search("crash")      // hasCrashingPod, podCrashLoopDetected, …
pods := reg.ByDomain("pods")
```

If you add functions of your own, document them the same way and merge them with `reg.With(...)`. The `gendoc` package reads doc comments in KEL's format from any Go package.

## Status

KEL is before 1.0. Function names and signatures may still change between minor versions; release notes list every breaking change. It started inside [Inrun](https://github.com/intentd/inrun), which uses it in every **Blueprint** expression, and is now a separate module with no dependency on Inrun.

## Contributing

KEL grows by what people have learned the hard way. If you've debugged a state that a function here reads wrong, or written the same check into a second project, that's a contribution: a fix with a test that shows the case, or a new function.

A good KEL function:

- answers a question that comes up across projects, not one project's own question. Functions about a widely used tool's output, such as a Prometheus query result, belong in `ecosystem`
- takes its input in the shape its source produces: Kubernetes objects as the API returns them, a Prometheus result as the HTTP API returns it
- returns zero values for missing data, and an error only when it cannot answer
- has a doc comment in the standard shape, with keywords and an example, and tests. A test built from a real object (`kubectl get -o json`) is the best evidence for a subtle case: [testdata/live](testdata/live) holds snapshots of fourteen states from a real cluster, and how to capture more

Document a new function like this:

```go
// PodCount returns the number of pods.
//
// Keywords: pods, count, total
//
// Example:
//
//	{{ podCount .pods }}  → 3
func PodCount(pods interface{}) int {
```

Then run `go generate ./...` to update the registry. `go test ./...` fails if the registry is out of date, if a registered function has no documentation, if a signature breaks the rules above, or if a documented example gives a different answer for typed or unstructured objects than for maps.

## License

[Apache 2.0](LICENSE)
