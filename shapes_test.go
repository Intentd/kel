package kel_test

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/intentd/kel"
	"github.com/intentd/kel/internal/testutil"
)

// typedFixtures maps each Kubernetes object in testdata/examples.yaml to the
// client-go type it is in Go programs. Custom resources (helmRelease,
// certificate) have no typed form and stay maps, as they do in real code.
var typedFixtures = map[string]any{
	"deployment": appsv1.Deployment{}, "statefulset": appsv1.StatefulSet{},
	"replicaset": appsv1.ReplicaSet{}, "replicaSets": appsv1.ReplicaSet{},
	"pod": corev1.Pod{}, "pods": corev1.Pod{}, "jobPods": corev1.Pod{},
	"events": corev1.Event{}, "node": corev1.Node{}, "service": corev1.Service{},
	"pvc": corev1.PersistentVolumeClaim{}, "pv": corev1.PersistentVolume{},
	"secret": corev1.Secret{}, "job": batchv1.Job{}, "cronjob": batchv1.CronJob{},
	"hpa": autoscalingv2.HorizontalPodAutoscaler{}, "ingress": networkingv1.Ingress{},
	"endpointSlices": discoveryv1.EndpointSlice{}, "storageClass": storagev1.StorageClass{},
}

// TestExamplesInEveryShape runs every documented example again with the
// fixture objects passed the other ways Go programs hold them, and checks
// each renders exactly what it renders for plain maps:
//
//   - unstructured: *unstructured.Unstructured, lists as []unstructured.Unstructured
//   - typed: *corev1.Pod and the like, lists as []corev1.Pod, with TypeMeta
//     empty as objects from a typed list call are
func TestExamplesInEveryShape(t *testing.T) {
	raw, err := os.ReadFile("testdata/examples.yaml")
	if err != nil {
		t.Fatal(err)
	}
	load := func() map[string]any {
		var data map[string]any
		if err := yaml.Unmarshal(raw, &data); err != nil {
			t.Fatal(err)
		}
		return data
	}
	maps := load()

	for _, shape := range []string{"unstructured", "typed"} {
		t.Run(shape, func(t *testing.T) {
			data, converted := load(), map[string]bool{}
			for key, v := range data {
				if c, ok := reshape(t, shape, key, v); ok {
					data[key], converted[key] = c, true
				}
			}
			// Field access into a converted object (.pod.status) is template
			// syntax for maps, not something KEL reads; skip those examples.
			var keys []string
			for k := range converted {
				keys = append(keys, regexp.QuoteMeta(k))
			}
			fieldAccess := regexp.MustCompile(`\.(` + strings.Join(keys, "|") + `)\.[a-zA-Z]`)

			ran, skipped := 0, 0
			for _, f := range kel.Builtin().Functions {
				if varies[f.Name] {
					continue
				}
				for _, line := range strings.Split(f.Example, "\n") {
					if !strings.Contains(line, "{{") {
						continue
					}
					src := strings.TrimSpace(line)
					if m := exampleResult.FindStringSubmatch(line); m != nil {
						src = strings.TrimSpace(m[1])
					}
					if fieldAccess.MatchString(src) {
						skipped++
						continue
					}
					want, wantErr := render(src, maps)
					got, gotErr := render(src, data)
					if (wantErr != nil) != (gotErr != nil) || got != want {
						t.Errorf("%s: %s\n  maps:  %q %v\n  %s: %q %v", f.Name, src, want, wantErr, shape, got, gotErr)
					}
					ran++
				}
			}
			t.Logf("%d examples compared, %d skipped for field access", ran, skipped)
			if ran < 200 {
				t.Errorf("only %d examples ran; the skip rule is too broad", ran)
			}
		})
	}
}

// reshape converts a fixture into shape; false when it stays as it is.
func reshape(t *testing.T, shape, key string, v any) (any, bool) {
	switch shape {
	case "unstructured":
		switch o := v.(type) {
		case map[string]any:
			if _, isObject := o["metadata"]; isObject && key != "metadata" {
				return &unstructured.Unstructured{Object: o}, true
			}
		case []any:
			var items []unstructured.Unstructured
			for _, it := range o {
				m, ok := it.(map[string]any)
				if !ok {
					return nil, false
				}
				items = append(items, unstructured.Unstructured{Object: m})
			}
			return items, true
		}
	case "typed":
		proto, ok := typedFixtures[key]
		if !ok {
			return nil, false
		}
		if list, ok := v.([]any); ok {
			return testutil.TypedList(t, list, proto), true
		}
		return testutil.Typed(t, v, proto), true
	}
	return nil, false
}

func render(src string, data any) (string, error) {
	tmpl, err := template.New("x").Funcs(kel.FuncMap()).Parse(src)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return "", fmt.Errorf("%v", err)
	}
	return b.String(), nil
}
