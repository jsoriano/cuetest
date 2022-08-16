package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/encoding/json"
	"cuelang.org/go/encoding/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONSchemaBackwardsCompatibility(t *testing.T) {
	cases := []struct {
		oldVersion string
		newVersion string
		compatible bool
	}{
		{"1.14.0", "1.14.1", true},
		{"1.14.1", "1.14.1", true},
		{"1.14.0", "2.0.0", false},
		{"1.14.1", "2.0.0", false},
	}

	for _, c := range cases {
		t.Run(c.oldVersion+" to "+c.newVersion, func(t *testing.T) {
			oldSchema := loadJSONSchema(t, c.oldVersion)
			newSchema := loadJSONSchema(t, c.newVersion)

			t.Log(c.oldVersion + ":")
			t.Log(oldSchema)
			t.Log(c.newVersion + ":")
			t.Log(newSchema)

			err := newSchema.Subsume(oldSchema)
			if c.compatible {
				if !assert.NoError(t, err) {
					t.Log(cueerrors.Details(err, nil))
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func loadJSONSchema(t *testing.T, version string) cue.Value {
	t.Helper()

	schemaPath := filepath.Join("testdata", fmt.Sprintf("jsonschema-%s.json", version))
	d, err := ioutil.ReadFile(schemaPath)
	require.NoError(t, err)

	expr, err := json.Extract(schemaPath, d)
	require.NoError(t, err)

	ctx := cuecontext.New()
	v := ctx.BuildExpr(expr)
	require.NoError(t, err)

	jssConfig := jsonschema.Config{
		Strict: true,
	}
	jsonschema, err := jsonschema.Extract(v, &jssConfig)
	require.NoError(t, err)

	v = ctx.BuildFile(jsonschema)
	require.NoError(t, v.Err())
	return v
}
