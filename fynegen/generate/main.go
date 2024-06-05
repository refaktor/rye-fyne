package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"log"
	"os"
	"slices"
	"strings"

	"golang.org/x/mod/module"

	"github.com/refaktor/rye-front/fynegen/generate/repo"
)

var fset = token.NewFileSet()

func makeMakeRetArgErr(argn int, funcName string) func(allowedTypes ...string) string {
	return func(allowedTypes ...string) string {
		allowedTypesPfx := make([]string, len(allowedTypes))
		for i := range allowedTypes {
			allowedTypesPfx[i] = "env." + allowedTypes[i]
		}
		return fmt.Sprintf(
			`return evaldo.MakeArgError(ps, %v, []env.Type{%v}, "%v")`,
			argn,
			strings.Join(allowedTypesPfx, ", "),
			funcName,
		)
	}
}

func GenerateBinding(data *Data, fn *Func, indent int) (name string, code string, err error) {
	name = FuncRyeIdent(fn)

	var cb CodeBuilder
	cb.Indent = indent

	params := fn.Params
	if fn.Recv != nil {
		recvName, _ := NewIdent(nil, &ast.Ident{Name: "__recv"})
		params = append([]NamedIdent{{Name: recvName, Type: *fn.Recv}}, params...)
	}

	if len(params) > 5 {
		return "", "", errors.New("can only handle at most 5 parameters")
	}

	cb.Linef(`"%v": {`, name)
	cb.Indent++
	cb.Linef(`Doc: "%v",`, FuncGoIdent(fn))
	cb.Linef(`Argsn: %v,`, len(params))
	cb.Linef(`Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {`)
	cb.Indent++
	for i, param := range params {
		cb.Linef(`var arg%vVal %v`, i, param.Type.GoName)
		param.Type.MarkUsed(data)
		if _, found := ConvRyeToGo(
			data,
			&cb,
			param.Type,
			fmt.Sprintf(`arg%v`, i),
			fmt.Sprintf(`arg%vVal`, i),
			makeMakeRetArgErr(i, name),
		); !found {
			return "", "", errors.New("unhandled type conversion (rye to go): " + param.Type.GoName)
		}
	}

	var args strings.Builder
	{
		start := 0
		if fn.Recv != nil {
			start = 1
		}
		for i := start; i < len(params); i++ {
			param := params[i]
			if i != start {
				args.WriteString(`, `)
			}
			expand := ""
			if param.Type.IsEllipsis {
				expand = "..."
			}
			args.WriteString(fmt.Sprintf(`arg%vVal%v`, i, expand))
		}
	}

	var assign strings.Builder
	{
		for i := range fn.Results {
			if i != 0 {
				assign.WriteString(`, `)
			}
			assign.WriteString(fmt.Sprintf(`res%v`, i))
		}
		if len(fn.Results) > 0 {
			assign.WriteString(` := `)
		}
	}

	recv := ""
	if fn.Recv != nil {
		recv = `arg0Val.`
	}
	cb.Linef(`%v%v%v(%v)`, assign.String(), recv, fn.Name.GoName, args.String())
	fn.Name.MarkUsed(data)
	if len(fn.Results) > 0 {
		for i, result := range fn.Results {
			cb.Linef(`var res%vObj env.Object`, i)
			if _, found := ConvGoToRye(
				data,
				&cb,
				result.Type,
				fmt.Sprintf(`res%v`, i),
				fmt.Sprintf(`res%vObj`, i),
				nil,
			); !found {
				return "", "", errors.New("unhandled type conversion (go to rye): " + result.Type.GoName)
			}
		}
		if len(fn.Results) == 1 {
			cb.Linef(`return res0Obj`)
		} else {
			cb.Linef(`return env.NewDict(map[string]any{`)
			cb.Indent++
			for i, result := range fn.Results {
				cb.Linef(`"%v": res%vObj,`, result.Name.RyeName, i)
			}
			cb.Indent--
			cb.Linef(`})`)
		}
	} else {
		if fn.Recv == nil {
			cb.Linef(`return nil`)
		} else {
			cb.Linef(`return arg0`)
		}
	}
	cb.Indent--
	cb.Linef(`},`)
	cb.Indent--
	cb.Linef(`},`)

	return name, cb.String(), nil
}

