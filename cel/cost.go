package kelcel

import (
	"math"
	"strings"

	"cel.dev/cel-go/common/cost"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
)

// Estimator is what CostEstimator returns: the cost of KEL's calls, both
// estimated before an expression runs and measured while it runs.
type Estimator interface {
	cost.Estimator
	cost.ActualCostEstimator
}

// CostEstimator gives CEL the cost of KEL's calls, for cost limits: one per
// call, plus one per element of each list it walks. An object costs one: KEL
// reads a few fields of it, not all of it. Other calls are left to CEL.
//
//	estimate, err := env.EstimateCost(ast, kelcel.CostEstimator())
//	prg, err := env.Program(ast, cel.CostTracking(kelcel.CostEstimator()), cel.CostLimit(10000))
//
// Before running, a list's size is known only when the expression says it (a
// literal, or a declared size); otherwise its cost has no upper bound, as for
// CEL's own functions. An object declared as dyn that turns out to be a list
// is counted by its elements when the expression runs.
func CostEstimator() Estimator { return costs{} }

type costs struct{}

func isKEL(overloadID string) bool { return strings.HasPrefix(overloadID, "kel_") }

func (costs) EstimateCallCost(_, overloadID string, target *cost.AstNode, args []cost.AstNode) *cost.CallEstimate {
	if !isKEL(overloadID) {
		return nil
	}
	nodes := args
	if target != nil {
		nodes = append([]cost.AstNode{*target}, args...)
	}
	est := cost.Estimate{Min: 1, Max: 1}
	for _, n := range nodes {
		if n.Type() == nil || n.Type().Kind() != types.ListKind {
			continue
		}
		size := n.ComputedSize()
		if size == nil {
			est.Max = math.MaxUint64
			continue
		}
		est.Min = add(est.Min, size.Min)
		est.Max = add(est.Max, size.Max)
	}
	return &cost.CallEstimate{CostEstimate: est}
}

func (costs) EstimateSize(cost.AstNode) *cost.SizeEstimate { return nil }

func (costs) CallCost(_, overloadID string, args []ref.Val, _ ref.Val) *uint64 {
	if !isKEL(overloadID) {
		return nil
	}
	c := uint64(1)
	for _, a := range args {
		if l, ok := a.(traits.Lister); ok {
			if n, ok := l.Size().Value().(int64); ok && n > 0 {
				c = add(c, uint64(n))
			}
		}
	}
	return &c
}

// add is a + b, held at the largest value rather than wrapping around.
func add(a, b uint64) uint64 {
	if a > math.MaxUint64-b {
		return math.MaxUint64
	}
	return a + b
}
