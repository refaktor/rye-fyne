package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

func ParseDirFull(fset *token.FileSet, dirPath string) (pkgs map[string]*ast.Package, err error) {
	pkgs = make(map[string]*ast.Package)

	var module string
	{
		modPath := filepath.Join(dirPath, "go.mod")
		data, err := os.ReadFile(modPath)
		if err != nil {
			return nil, err
		}
		mod, err := modfile.Parse(modPath, data, nil)
		if err != nil {
			return nil, err
		}
		module = mod.Module.Mod.Path
	}

	var doParseDir func(fsPath, pkgPath string) error
	doParseDir = func(fsPath, pkgPath string) error {
		ents, err := os.ReadDir(fsPath)
		if err != nil {
			return err
		}
		for _, ent := range ents {
			fsPath := filepath.Join(fsPath, ent.Name())
			if ent.IsDir() {
				if ent.Name() == "test" || ent.Name() == "internal" || ent.Name() == "cmd" {
					continue
				}
				pkgPath := pkgPath + "/" + ent.Name()
				if err := doParseDir(fsPath, pkgPath); err != nil {
					return err
				}
			} else if strings.HasSuffix(ent.Name(), ".go") {
				if strings.HasSuffix(ent.Name(), "_test.go") {
					continue
				}
				src, err := parser.ParseFile(fset, fsPath, nil, 0)
				if err != nil {
					return err
				}
				pkgName := src.Name.Name
				if strings.HasSuffix(pkgName, "_test") || pkgName == "internal" || pkgName == "main" {
					continue
				}
				pkgPath := pkgPath
				pkg, ok := pkgs[pkgPath]
				if !ok {
					pkg = &ast.Package{
						Name:  pkgName,
						Files: make(map[string]*ast.File),
					}
					pkgs[pkgPath] = pkg
				}
				pkg.Files[fsPath] = src
			}
		}
		return nil
	}

	if err := doParseDir(dirPath, module); err != nil {
		return nil, err
	}
	return pkgs, nil
}
