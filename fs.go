package main

import (
	"io/fs"

	"cuelang.org/go/cue/ast"
	cueyaml "cuelang.org/go/pkg/encoding/yaml"
)

func FSExpr(root fs.FS) (ast.Expr, error) {
	// TODO: Try to evaluate lazily to avoid loading the whole FS in memory.
	var entries []interface{}
	err := fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}

		if d.IsDir() {
			sub, err := fs.Sub(root, path)
			if err != nil {
				return err
			}
			subExpr, err := FSExpr(sub)
			if err != nil {
				return err
			}

			field := ast.Field{
				Label: ast.NewString(d.Name()),
				Value: subExpr,
			}
			entries = append(entries, &field)
			return fs.SkipDir
		}

		content, err := fs.ReadFile(root, path)
		if err != nil {
			return err
		}

		expr, err := cueyaml.Unmarshal(content)
		if err != nil {
			return err
		}

		field := ast.Field{
			Label: ast.NewString(d.Name()),
			Value: expr,
		}
		entries = append(entries, &field)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ast.NewStruct(entries...), nil
}
