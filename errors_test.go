package main

import (
	_ "embed"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	cueyaml "cuelang.org/go/pkg/encoding/yaml"
	"github.com/stretchr/testify/require"
)

func TestHumanErrors(t *testing.T) {
	schema := "foo: *0 | 1 | 2\nbar: int"
	doc := "foo: 42\nbar: foo\nbaz: false\n"

	cueCtx := cuecontext.New()

	spec := cueCtx.CompileString(schema, cue.Filename("spec.cue"))
	require.NoError(t, spec.Err())

	expr, err := cueyaml.Unmarshal([]byte(doc))
	require.NoError(t, err)

	v := cueCtx.BuildExpr(expr, cue.Filename("document.yml"))
	v = v.Unify(spec)
	errs := v.Validate(cue.Concrete(true))
	for i, err := range ValidationErrors("document.yml", errs) {
		t.Logf("#%d: %v", i, cueerrors.Details(err, nil))
	}
}

var (
	reEmptyDisjunctionErr  = regexp.MustCompile(`^(.*): (\d+) errors in empty disjunction`)
	reConflictingValuesErr = regexp.MustCompile(`^(.*): conflicting values (.*) and (.*)`)
)

func ValidationErrors(filename string, err error) []error {
	var result []error
	for i, errs := 0, cueerrors.Errors(err); i < len(errs); i++ {
		e := errs[i]
		if m := reEmptyDisjunctionErr.FindStringSubmatch(e.Error()); len(m) > 0 {
			field := string(m[1])
			n, _ := strconv.Atoi(string(m[2]))
			pos := errs[i+1].InputPositions()[0]
			var expected []string
			var found string
			for _, conflict := range errs[i+1 : i+n] {
				m := reConflictingValuesErr.FindStringSubmatch(conflict.Error())
				expected = append(expected, string(m[2]))
				if found == "" {
					found = string(m[3])
				}
			}
			err := fmt.Errorf("%s:%d:%d: %s: found %s, expected one of: %s",
				filename, pos.Line(), pos.Column(), field,
				found, strings.Join(expected, ", "),
			)
			result = append(result, err)
			i += n
			continue
		}

		pos := e.InputPositions()[0]
		err := fmt.Errorf("%s:%d:%d: %s", filename, pos.Line(), pos.Column(), e)
		result = append(result, err)
	}
	return result
}
