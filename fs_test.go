package main

import (
	"embed"
	_ "embed"
	"io/fs"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/package-2.0.0
var packagesFS embed.FS

func packageFS(t *testing.T, path string) fs.FS {
	t.Helper()
	p, err := fs.Sub(packagesFS, path)
	require.NoError(t, err)
	return p
}

func TestDirValidation(t *testing.T) {
	cases := []struct {
		title string
		fs    fs.FS
		valid bool
	}{
		{"2.0.0 valid", packageFS(t, "testdata/package-2.0.0"), true},
	}

	cueCtx := cuecontext.New()

	spec := cueCtx.CompileString(manifestCue, cue.Filename("manifest.cue"))
	requireNoCueErr(t, spec.Err())

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			manifestYml, err := fs.ReadFile(c.fs, "manifest.yml")
			require.NoError(t, err)

			var manifest struct {
				FormatVersion string `yaml:"format_version"`
			}
			err = yaml.Unmarshal(manifestYml, &manifest)
			require.NoError(t, err)

			versionedSpec := spec.FillPath(cue.ParsePath("spec_version"), manifest.FormatVersion)
			manifestDef := versionedSpec.LookupDef("Package")
			requireNoCueErr(t, manifestDef.Err())

			packageExpr, err := FSExpr(c.fs)
			requireNoCueErr(t, err)

			packageValue := manifestDef.Context().BuildExpr(packageExpr)
			requireNoCueErr(t, err)

			t.Run("with subsume", func(t *testing.T) {
				requireNoCueErr(t, manifestDef.Subsume(packageValue, cue.Concrete(true)))
			})

			t.Run("with unify", func(t *testing.T) {
				v := packageValue.Unify(manifestDef)
				requireNoCueErr(t, v.Err())
				requireNoCueErr(t, v.Validate(cue.Concrete(true)))
			})
		})
	}
}
