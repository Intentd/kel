package kelcel_test

import (
	"testing"
	"text/template"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/cost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/intentd/kel"
	kelcel "github.com/intentd/kel/cel"
	"github.com/intentd/kel/ecosystem"
	"github.com/intentd/kel/kube"
)

type m = map[string]interface{}
type l = []interface{}

// permission is a RabbitMQ Permission whose User does not exist yet.
var permission = m{
	"apiVersion": "rabbitmq.com/v1beta1",
	"kind":       "Permission",
	"metadata":   m{"name": "orders-app", "namespace": "payments", "generation": int64(1), "labels": m{"team": "orders"}},
	"status": m{
		"observedGeneration": int64(1),
		"conditions": l{
			m{"type": "Ready", "status": "False", "reason": "FailedCreateOrUpdate", "message": "failed create Permission, missing User"},
		},
	},
}

// database is a CloudNativePG Database: a field, not conditions.
var database = m{
	"metadata": m{"name": "orders", "generation": int64(1)},
	"status":   m{"applied": true, "observedGeneration": int64(1)},
}

func eval(t *testing.T, expr string, vars map[string]interface{}) interface{} {
	t.Helper()
	env, err := cel.NewEnv(kelcel.Library(), cel.Variable("object", cel.DynType), cel.Variable("db", cel.DynType))
	require.NoError(t, err)
	ast, iss := env.Compile(expr)
	require.NoError(t, iss.Err(), expr)
	prg, err := env.Program(ast)
	require.NoError(t, err)
	out, _, err := prg.Eval(vars)
	require.NoError(t, err, expr)
	return out.Value()
}

func TestSameAnswersAsTemplates(t *testing.T) {
	vars := map[string]interface{}{"object": permission, "db": database}
	cases := []struct {
		expr string
		want interface{}
	}{
		{`conditionsState(object)`, "failed"},
		{`object.conditionsState()`, "failed"},
		{`conditionsState(object) == "failed"`, true},
		{`conditionsStateMessage(object)`, "failed create Permission, missing User"},
		{`conditionsHealthy(object)`, false},
		{`hasCondition(object, "Ready")`, false},
		{`conditionReason(object, "Ready")`, "FailedCreateOrUpdate"},
		{`name(object) + "/" + namespaceOf(object)`, "orders-app/payments"},
		{`object.namespaceOf()`, "payments"},
		{`getLabel(object, "team")`, "orders"},
		{`generation(object) == observedGeneration(object)`, true},
		{`expectState(db.status.applied, true, false)`, "healthy"},
		{`firstState([conditionsState(db), expectState(db.status.applied, true, false)])`, "healthy"},
		{`isTerminating(object)`, false},
	}
	for _, c := range cases {
		t.Run(c.expr, func(t *testing.T) {
			assert.Equal(t, c.want, eval(t, c.expr, vars))
		})
	}
	// The Go function behind both gives the same answer.
	assert.Equal(t, kube.StateFailed, kube.ConditionsState(permission))
}

func TestMapsAndListsWrittenInCEL(t *testing.T) {
	// Objects built in the expression itself, not handed in as Go maps.
	got := eval(t, `conditionsState({"status": {"conditions": [{"type": "Ready", "status": "True"}]}})`, nil)
	assert.Equal(t, "healthy", got)
	got = eval(t, `hasCrashingPod([{"status": {"phase": "Running", "containerStatuses": [{"name": "app", "state": {"waiting": {"reason": "CrashLoopBackOff"}}}]}}])`, nil)
	assert.Equal(t, true, got)
}

func TestCheckedBeforeRunning(t *testing.T) {
	env, err := cel.NewEnv(kelcel.Library(), cel.Variable("object", cel.DynType))
	require.NoError(t, err)
	for _, expr := range []string{
		`conditionsState(object, "Ready")`, // too many arguments
		`hasCondition(object)`,             // too few
		`hasCondition(object, 1)`,          // the condition type is a string
		`conditionsState(object) == 1`,     // a state is a string
		`notAFunction(object)`,             // unknown
	} {
		_, iss := env.Compile(expr)
		assert.Error(t, iss.Err(), expr)
	}
}

