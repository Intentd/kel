<p align="center">
  <img src="./assets/kel.svg" alt="KEL, the Kubernetes Expression Library" width="500">
</p>

Functions that know Kubernetes, for Go templates and any engine that calls Go functions.

```yaml
{{ if hasCrashingPod .pods }}
crashing: {{ join (podContainerReasons .pods) ", " }}
{{ end }}
```

KEL is a library, not a language. There is no parser, grammar or runtime to learn: Go's `text/template` already composes functions with pipes. KEL adds functions that understand Kubernetes objects (Deployments, Pods, conditions, Events, …), and general helpers for the values around them: strings, numbers, time, cron, semver, Prometheus results and more.

## Why

Every team that runs Kubernetes ends up writing the same functions. `hasCrashingPod`. `rolloutComplete`. `allReplicasReady`. `isTerminating`. The names change a little and the logic doesn't. Each copy lives in one controller, and when that project ends the knowledge goes with it.

Helm showed that Go templates work well for writing Kubernetes resources: values in, manifests out, before anything reaches the cluster. KEL brings the same templates to the other side, where objects are live and the question is about their state: has the rollout finished, is a pod crashing, why isn't it Ready.

CEL already runs at that point, in admission policies, CRD validation rules, and tools like kro that read other resources' status. Its Kubernetes extensions cover quantities, IPs and URLs, but not what the fields mean: to ask whether a rollout is complete, you compare `observedGeneration`, `updatedReplicas` and `availableReplicas` yourself, in every rule that needs it. KEL is where that logic is written once.

The Kubernetes API, meanwhile, hardly moves. A Deployment's status has the same shape in every cluster, and so do conditions, Pods and Events. So the reasoning about them can be written once, tested once and used everywhere: in a controller, an admission check, a test assertion or a CLI. KEL is that reasoning, written once.

## Install

```bash
go get github.com/intentd/kel
```

## Use it in a template

A complete program is in [examples/main.go](examples/main.go): a Deployment mid-rollout with a crash-looping pod, asked about from a template and from Go. Run it with `go run ./examples`.

`kel.FuncMap()` returns every function as a `template.FuncMap`:

```go
data := map[string]any{
	"deployment": deployment, // as the API returns it, e.g. unstructured.Unstructured.Object
	"pods":       pods,       // a list of Pods, or a PodList
}

tmpl := template.Must(template.New("status").Funcs(kel.FuncMap()).Parse(
	`{{ if rolloutComplete .deployment }}rolled out{{ else }}rolling out{{ end }}`,
))
err := tmpl.Execute(os.Stdout, data)
```

Functions read objects in the shape the Kubernetes API returns them: the `map[string]any` a dynamic client or `unstructured.Unstructured` gives you, or the same decoded from JSON or YAML. They do no I/O. A function about pods takes the pods as an argument, so you decide what to fetch and pass it in.

## Use it from Go

Every function is also an exported Go function, so the same logic works without templates:

```go
import "github.com/intentd/kel/kube"

if !kube.RolloutComplete(deployment) && kube.HasCrashingPod(pods) {
	log.Printf("rollout stuck: %v", kube.PodContainerReasons(pods))
}
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
| `{{ conditionReason .obj "Ready" }}` | Why is the Ready condition in its current state? |
| `{{ podOOMKilledDetected .pods }}` | Was any container killed for running out of memory? |
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

KEL is before 1.0. Function names and signatures may still change between minor versions; release notes list every breaking change. It started inside [Inrun](https://github.com/intentd/inrun), which uses it in every Catalog expression, and is now a separate module with no dependency on Inrun.

## Contributing

A good KEL function:

- answers a question that comes up across projects, not one project's own question. Functions about a widely used tool's output, such as a Prometheus query result, belong in `ecosystem`
- takes its input in the shape its source produces: Kubernetes objects as the API returns them, a Prometheus result as the HTTP API returns it
- returns zero values for missing data, and an error only when it cannot answer
- has a doc comment in the standard shape, with keywords and an example, and tests

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

Then run `go generate ./...` to update the registry. `go test ./...` fails if the registry is out of date, if a registered function has no documentation, or if a signature breaks the rules above.

## License

[Apache 2.0](LICENSE)
