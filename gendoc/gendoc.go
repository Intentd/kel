// Package gendoc reads KEL-style documentation from Go source: the doc
// comments of the functions a package registers, and of the builders that
// register them. KEL's registry and reference pages are generated from it,
// and a program that adds its own functions beside KEL's can document them
// the same way.
//
// A builder is a function returning a template.FuncMap literal. Its doc
// comment names the domain:
//
//	// cronFuncs — Cron
//	//
//	// Domain: cron
//	//
//	// Work with five-field cron expressions …
//	func cronFuncs() template.FuncMap {
//		return template.FuncMap{"cronValid": CronValid}
//	}
//
// Each registered function's doc comment starts with its Go name and has a
// Keywords line and an example:
//
//	// CronValid reports whether expr is a valid cron expression.
//	//
//	// Keywords: cron, schedule, validate
//	//
//	// Example:
//	//
//	//	{{ cronValid .spec.schedule }}  → true
//	//
//	// See also: cronDescribe
//	func CronValid(expr string) bool { … }
//
// Parse reports every place the source departs from this shape, so a
// function cannot be registered without being documented.
package gendoc

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Function is one documented function.
type Function struct {
	Name        string   // template name, e.g. "cronValid"
	GoName      string   // Go name, e.g. "CronValid"
	Domain      string   // e.g. "cron"
	Description string   // the doc text before Keywords, with the Go name replaced by Name
	Example     string   // the first example, without its indentation
	Keywords    []string // from the Keywords line
	SeeAlso     []string // from the See also line
}

// Domain is one group of functions, from a builder's doc comment.
type Domain struct {
	Name  string // e.g. "cron"
	Title string // e.g. "Cron"
	Intro string // the paragraphs after the Domain line
}

// Parse reads the non-test Go files in each directory. It returns the
// documented functions and domains, or an error listing every problem.
// Domains declared by more than one builder are merged: the first title,
// every intro.
func Parse(dirs ...string) ([]Function, []Domain, error) {
	var (
		fns     []Function
		domains []Domain
		byName  = map[string]string{}
		problem []string
	)
	for _, dir := range dirs {
		f, d, p := parseDir(dir)
		for _, fn := range f {
			if where, dup := byName[fn.Name]; dup {
				problem = append(problem, fmt.Sprintf("%s: %q is registered twice (also in %s)", dir, fn.Name, where))
				continue
			}
			byName[fn.Name] = dir
			fns = append(fns, fn)
		}
		domains = mergeDomains(domains, d)
		problem = append(problem, p...)
	}
	for _, fn := range fns {
		for _, s := range fn.SeeAlso {
			if _, ok := byName[s]; !ok {
				problem = append(problem, fmt.Sprintf("%s: See also %q is not a registered function", fn.Name, s))
			}
		}
	}
	sort.Slice(fns, func(i, j int) bool {
		if fns[i].Domain != fns[j].Domain {
			return fns[i].Domain < fns[j].Domain
		}
		return fns[i].Name < fns[j].Name
	})
	sort.Slice(domains, func(i, j int) bool { return domains[i].Name < domains[j].Name })
	if len(problem) > 0 {
		sort.Strings(problem)
		return fns, domains, errors.New(strings.Join(problem, "\n"))
	}
	return fns, domains, nil
}

func mergeDomains(into, add []Domain) []Domain {
	for _, d := range add {
		merged := false
		for i := range into {
			if into[i].Name == d.Name {
				if d.Intro != "" && !strings.Contains(into[i].Intro, d.Intro) {
					into[i].Intro = strings.TrimSpace(into[i].Intro + "\n\n" + d.Intro)
				}
				merged = true
			}
		}
		if !merged {
			into = append(into, d)
		}
	}
	return into
}

func parseDir(dir string) ([]Function, []Domain, []string) {
	fset := token.NewFileSet()
	paths, _ := filepath.Glob(filepath.Join(dir, "*.go"))
	decls := map[string]*ast.FuncDecl{}
	type registration struct {
		builder *ast.FuncDecl
		name    string
		goName  string
		pos     token.Position
	}
	var regs []registration
	var problem []string

	for _, p := range paths {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, p, nil, parser.ParseComments)
		if err != nil {
			problem = append(problem, err.Error())
			continue
		}
		for _, d := range file.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Recv != nil {
				continue
			}
			decls[fd.Name.Name] = fd
			if !returnsFuncMap(fd) {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				lit, ok := n.(*ast.CompositeLit)
				if !ok || !isFuncMap(lit.Type) {
					return true
				}
				for _, e := range lit.Elts {
					kv, ok := e.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					key, ok := kv.Key.(*ast.BasicLit)
					if !ok || key.Kind != token.STRING {
						continue
					}
					name, _ := strconv.Unquote(key.Value)
					regs = append(regs, registration{fd, name, exprName(kv.Value), fset.Position(kv.Pos())})
				}
				return false
			})
		}
	}

	var fns []Function
	var domains []Domain
	seenBuilder := map[*ast.FuncDecl]string{}
	for _, r := range regs {
		domain, ok := seenBuilder[r.builder]
		if !ok {
			d, err := parseBuilder(r.builder)
			if err != nil {
				problem = append(problem, fmt.Sprintf("%s: builder %s: %v", fset.Position(r.builder.Pos()), r.builder.Name.Name, err))
			} else {
				domains = mergeDomains(domains, []Domain{d})
			}
			domain = d.Name
			seenBuilder[r.builder] = domain
		}
		fd := decls[r.goName]
		if fd == nil {
			problem = append(problem, fmt.Sprintf("%s: %q is registered as %s, which is not declared in this package (wrap it in a documented function)", r.pos, r.name, r.goName))
			continue
		}
		fn, err := parseFunction(fd, r.name, domain)
		if err != nil {
			problem = append(problem, fmt.Sprintf("%s: %s (%q): %v", fset.Position(fd.Pos()), r.goName, r.name, err))
			continue
		}
		fns = append(fns, fn)
	}
	return fns, domains, problem
}

