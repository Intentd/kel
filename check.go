package kel

import (
	"fmt"
	"reflect"
	"text/template"
	"text/template/parse"
)

// Check reports whether expr is a template that can run with funcs: it parses,
// every function it calls exists, and each call passes as many arguments as
// the function takes. A value piped into a call counts as its last argument.
// Parsing alone finds syntax errors and unknown names; the argument counts are
// the mistake it misses, and otherwise only show up when the template runs.
// funcs nil is KEL's functions; pass FuncMap() with your own added to check
// templates that use both.
//
//	if err := kel.Check(`{{ conditionsState .obj "Ready" }}`, nil); err != nil {
//	    // conditionsState takes 1 argument, got 2
//	}
func Check(expr string, funcs template.FuncMap) error {
	if funcs == nil {
		funcs = funcMap()
	}
	tmpl, err := template.New("check").Funcs(funcs).Parse(expr)
	if err != nil {
		return err
	}
	for _, t := range tmpl.Templates() {
		if t.Tree == nil || t.Tree.Root == nil {
			continue
		}
		if err := checkNode(t.Tree, t.Tree.Root, funcs); err != nil {
			return err
		}
	}
	return nil
}

// funcMap is KEL's functions, read-only: Check never hands it out.
func funcMap() template.FuncMap {
	return funcs
}

// checkNode walks a parsed template and checks every function call in it.
func checkNode(tree *parse.Tree, node parse.Node, funcs template.FuncMap) error {
	switch n := node.(type) {
	case *parse.ListNode:
		if n == nil {
			return nil
		}
		for _, c := range n.Nodes {
			if err := checkNode(tree, c, funcs); err != nil {
				return err
			}
		}
	case *parse.ActionNode:
		return checkPipe(tree, n.Pipe, funcs)
	case *parse.IfNode:
		return checkBranch(tree, &n.BranchNode, funcs)
	case *parse.RangeNode:
		return checkBranch(tree, &n.BranchNode, funcs)
	case *parse.WithNode:
		return checkBranch(tree, &n.BranchNode, funcs)
	case *parse.TemplateNode:
		return checkPipe(tree, n.Pipe, funcs)
	}
	return nil
}

func checkBranch(tree *parse.Tree, b *parse.BranchNode, funcs template.FuncMap) error {
	if err := checkPipe(tree, b.Pipe, funcs); err != nil {
		return err
	}
	if err := checkNode(tree, b.List, funcs); err != nil {
		return err
	}
	return checkNode(tree, b.ElseList, funcs)
}

// checkPipe checks each command of a pipeline. Every command after the first
// receives the previous one's result as an extra, last argument.
func checkPipe(tree *parse.Tree, pipe *parse.PipeNode, funcs template.FuncMap) error {
	if pipe == nil {
		return nil
	}
	for i, cmd := range pipe.Cmds {
		piped := 0
		if i > 0 {
			piped = 1
		}
		if err := checkCommand(tree, cmd, piped, funcs); err != nil {
			return err
		}
	}
	return nil
}

// checkCommand checks one call: its argument count against the function's
// signature, and any parenthesised pipeline among its arguments.
func checkCommand(tree *parse.Tree, cmd *parse.CommandNode, piped int, funcs template.FuncMap) error {
	for _, arg := range cmd.Args {
		if p, ok := arg.(*parse.PipeNode); ok {
			if err := checkPipe(tree, p, funcs); err != nil {
				return err
			}
		}
	}
	if len(cmd.Args) == 0 {
		return nil
	}
	ident, ok := cmd.Args[0].(*parse.IdentifierNode)
	if !ok {
		return nil
	}
	fn, ok := funcs[ident.Ident]
	if !ok {
		return nil // a template built-in (and, eq, len, printf...): Go checks those when it runs
	}
	ft := reflect.TypeOf(fn)
	if ft == nil || ft.Kind() != reflect.Func {
		return nil
	}
	got := len(cmd.Args) - 1 + piped
	if want := ft.NumIn(); ft.IsVariadic() {
		if got < want-1 {
			return callError(tree, ident, fmt.Sprintf("at least %s", arguments(want-1)), got)
		}
	} else if got != want {
		return callError(tree, ident, arguments(want), got)
	}
	return nil
}

func callError(tree *parse.Tree, ident *parse.IdentifierNode, want string, got int) error {
	loc, _ := tree.ErrorContext(ident)
	return fmt.Errorf("%s: %s takes %s, got %d", loc, ident.Ident, want, got)
}

func arguments(n int) string {
	if n == 1 {
		return "1 argument"
	}
	return fmt.Sprintf("%d arguments", n)
}