func TestEveryOfferedFunctionIsKEL(t *testing.T) {
	funcs := kel.FuncMap()
	for celName, kelName := range kelcel.Functions() {
		_, ok := funcs[kelName]
		assert.True(t, ok, "%s calls %s, which is not a KEL function", celName, kelName)
	}
	assert.Equal(t, "namespace", kelcel.Functions()["namespaceOf"], "renamed: namespace is reserved in CEL")
	_, err := cel.NewEnv(kelcel.Library())
	assert.NoError(t, err)
}

func TestEveryKubeAndEcosystemFunctionIsOfferedOrExcluded(t *testing.T) {
	offered := map[string]bool{}
	for _, kelName := range kelcel.Functions() {
		offered[kelName] = true
	}
	excluded := kelcel.Excluded()
	for _, funcs := range []template.FuncMap{kube.Functions(), ecosystem.Functions()} {
		for name := range funcs {
			_, skip := excluded[name]
			assert.True(t, offered[name] != skip, "%s: offered %v, excluded %v", name, offered[name], skip)
		}
	}
	for name, why := range excluded {
		assert.NotEmpty(t, why, "%s is excluded without a reason", name)
	}
	assert.Greater(t, len(kelcel.Functions()), 150)
}

// deployment has two containers, for the optional-argument forms.
var deployment = m{
	"metadata": m{"name": "web", "labels": m{"app": "web", "tier": "front"}},
	"spec": m{"template": m{"spec": m{"containers": l{
		m{"name": "app", "image": "nginx:1.27", "ports": l{m{"containerPort": int64(8080)}}},
		m{"name": "sidecar", "image": "envoy:1.30"},
	}}}},
}

func TestOptionalArguments(t *testing.T) {
	vars := map[string]interface{}{"object": deployment, "db": database}
	cases := []struct {
		expr string
		want interface{}
	}{
		{`containerImage(object)`, "nginx:1.27"},      // none
		{`containerImage(object, 1)`, "envoy:1.30"},   // one
		{`containerImage(object, [1])`, "envoy:1.30"}, // a list
		{`object.containerImage(1)`, "envoy:1.30"},    // as a method
		{`get(db, "status")`, map[string]interface{}{"applied": true, "observedGeneration": int64(1)}},
		{`get(db, ["status", "applied"])`, true},
		{`labelMatches(object, ["app", "web", "tier", "front"])`, true},
		{`firstState("unknown")`, "unknown"},
		{`firstState(["unknown", "healthy"])`, "healthy"},
	}
	for _, c := range cases {
		t.Run(c.expr, func(t *testing.T) {
			assert.Equal(t, c.want, eval(t, c.expr, vars))
		})
	}
}

func TestResultsCELCanUse(t *testing.T) {
	vars := map[string]interface{}{"object": deployment, "db": database}
	// A map result (KEL's named Map) indexes like any CEL map.
	assert.Equal(t, "web", eval(t, `labels(object)["app"]`, vars))
	assert.Equal(t, true, eval(t, `status(db).applied`, vars))
	assert.Equal(t, 2.0*1024*1024*1024, eval(t, `parseQuantity("2Gi")`, nil))

	// A Go error becomes a CEL error.
	env, err := cel.NewEnv(kelcel.Library())
	require.NoError(t, err)
	ast, iss := env.Compile(`parseQuantity("not a quantity")`)
	require.NoError(t, iss.Err())
	prg, err := env.Program(ast)
	require.NoError(t, err)
	_, _, err = prg.Eval(map[string]interface{}{})
	assert.Error(t, err)
}

