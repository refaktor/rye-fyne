package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func ParseDirFull(fset *token.FileSet, path string) (pkgs map[string]*ast.Package, err error) {
	pkgs = make(map[string]*ast.Package)

	var doParseDir func(path string) error
	doParseDir = func(path string) error {
		ents, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, ent := range ents {
			path := filepath.Join(path, ent.Name())
			if ent.IsDir() {
				if ent.Name() == "test" || ent.Name() == "internal" {
					continue
				}
				if err := doParseDir(path); err != nil {
					return err
				}
			} else if strings.HasSuffix(ent.Name(), ".go") {
				if strings.HasSuffix(ent.Name(), "_test.go") {
					continue
				}
				src, err := parser.ParseFile(fset, path, nil, 0)
				if err != nil {
					return err
				}
				name := src.Name.Name
				if strings.HasSuffix(name, "_test") || name == "internal" || name == "main" {
					continue
				}
				pkg, ok := pkgs[name]
				if !ok {
					pkg = &ast.Package{
						Name:  name,
						Files: make(map[string]*ast.File),
					}
					pkgs[name] = pkg
				}
				pkg.Files[path] = src
			}
		}
		return nil
	}

	if err := doParseDir(path); err != nil {
		return nil, err
	}
	return pkgs, nil
}
