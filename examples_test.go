package kel_test

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/intentd/kel"
)

// exampleResult splits "{{ expr }}  → result".
var exampleResult = regexp.MustCompile(`^(.*?\}\})\s*→\s*(.*)$`)

// varies lists functions whose result changes between runs or with the
// clock: their examples must run, but their results are not compared.
var varies = map[string]bool{
	"randomAlphanumeric": true, "randomBase64": true, "randomHex": true, "uuidv4": true,
	"timeAgo": true, "timeSince": true, "timeUntil": true, "nextCron": true,
	"weekday": true, "weekend": true, "timeInWindow": true, "timeNotInWindow": true,
	"cronJobNextScheduleTime": true, "podRestartedWithin": true,
}

// TestExamples runs every documented example against testdata/examples.yaml
// and checks it renders the result written after the arrow.
func TestExamples(t *testing.T) {
	raw, err := os.ReadFile("testdata/examples.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]any
	if err := yaml.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	for _, f := range kel.Builtin().Functions {
		for _, line := range strings.Split(f.Example, "\n") {
			if !strings.Contains(line, "{{") {
				continue
			}
			src, want := strings.TrimSpace(line), ""
			if m := exampleResult.FindStringSubmatch(line); m != nil {
				src, want = strings.TrimSpace(m[1]), strings.Trim(strings.TrimSpace(m[2]), `"`)
			}
			tmpl, err := template.New(f.Name).Funcs(kel.FuncMap()).Parse(src)
			if err != nil {
				t.Errorf("%s: %s: %v", f.Name, src, err)
				continue
			}
			var got bytes.Buffer
			if err := tmpl.Execute(&got, data); err != nil {
				t.Errorf("%s: %s: %v", f.Name, src, err)
				continue
			}
			if want != "" && !varies[f.Name] && got.String() != want {
				t.Errorf("%s: %s = %q, documented as %q", f.Name, src, got.String(), want)
			}
		}
	}
}
