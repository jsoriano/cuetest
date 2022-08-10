package main

import (
	_ "embed"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	cueyaml "cuelang.org/go/pkg/encoding/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHumanErrors(t *testing.T) {
	cases := []struct {
		Title    string
		Schema   string
		Doc      string
		Expected []string
	}{
		{
			Title:  "unexpected value",
			Schema: "foo: 0 | 1 | 2",
			Doc:    "foo: 42",
			Expected: []string{
				`document.yml:1:7: foo: found 42, expected one of: 0, 1`,
			},
		},
		{
			Title:  "wrong type",
			Schema: "bar?: int",
			Doc:    "bar: foo",
			Expected: []string{
				`document.yml:1:7: bar: conflicting values "foo" and int (mismatched types string and int)`,
			},
		},
		{
			Title:  "unexpected field",
			Schema: "foo?: *0 | 1 | 2",
			Doc:    "bar: 42",
			Expected: []string{
				`document.yml:1:2: bar: field not allowed`,
			},
		},
		{
			Title:  "two errors expected",
			Schema: "foo: 0 | 1 | 2\nbar: int\n",
			Doc:    "foo: 42\nbar: foo\n",
			Expected: []string{
				`document.yml:1:7: foo: found 42, expected one of: 0, 1`,
				`document.yml:2:7: bar: conflicting values "foo" and int (mismatched types string and int)`,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Title, func(t *testing.T) {
			cueCtx := cuecontext.New()

			// It seems that definitions need to be used to detect unexpected fields.
			spec := cueCtx.CompileString("#Doc: {\n"+c.Schema+"\n}", cue.Filename("spec.cue")).LookupDef("Doc")
			require.NoError(t, spec.Err())

			expr, err := cueyaml.Unmarshal([]byte(c.Doc))
			require.NoError(t, err)

			v := spec.Context().BuildExpr(expr, cue.Filename("document.yml"))
			v = v.Unify(spec)
			errs := v.Validate(cue.Concrete(true))
			assert.NotEmpty(t, errs)

			formatted := ValidationErrors("document.yml", errs)
			for i, err := range formatted {
				t.Logf("#%d: %v", i, cueerrors.Details(err, nil))
			}

			assert.EqualValues(t, c.Expected, errorsToStrings(formatted))
		})
	}
}

func errorsToStrings(errs []error) []string {
	var formatted []string
	for _, err := range errs {
		formatted = append(formatted, err.Error())
	}
	return formatted
}
