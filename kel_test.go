package kel_test

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/intentd/kel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpressions(t *testing.T) {
	deployment := map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]any{
			"name":      "my-app",
			"namespace": "production",
			"uid":       "4b3f8d21-8e3a-4f8c-b9d2-1a2b3c4d5e6f",
			"labels": map[string]any{
				"app":                    "frontend",
				"app.kubernetes.io/name": "my-app",
			},
			"annotations": map[string]any{
				"example.com/phase":      "Ready",
				"autoscale/min-replicas": "2",
			},
		},
		"spec": map[string]any{
			"replicas": int64(3),
			"template": map[string]any{
				"spec": map[string]any{
					"containers": []any{
						map[string]any{
							"name":  "app",
							"image": "nginx:1.25",
							"ports": []any{
								map[string]any{"containerPort": int64(8080)},
							},
							"env": []any{
								map[string]any{"name": "APP_ENV", "value": "production"},
							},
						},
					},
				},
			},
		},
		"status": map[string]any{
			"replicas":           int64(3),
			"readyReplicas":      int64(3),
			"availableReplicas":  int64(3),
			"updatedReplicas":    int64(3),
			"observedGeneration": int64(3),
			"conditions": []any{
				map[string]any{
					"type":    "Available",
					"status":  "True",
					"reason":  "MinimumReplicasAvailable",
					"message": "Deployment has minimum availability.",
				},
				map[string]any{
					"type":    "Progressing",
					"status":  "True",
					"reason":  "NewReplicaSetAvailable",
					"message": "ReplicaSet has successfully progressed.",
				},
			},
		},
	}

	pod := map[string]any{
		"kind": "Pod",
		"metadata": map[string]any{
			"name":      "web-abc",
			"namespace": "production",
			"labels": map[string]any{
				"app":                    "frontend",
				"app.kubernetes.io/name": "my-app",
			},
		},
		"status": map[string]any{
			"phase": "Running",
			"podIP": "10.0.0.1",
		},
	}

	service := map[string]any{
		"kind": "Service",
		"metadata": map[string]any{
			"name":      "my-app-svc",
			"namespace": "production",
		},
		"spec": map[string]any{
			"clusterIP": "10.96.0.42",
			"ports": []any{
				map[string]any{"port": int64(80), "nodePort": int64(31234)},
			},
		},
		"status": map[string]any{
			"loadBalancer": map[string]any{
				"ingress": []any{
					map[string]any{"ip": "34.123.45.67"},
				},
			},
		},
	}

	ingress := map[string]any{
		"kind": "Ingress",
		"spec": map[string]any{
			"ingressClassName": "nginx",
			"rules": []any{
				map[string]any{"host": "api.example.com"},
				map[string]any{"host": "www.example.com"},
			},
			"tls": []any{
				map[string]any{"hosts": []any{"api.example.com"}},
			},
		},
		"status": map[string]any{
			"loadBalancer": map[string]any{
				"ingress": []any{
					map[string]any{"hostname": "abc123.us-east-1.elb.amazonaws.com"},
				},
			},
		},
	}

	cronjob := map[string]any{
		"kind": "CronJob",
		"spec": map[string]any{
			"schedule": "*/5 * * * *",
		},
		"status": map[string]any{
			"active":           []any{map[string]any{"name": "job-1"}},
			"lastScheduleTime": "2026-05-19T10:00:00Z",
		},
	}

	data := map[string]any{
		"spec": map[string]any{
			"replicas":   3,
			"image":      "nginx:1.25",
			"regions":    []any{"us-east-1", "eu-west-1"},
			"hosts":      []string{"api.example.com", "admin.example.com"},
			"schedule":   "*/5 * * * *",
			"enabled":    true,
			"logLevel":   "info",
			"domain":     "https://www.acme.com/",
			"imageRef":   "myorg/app:v1.2.3",
			"ownerEmail": "team@myorg.io",
			"port":       "8080",
			"tags":       "latest,stable,v2",
		},
		"metadata": map[string]any{
			"name":      "hello-intent",
			"namespace": "default",
			"labels": map[string]any{
				"app":                    "frontend",
				"app.kubernetes.io/name": "my-app",
			},
			"annotations": map[string]any{
				"example.com/phase":      "Ready",
				"autoscale/min-replicas": "2",
			},
		},
		"children": map[string]any{
			"deployment": deployment,
			"pod":        pod,
			"service":    service,
			"ingress":    ingress,
			"cronjob":    cronjob,
		},
	}

	cases := []struct {
		name string
		expr string
		want string
	}{
		// ── baseline: the three you already have ─────────────────────────
		{"toInt on int", `{{ toInt .spec.replicas }}`, "3"},
		{"field access", `{{ .spec.image }}`, "nginx:1.25"},
		{"pipe add", `{{ .spec.replicas | add 1 }}`, "4"},

		// ── kubernetes metadata ──────────────────────────────────────────
		{"resourceName", `{{ resourceName .children.deployment }}`, "my-app"},
		{"resourceNamespace", `{{ resourceNamespace .children.deployment }}`, "production"},
		{"resourceUID", `{{ resourceUID .children.deployment }}`, "4b3f8d21-8e3a-4f8c-b9d2-1a2b3c4d5e6f"},
		{"resourceExists on map", `{{ resourceExists .children.deployment }}`, "true"},
		{"resourceExists on nil", `{{ resourceExists .children.missing }}`, "false"},

		// ── labels and annotations ───────────────────────────────────────
		{"getLabel", `{{ getLabel .children.deployment "app" }}`, "frontend"},
		{"getLabel missing", `{{ getLabel .children.deployment "nope" }}`, ""},
		{"hasLabel true", `{{ hasLabel .children.deployment "app" }}`, "true"},
		{"hasLabel false", `{{ hasLabel .children.deployment "nope" }}`, "false"},
		{"getAnnotation", `{{ getAnnotation .children.deployment "example.com/phase" }}`, "Ready"},
		{"getAnnotationInt", `{{ getAnnotationInt .children.deployment "autoscale/min-replicas" }}`, "2"},
		{"labelMatches true", `{{ labelMatches .children.deployment "app" "frontend" }}`, "true"},
		{"labelMatches extra labels ignored", `{{ labelMatches .children.deployment "app" "frontend" "tier" "web" }}`, "false"},

		// ── status and conditions ────────────────────────────────────────
		{"hasCondition Available", `{{ hasCondition .children.deployment "Available" }}`, "true"},
		{"hasCondition Missing", `{{ hasCondition .children.deployment "Missing" }}`, "false"},
		{"conditionReason", `{{ conditionReason .children.deployment "Available" }}`, "MinimumReplicasAvailable"},
		{"conditionMessage", `{{ conditionMessage .children.deployment "Progressing" }}`, "ReplicaSet has successfully progressed."},
		{"getStatus", `{{ getStatus .children.deployment "readyReplicas" }}`, "3"},

		// ── replica helpers ──────────────────────────────────────────────
		{"desiredReplicas", `{{ desiredReplicas .children.deployment }}`, "3"},
		{"readyReplicas", `{{ readyReplicas .children.deployment }}`, "3"},
		{"availableReplicas", `{{ availableReplicas .children.deployment }}`, "3"},
		{"updatedReplicas", `{{ updatedReplicas .children.deployment }}`, "3"},
		{"allReplicasReady", `{{ allReplicasReady .children.deployment }}`, "true"},
		{"rolloutComplete", `{{ rolloutComplete .children.deployment }}`, "true"},

		// ── container helpers ────────────────────────────────────────────
		{"containerImage", `{{ containerImage .children.deployment 0 }}`, "nginx:1.25"},
		{"containerPort  8080", `{{ containerPort  .children.deployment 0 }}`, "8080"},
		{"hasContainerPort  true", `{{ hasContainerPort  .children.deployment 0 8080 }}`, "true"},
		{"hasContainerPort  false", `{{ hasContainerPort  .children.deployment 0 9999 }}`, "false"},
		{"containerEnv", `{{ containerEnv .children.deployment "APP_ENV" 0 }}`, "production"},

		// ── pod helpers ──────────────────────────────────────────────────
		{"phase", `{{ phase .children.pod }}`, "Running"},

		// ── service helpers ──────────────────────────────────────────────
		{"serviceClusterIP", `{{ serviceClusterIP .children.service }}`, "10.96.0.42"},
		{"serviceNodePort", `{{ serviceNodePort .children.service }}`, "31234"},
		{"serviceLoadBalancerIP", `{{ serviceLoadBalancerIP .children.service }}`, "34.123.45.67"},

		// ── ingress helpers ──────────────────────────────────────────────
		{"ingressClassName", `{{ ingressClassName .children.ingress }}`, "nginx"},
		{"ingressHost", `{{ ingressHost .children.ingress }}`, "abc123.us-east-1.elb.amazonaws.com"},
		{"ingressRules", `{{ join (ingressRules .children.ingress) ", " }}`, "api.example.com, www.example.com"},
		{"ingressReady", `{{ ingressReady .children.ingress }}`, "true"},

		// ── cronjob helpers ──────────────────────────────────────────────
		{"cronJobActiveCount", `{{ cronJobActiveCount .children.cronjob }}`, "1"},
		{"cronJobLastScheduleTime", `{{ cronJobLastScheduleTime .children.cronjob }}`, "2026-05-19T10:00:00Z"},

		// ── strings ──────────────────────────────────────────────────────
		{"slugify", `{{ slugify "My Team / Platform" }}`, "my-team-platform"},
		{"camelToKebab", `{{ camelToKebab "WebsiteOperator" }}`, "website-operator"},
		{"concat", `{{ concat "webapp" "-" "prod" }}`, "webapp-prod"},
		{"domainHost", `{{ domainHost .spec.domain }}`, "www.acme.com"},
		{"domainBare", `{{ domainBare .spec.domain }}`, "acme.com"},
		{"split then index", `{{ index (split .spec.tags ",") 0 }}`, "latest"},
		{"join", `{{ join .spec.hosts ", " }}`, "api.example.com, admin.example.com"},

		// ── conditionals ─────────────────────────────────────────────────
		{"default on absent", `{{ default .spec.missing 2 }}`, "2"},
		{"default on zero", `{{ default .spec.missing 2 }}`, "2"},
		{"empty on nil", `{{ empty .spec.missing }}`, "true"},
		{"notEmpty on value", `{{ notEmpty .spec.image }}`, "true"},
		{"ternary", `{{ ternary .spec.enabled "on" "off" }}`, "on"},
		{"eqTernary match", `{{ eqTernary .spec.logLevel "info" "loud" "quiet" }}`, "loud"},
		{"eqTernary no match", `{{ eqTernary .spec.logLevel "debug" "loud" "quiet" }}`, "quiet"},
		{"boolTernary", `{{ boolTernary .spec.enabled "on" "off" }}`, "on"},

		// ── validation ───────────────────────────────────────────────────
		{"isValidEmail", `{{ isValidEmail .spec.ownerEmail }}`, "true"},
		{"isValidImageRef", `{{ isValidImageRef .spec.imageRef }}`, "true"},
		{"isValidPort", `{{ isValidPort .spec.port }}`, "true"},
		{"isValidLabelKey valid", `{{ isValidLabelKey "platform.myorg.io/tier" }}`, "true"},
		{"isValidLabelKey invalid", `{{ isValidLabelKey "bad key" }}`, "false"},
		{"isDNS1123Subdomain", `{{ isDNS1123Subdomain "my-app.example.com" }}`, "true"},

		// ── types ────────────────────────────────────────────────────────
		{"typeOf list", `{{ typeOf .spec.regions }}`, "slice"},
		{"typeList", `{{ typeList .spec.regions }}`, "true"},
		{"typeString false", `{{ typeString .spec.regions }}`, "false"},

		// ── collections ──────────────────────────────────────────────────
		{"listLen", `{{ listLen .spec.regions }}`, "2"},
		{"listGet", `{{ listGet .spec.regions 0 }}`, "us-east-1"},
		{"listHas true", `{{ listHas .spec.regions "us-east-1" }}`, "true"},
		{"listHas false", `{{ listHas .spec.regions "ap-south-1" }}`, "false"},
		{"mapGet", `{{ mapGet .metadata.labels "app" }}`, "frontend"},

		// ── math ─────────────────────────────────────────────────────────
		{"clamp low", `{{ clamp .spec.replicas 5 10 }}`, "5"},
		{"clamp high", `{{ clamp .spec.replicas 1 2 }}`, "2"},
		{"min", `{{ min .spec.replicas 2 }}`, "2"},
		{"max", `{{ max .spec.replicas 5 }}`, "5"},

		// ── safe access ──────────────────────────────────────────────────
		{"getIntOr present", `{{ getIntOr .spec.replicas 1 }}`, "3"},
		{"getIntOr absent", `{{ getIntOr .spec.missing 1 }}`, "1"},
		{"getStringOr present", `{{ getStringOr .spec.image "fallback" }}`, "nginx:1.25"},
		{"getStringOr absent", `{{ getStringOr .spec.missing "fallback" }}`, "fallback"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := template.New("t").Funcs(kel.FuncMap()).Parse(tc.expr)
			require.NoError(t, err, "parse %s", tc.expr)

			var buf bytes.Buffer
			require.NoError(t, tmpl.Execute(&buf, data), "execute %s", tc.expr)
			assert.Equal(t, tc.want, buf.String())
		})
	}
}