func TestCosts(t *testing.T) {
	env, err := cel.NewEnv(kelcel.Library(), cel.Variable("pods", cel.DynType), cel.Variable("object", cel.DynType))
	require.NoError(t, err)
	pod := m{"status": m{"phase": "Running"}}

	// Estimated before running: a call on an object costs 1 more than reading
	// the object; a literal list adds its size.
	estimate := func(expr string) uint64 {
		ast, iss := env.Compile(expr)
		require.NoError(t, iss.Err())
		est, err := env.EstimateCost(ast, kelcel.CostEstimator())
		require.NoError(t, err)
		return est.Min
	}
	assert.Equal(t, uint64(1), estimate(`conditionsState(object)`)-estimate(`object`))

	// KEL's share of a call over a literal list: the same expression estimated
	// without KEL's estimator charges the call 1.
	ast, iss := env.Compile(`hasCrashingPod([{}, {}, {}])`)
	require.NoError(t, iss.Err())
	withKEL, err := env.EstimateCost(ast, kelcel.CostEstimator())
	require.NoError(t, err)
	without, err := env.EstimateCost(ast, noEstimates{})
	require.NoError(t, err)
	assert.Equal(t, uint64(3), withKEL.Min-without.Min, "1 for the call + 3 pods, against CEL's 1")

	// Measured while running: the pods actually walked.
	actual := func(pods l) uint64 {
		ast, iss := env.Compile(`hasCrashingPod(pods)`)
		require.NoError(t, iss.Err())
		prg, err := env.Program(ast, cel.CostTracking(kelcel.CostEstimator()))
		require.NoError(t, err)
		_, det, err := prg.Eval(map[string]interface{}{"pods": pods})
		require.NoError(t, err)
		return *det.ActualCost()
	}
	one, three := actual(l{pod}), actual(l{pod, pod, pod})
	assert.Equal(t, uint64(2), three-one, "each extra pod costs 1")

	// A cost limit stops an expression that walks too much.
	ast, iss = env.Compile(`hasCrashingPod(pods)`)
	require.NoError(t, iss.Err())
	prg, err := env.Program(ast, cel.CostTracking(kelcel.CostEstimator()), cel.CostLimit(3))
	require.NoError(t, err)
	_, _, err = prg.Eval(map[string]interface{}{"pods": l{pod, pod, pod, pod, pod}})
	assert.Error(t, err)
}

// noEstimates leaves every cost to CEL's defaults.
type noEstimates struct{}

func (noEstimates) EstimateCallCost(string, string, *cost.AstNode, []cost.AstNode) *cost.CallEstimate {
	return nil
}
func (noEstimates) EstimateSize(cost.AstNode) *cost.SizeEstimate { return nil }

func TestEnvAndCheck(t *testing.T) {
	assert.NoError(t, kelcel.Check(`conditionsState(object) == "healthy"`, []string{"object"}))
	assert.Error(t, kelcel.Check(`conditionsState(object, "Ready")`, []string{"object"}), "too many arguments")
	assert.Error(t, kelcel.Check(`conditionsState(other)`, []string{"object"}), "undeclared variable")
	assert.Error(t, kelcel.Check(`conditionsState(object`, []string{"object"}), "syntax")

	// Extra options add to KEL's library.
	assert.NoError(t, kelcel.Check(`conditionsState(object) == want`, []string{"object"}, cel.Variable("want", cel.StringType)))
}

func TestNative(t *testing.T) {
	env, err := kelcel.Env([]string{"object"})
	require.NoError(t, err)
	ast, iss := env.Compile(`{"name": name(object), "ports": [1, 2], "nested": {"ok": true}}`)
	require.NoError(t, iss.Err())
	prg, err := env.Program(ast)
	require.NoError(t, err)
	out, _, err := prg.Eval(map[string]interface{}{"object": permission})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"name":   "orders-app",
		"ports":  []interface{}{int64(1), int64(2)},
		"nested": map[string]interface{}{"ok": true},
	}, kelcel.Native(out))
}

func TestCronAndSemver(t *testing.T) {
	vars := map[string]interface{}{"object": m{"spec": m{"schedule": "5 4 * * *", "version": "1.4.2"}}}
	assert.Equal(t, "4", eval(t, `cronHour(object.spec.schedule)`, vars))
	assert.Equal(t, true, eval(t, `cronValid(object.spec.schedule)`, vars))
	assert.Equal(t, true, eval(t, `semverConstraint(object.spec.version, "^1.4")`, vars))
	assert.Equal(t, uint64(1), eval(t, `semverMajor(object.spec.version)`, vars), "a CEL uint, as KEL returns it")
	assert.Contains(t, kelcel.Functions(), "cronMinute")
}
