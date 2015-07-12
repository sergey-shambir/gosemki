package main

import (
	"go/ast"
	"strings"
)

func GcDefaultImporter(imports map[string]*ast.Object, path string) (*ast.Object, error) {
	pkg := imports[path]
	if pkg == nil {
		name := path[strings.LastIndex(path, "/")+1:]
		pkg = ast.NewObj(ast.Pkg, name)
		pkg.Data = ast.NewScope(nil) // required by ast.NewPackage for dot-import
		imports[path] = pkg
	}
	imports[path] = pkg
	return pkg, nil
}