func GenerateGetterOrSetter(data *Data, field NamedIdent, structName Ident, indent int, ptrToStruct, setter bool) (name string, code string, err error) {
	if ptrToStruct {
		var err error
		structName, err = NewIdent(structName.File, &ast.StarExpr{X: structName.Expr})
		if err != nil {
			return "", "", err
		}
	}

	if setter {
		name = fmt.Sprintf("%v//%v!", structName.RyeName, field.Name.RyeName)
	} else {
		name = fmt.Sprintf("%v//%v?", structName.RyeName, field.Name.RyeName)
	}

	var cb CodeBuilder
	cb.Indent = indent

	cb.Linef(`"%v": {`, name)
	cb.Indent++
	if setter {
		cb.Linef(`Doc: "Set %v %v value",`, structName.GoName, field.Name.GoName)
		cb.Linef(`Argsn: 2,`)
	} else {
		cb.Linef(`Doc: "Get %v %v value",`, structName.GoName, field.Name.GoName)
		cb.Linef(`Argsn: 1,`)
	}
	cb.Linef(`Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {`)
	cb.Indent++

	cb.Linef(`var self %v`, structName.GoName)
	structName.MarkUsed(data)
	if _, found := ConvRyeToGo(
		data,
		&cb,
		structName,
		`arg0`,
		`self`,
		makeMakeRetArgErr(0, name),
	); !found {
		return "", "", errors.New("unhandled type conversion (go to rye): " + structName.GoName)
	}

	if setter {
		if _, found := ConvRyeToGo(
			data,
			&cb,
			field.Type,
			`arg1`,
			`self.`+field.Name.GoName,
			makeMakeRetArgErr(1, name),
		); !found {
			return "", "", errors.New("unhandled type conversion (go to rye): " + structName.GoName)
		}

		cb.Linef(`return arg0`)
	} else {
		cb.Linef(`var resObj env.Object`)
		if _, found := ConvGoToRye(
			data,
			&cb,
			field.Type,
			`self.`+field.Name.GoName,
			`resObj`,
			nil,
		); !found {
			return "", "", errors.New("unhandled type conversion (go to rye): " + field.Type.GoName)
		}
		cb.Linef(`return resObj`)
	}

	cb.Indent--
	cb.Linef(`},`)
	cb.Indent--
	cb.Linef(`},`)

	return name, cb.String(), nil
}

