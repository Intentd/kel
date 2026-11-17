// Package kelcel offers KEL's functions to CEL, the expression language
// Kubernetes uses for validation rules and admission policies.
//
// Library returns a CEL library: add it to any cel.Env and its expressions can
// ask KEL's questions about the objects they are given.
//
//	env, err := cel.NewEnv(
//	    kelcel.Library(),
//	    cel.Variable("object", cel.DynType),
//	)
//	// conditionsState(object) == "failed"
//	// object.conditionsState() == "failed"
//
// The functions are KEL's own, called through their Go signatures, so a
// template and a CEL expression get the same answer.
//
// Every function of KEL's kube and ecosystem domains is offered, and its cron
// and semver functions, except the few that read the clock (Excluded says
// which and why): CEL expressions must give the same answer every time. KEL's
// other general helpers (strings, lists, maps, math, conversions) are not
// offered: CEL has its own.
//
// Functions keep their KEL names, except where CEL reserves the word:
// namespace is namespaceOf. A function with a trailing optional argument
// (containerImage, get, labelMatches, firstState) takes none, one, or a list:
// containerImage(object), containerImage(object, 1), get(object, ["status", "phase"]).
//
// Env builds an environment with the library and the given variables; Check
// compiles an expression in it without running it (the CEL counterpart of
// kel.Check); Native turns a result back into plain Go maps and lists.
//
// CostEstimator gives CEL the cost of KEL's calls, for cost limits: one per
// call, plus one per element of each list the call walks (pods, events,
// metric series).
package kelcel