func parseBuilder(fd *ast.FuncDecl) (Domain, error) {
	if fd.Doc == nil {
		return Domain{}, errors.New("no doc comment; want \"<name> — <Title>\" and a Domain line")
	}
	paras := paragraphs(fd.Doc.Text())
	if len(paras) == 0 {
		return Domain{}, errors.New("empty doc comment")
	}
	head := paras[0]
	prefix := fd.Name.Name + " — "
	if !strings.HasPrefix(head, prefix) {
		return Domain{}, fmt.Errorf("doc comment must start with %q", prefix+"<Title>")
	}
	d := Domain{Title: strings.TrimSpace(strings.TrimPrefix(head, prefix))}
	rest := paras[1:]
	if len(rest) == 0 || !strings.HasPrefix(rest[0], "Domain:") {
		return Domain{}, errors.New("second paragraph must be \"Domain: <name>\"")
	}
	d.Name = strings.TrimSpace(strings.TrimPrefix(rest[0], "Domain:"))
	if d.Name == "" {
		return Domain{}, errors.New("empty Domain")
	}
	d.Intro = strings.Join(rest[1:], "\n\n")
	return d, nil
}

func parseFunction(fd *ast.FuncDecl, name, domain string) (Function, error) {
	if fd.Doc == nil {
		return Function{}, errors.New("no doc comment")
	}
	text := fd.Doc.Text()
	fn := Function{Name: name, GoName: fd.Name.Name, Domain: domain}
	lines := strings.Split(text, "\n")

	var desc []string
	section := "desc"
	var example []string
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "Keywords:"):
			for _, k := range strings.Split(strings.TrimPrefix(l, "Keywords:"), ",") {
				if k = strings.TrimSpace(k); k != "" {
					fn.Keywords = append(fn.Keywords, k)
				}
			}
			section = "after"
		case strings.HasPrefix(l, "See also:"):
			for _, s := range strings.Split(strings.TrimPrefix(l, "See also:"), ",") {
				if s = strings.TrimSpace(s); s != "" {
					fn.SeeAlso = append(fn.SeeAlso, s)
				}
			}
			section = "after"
		case strings.TrimSpace(l) == "Example:":
			section = "example"
		case section == "desc":
			desc = append(desc, l)
		case section == "example":
			if strings.HasPrefix(l, "\t") {
				example = append(example, strings.TrimPrefix(l, "\t"))
			} else if strings.TrimSpace(l) == "" && len(example) > 0 {
				example = append(example, "")
			}
		}
	}
	fn.Description = strings.TrimSpace(joinParagraphs(desc))
	fn.Example = strings.TrimRight(strings.Join(example, "\n"), "\n")

	switch {
	case !strings.HasPrefix(fn.Description, fn.GoName+" "):
		return fn, fmt.Errorf("doc comment must start with %q", fn.GoName)
	case len(fn.Keywords) == 0:
		return fn, errors.New("no Keywords line")
	case fn.Example == "":
		return fn, errors.New("no Example")
	}
	fn.Description = name + strings.TrimPrefix(fn.Description, fn.GoName)
	return fn, nil
}

// joinParagraphs joins wrapped lines into paragraphs; list items and
// indented lines keep their own line.
func joinParagraphs(lines []string) string {
	var out []string
	var cur []string
	flush := func() {
		if len(cur) > 0 {
			out = append(out, strings.Join(cur, " "))
			cur = nil
		}
	}
	for _, l := range lines {
		t := strings.TrimSpace(l)
		switch {
		case t == "":
			flush()
			out = append(out, "")
		case strings.HasPrefix(t, "- "):
			flush()
			cur = append(cur, t)
		default:
			if n := len(cur); n > 0 && hyphenated(cur[n-1]) {
				cur[n-1] += t
			} else {
				cur = append(cur, t)
			}
		}
	}
	flush()
	s := strings.Join(out, "\n")
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return s
}

func paragraphs(text string) []string {
	var out []string
	for _, p := range strings.Split(strings.TrimSpace(text), "\n\n") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, joinParagraphs(strings.Split(p, "\n")))
		}
	}
	return out
}

// hyphenated reports whether a line ends in a word broken at a hyphen
// ("day-" before "of-week"), so the next line joins it without a space.
func hyphenated(line string) bool {
	n := len(line)
	return n >= 2 && line[n-1] == '-' && line[n-2] != ' ' && line[n-2] != '-'
}

func returnsFuncMap(fd *ast.FuncDecl) bool {
	if fd.Type.Results == nil || len(fd.Type.Results.List) != 1 {
		return false
	}
	return isFuncMap(fd.Type.Results.List[0].Type)
}

func isFuncMap(e ast.Expr) bool {
	sel, ok := e.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "FuncMap"
}

func exprName(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return v.Sel.Name
	}
	return ""
}