func main() {
	outFile := "../current/fynegen/builtins_fyne.go"

	dstPath := "_srcrepos"
	srcModule := "fyne.io/fyne/v2"
	srcModuleVersion := "v2.4.4"

	getRepo := func(pkg, semver string) (string, error) {
		have, dir, err := repo.Have(dstPath, pkg, semver)
		if err != nil {
			return "", err
		}
		if !have {
			log.Printf("downloading %v %v\n", pkg, semver)
			_, err := repo.Get(dstPath, pkg, semver)
			if err != nil {
				return "", err
			}
		}
		return dir, nil
	}

	srcDir, err := getRepo(srcModule, srcModuleVersion)
	if err != nil {
		fmt.Println("get repo:", err)
		os.Exit(1)
	}

	moduleNames := make(map[string]string)
	{
		addPkgNames := func(dir, modulePath string) ([]module.Version, error) {
			pkgNms, req, err := ParseDirModules(fset, dir, modulePath)
			if err != nil {
				return nil, err
			}
			for mod, name := range pkgNms {
				moduleNames[mod] = name
			}
			return req, nil
		}
		req, err := addPkgNames(srcDir, srcModule)
		if err != nil {
			fmt.Println("parse modules:", err)
			os.Exit(1)
		}
		for _, v := range req {
			dir, err := getRepo(v.Path, v.Version)
			if err != nil {
				fmt.Println("get repo:", err)
				os.Exit(1)
			}
			if _, err := addPkgNames(dir, v.Path); err != nil {
				fmt.Println("parse modules:", err)
				os.Exit(1)
			}
		}
	}

	pkgs, err := ParseDir(fset, srcDir, srcModule)
	if err != nil {
		fmt.Println("parse source:", err)
		os.Exit(1)
	}

	const bindingCodeIndent = 1

	data := NewData()
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			if err := data.AddFile(f, pkg.Path, moduleNames); err != nil {
				fmt.Printf("%v: %v\n", pkg.Name, err)
			}
		}
	}
	if err := data.ResolveInheritances(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	generatedFuncs := make(map[string]string)

	for _, iface := range data.Interfaces {
		for _, fn := range iface.Funcs {
			name, code, err := GenerateBinding(data, fn, bindingCodeIndent)
			if err != nil {
				fmt.Println(name+":", err)
				continue
			}
			generatedFuncs[name] = code
		}
	}

	for _, fn := range data.Funcs {
		name, code, err := GenerateBinding(data, fn, bindingCodeIndent)
		if err != nil {
			fmt.Println(name+":", err)
			continue
		}
		generatedFuncs[name] = code
	}

	for _, struc := range data.Structs {
		for _, f := range struc.Fields {
			for _, ptrToStruct := range []bool{false, true} {
				for _, setter := range []bool{false, true} {
					name, code, err := GenerateGetterOrSetter(data, f, struc.Name, bindingCodeIndent, ptrToStruct, setter)
					if err != nil {
						fmt.Println(struc.Name.GoName+"."+f.Name.GoName+":", err)
						continue
					}
					generatedFuncs[name] = code
				}
			}
		}
	}

	data.UsedImports["github.com/refaktor/rye/env"] = struct{}{}
	data.UsedImports["github.com/refaktor/rye/evaldo"] = struct{}{}

	var cb CodeBuilder
	cb.Linef(`//go:build b_fynegen`)
	cb.Linef(``)
	cb.Linef(`// Code generated by generator/generate. DO NOT EDIT.`)
	cb.Linef(``)
	cb.Linef(`package fynegen`)
	cb.Linef(``)
	cb.Linef(`import (`)
	cb.Indent++
	usedImportKeys := make([]string, 0, len(data.UsedImports))
	for k := range data.UsedImports {
		usedImportKeys = append(usedImportKeys, k)
	}
	slices.Sort(usedImportKeys)
	for _, imp := range usedImportKeys {
		cb.Linef(`"%v"`, imp)
	}
	cb.Indent--
	cb.Linef(`)`)
	cb.Linef(``)

	cb.Linef(`// Force-use evaldo and env packages since tracking them would be too complicated`)
	cb.Linef(`var _ = evaldo.BuiltinNames`)
	cb.Linef(`var _ = env.Object(nil)`)
	cb.Linef(``)

	cb.Linef(`func boolToInt64(x bool) int64 {`)
	cb.Indent++
	cb.Linef(`var res int64`)
	cb.Linef(`if x {`)
	cb.Indent++
	cb.Linef(`res = 1`)
	cb.Indent--
	cb.Linef(`}`)
	cb.Linef(`return res`)
	cb.Indent--
	cb.Linef(`}`)
	cb.Linef(``)

	cb.Linef(`var Builtins_fynegen = map[string]*env.Builtin{`)
	cb.Indent++

	cb.Linef(`"nil": {`)
	cb.Indent++
	cb.Linef(`Doc: "nil value for go types",`)
	cb.Linef(`Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {`)
	cb.Indent++
	cb.Linef(`return *env.NewInteger(0)`)
	cb.Indent--
	cb.Linef(`},`)
	cb.Indent--
	cb.Linef(`},`)

	generatedFuncKeys := make([]string, 0, len(generatedFuncs))
	for k := range generatedFuncs {
		generatedFuncKeys = append(generatedFuncKeys, k)
	}
	slices.Sort(generatedFuncKeys)
	for _, k := range generatedFuncKeys {
		cb.Write(generatedFuncs[k])
	}

	cb.Indent--
	cb.Linef(`}`)

	code, err := format.Source([]byte(cb.String()))
	if err != nil {
		fmt.Println("gofmt:", err)
		os.Exit(1)
	}
	//code := []byte(cb.String())

	if err := os.WriteFile(outFile, code, 0666); err != nil {
		panic(err)
	}
	log.Printf("Wrote bindings containing %v functions to %v", len(generatedFuncs), outFile)
}
