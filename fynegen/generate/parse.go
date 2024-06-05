package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func visitDir(fset *token.FileSet, dirPath string, mode parser.Mode, modulePathHint string, onFile func(f *ast.File, filename, module string) error) (require []module.Version, err error) {
	noGoMod := false

	var modulePath string
	goModPath := filepath.Join(dirPath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		data, err := os.ReadFile(goModPath)
		if err != nil {
			return nil, err
		}
		mod, err := modfile.Parse(goModPath, data, nil)
		if err != nil {
			return nil, err
		}
		require = make([]module.Version, len(mod.Require))
		for i, v := range mod.Require {
			require[i] = v.Mod
		}
		modulePath = mod.Module.Mod.Path
	} else {
		noGoMod = true
		modulePath = modulePathHint
	}

	var requireMap map[string]struct{}

	var doVisitDir func(fsPath, modPath string) error
	doVisitDir = func(fsPath, modPath string) error {
		ents, err := os.ReadDir(fsPath)
		if err != nil {
			return err
		}
		for _, ent := range ents {
			fsPath := filepath.Join(fsPath, ent.Name())
			if ent.IsDir() {
				if strings.HasPrefix(ent.Name(), "_") || ent.Name() == "test" || ent.Name() == "testdata" || ent.Name() == "internal" || ent.Name() == "cmd" {
					continue
				}
				modPath := modPath + "/" + ent.Name()
				if err := doVisitDir(fsPath, modPath); err != nil {
					return err
				}
			} else if strings.HasSuffix(ent.Name(), ".go") {
				if strings.HasSuffix(ent.Name(), "_test.go") {
					continue
				}
				f, err := parser.ParseFile(fset, fsPath, nil, mode)
				if err != nil {
					return err
				}
				if noGoMod {
					for _, imp := range f.Imports {
						pkg, err := strconv.Unquote(imp.Path.Value)
						if err != nil {
							return err
						}
						if sp := strings.Split(pkg, "/"); len(sp) > 3 {
							pkg = strings.Join(sp[:3], "/")
						}
						requireMap[pkg] = struct{}{}
					}
				}
				modName := f.Name.Name
				if strings.HasSuffix(modName, "_test") || modName == "internal" || modName == "main" {
					continue
				}
				if err := onFile(f, fsPath, modPath); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if noGoMod {
		require = make([]module.Version, 0, len(requireMap))
		for req := range requireMap {
			require = append(require, module.Version{Path: req})
		}
	}

	if err := doVisitDir(dirPath, modulePath); err != nil {
		return nil, err
	}
	return require, nil
}

func ParseDirModules(fset *token.FileSet, dirPath, modulePathHint string) (modules map[string]string, require []module.Version, err error) {
	modules = make(map[string]string)
	require, err = visitDir(fset, dirPath, parser.PackageClauseOnly|parser.ImportsOnly, modulePathHint, func(f *ast.File, filename, module string) error {
		if name, ok := modules[module]; ok && name != f.Name.Name {
			return fmt.Errorf("package module %v has conflicting names: %v and %v", module, name, f.Name.Name)
		}
		modules[module] = f.Name.Name
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return modules, require, nil
}

func ParseDir(fset *token.FileSet, dirPath string, modulePathHint string) (pkgs map[string]*ast.Package, err error) {
	pkgs = make(map[string]*ast.Package)
	_, err = visitDir(fset, dirPath, 0, modulePathHint, func(f *ast.File, filename, module string) error {
		pkg, ok := pkgs[module]
		if !ok {
			pkg = &ast.Package{
				Name:  f.Name.Name,
				Files: make(map[string]*ast.File),
			}
			pkgs[module] = pkg
		}
		pkg.Files[filename] = f
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pkgs, nil
}
