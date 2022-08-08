package main

import (
	_ "embed"
	"errors"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	cueyaml "cuelang.org/go/pkg/encoding/yaml"
	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/manifest.cue
var manifestCue string

//go:embed testdata/manifest-2.0.0.yml
var manifestYml2_0_0 string

//go:embed testdata/manifest-1.14.1.yml
var manifestYml1_14_1 string

//go:embed testdata/manifest-input.yml
var manifestYmlInput string

func TestValidation(t *testing.T) {
	cases := []struct {
		title       string
		manifestYml string
		valid       bool
	}{
		{"1.14.1 valid", manifestYml1_14_1, true},
		{"2.0.0 valid", manifestYml2_0_0, true},
		{"input valid", manifestYmlInput, true},
	}

	cueCtx := cuecontext.New()

	spec := cueCtx.CompileString(manifestCue, cue.Filename("manifest.cue"))
	requireNoCueErr(t, spec.Err())

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			var manifest struct {
				FormatVersion string `yaml:"format_version"`
			}
			err := yaml.Unmarshal([]byte(c.manifestYml), &manifest)
			require.NoError(t, err)

			versionedSpec := spec.FillPath(cue.ParsePath("spec_version"), manifest.FormatVersion)
			manifestDef := versionedSpec.LookupDef("PackageManifest")
			requireNoCueErr(t, manifestDef.Err())

			manifestExpr, err := cueyaml.Unmarshal([]byte(c.manifestYml))
			requireNoCueErr(t, err)

			manifestValue := manifestDef.Context().BuildExpr(manifestExpr)
			requireNoCueErr(t, err)

			t.Run("with subsume", func(t *testing.T) {
				requireNoCueErr(t, manifestDef.Subsume(manifestValue, cue.Concrete(true)))
			})

			t.Run("with unify", func(t *testing.T) {
				v := manifestValue.Unify(manifestDef)
				requireNoCueErr(t, v.Err())
				requireNoCueErr(t, v.Validate(cue.Concrete(true)))
			})

			t.Run("with validate", func(t *testing.T) {
				ok, err := cueyaml.Validate([]byte(c.manifestYml), manifestDef)
				assert.True(t, ok)
				requireNoCueErr(t, err)
			})
		})
	}
}

func TestBackwardsCompatibility(t *testing.T) {
	cueCtx := cuecontext.New()

	v := cueCtx.CompileString(manifestCue, cue.Filename("manifest.cue"))
	requireNoCueErr(t, v.Err())

	versions := []string{
		"1.14.0",
		"1.14.1",
		"1.15.0",
		"2.0.0",
		"2.1.0",
	}

	specVersion := cue.ParsePath("spec_version")
	for i := 1; i < len(versions); i++ {
		oldVersion := versions[i-1]
		newVersion := versions[i]
		t.Run(oldVersion+" to "+newVersion, func(t *testing.T) {
			oldMajor := semver.MustParse(oldVersion).Major()
			newMajor := semver.MustParse(newVersion).Major()

			if oldMajor != newMajor {
				t.Skip("no need to check backwards compatibility between majors")
			}

			oldDef := v.FillPath(specVersion, oldVersion).LookupDef("PackageManifest")
			requireNoCueErr(t, oldDef.Err())

			newDef := v.FillPath(specVersion, newVersion).LookupDef("PackageManifest")
			requireNoCueErr(t, newDef.Err())

			requireNoCueErr(t, oldDef.Subsume(newDef))
		})
	}
}

func requireNoCueErr(t *testing.T, err error) {
	t.Helper()

	var cueErr cueerrors.Error
	if !errors.As(err, &cueErr) {
		require.NoError(t, err)
		return
	}

	for _, err := range cueerrors.Errors(cueErr) {
		t.Logf("%s:%d:%d %s",
			err.Position().Filename(),
			err.Position().Line(),
			err.Position().Column(),
			err.Error())
	}
	t.FailNow()
}
