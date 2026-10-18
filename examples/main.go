// Command examples shows KEL on a Deployment in the middle of a rollout:
// from a template, from Go, and through the function registry. Run it from
// the repository root:
//
//	go run ./examples
//
// The objects are in objects.yaml and the template in report.tmpl; edit
// either and run it again.
package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/intentd/kel"
	"github.com/intentd/kel/kube"
)

//go:embed objects.yaml
var objects []byte

//go:embed report.tmpl
var report string

func main() {
	// In a real program the objects come from a dynamic client or an informer.
	var data map[string]any
	if err := yaml.Unmarshal(objects, &data); err != nil {
		log.Fatal(err)
	}

	// 1. From a template: add KEL's functions with Funcs.
	tmpl := template.Must(template.New("report").Funcs(kel.FuncMap()).Parse(report))
	fmt.Println("From a template:")
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		log.Fatal(err)
	}

	// 2. From Go: the same functions, called directly.
	deployment, pods, events := data["deployment"], data["pods"], data["events"]
	fmt.Println("\nFrom Go:")
	fmt.Printf("%s in %s\n",
		kube.Name(deployment),
		kube.Namespace(deployment),
	)

	fmt.Printf("  rollout:   ")
	if kube.RolloutComplete(deployment) {
		fmt.Println("complete")
	} else {
		fmt.Println("in progress")
	}

	fmt.Printf("  ready:     %d/%d pods\n",
		kube.ReadyPodCount(pods),
		kube.PodCount(pods),
	)

	if kube.HasCrashingPod(pods) {
		fmt.Printf(
			"  crashing:  %s (most restarts: %d)\n",
			strings.Join(kube.PodContainerReasons(pods), ", "),
			kube.PodMaxRestarts(pods),
		)
	}

	if kube.HasWarnings(events) {
		fmt.Printf("  warning:   %s\n",
			kube.FirstWarning(events),
		)
	}

	// 3. Look a function up: every function is documented in the registry.
	f, _ := kel.Builtin().Get("rolloutComplete")
	fmt.Printf("\nAbout rolloutComplete (%s):\n  %s\n", f.Domain, f.Example)
}
