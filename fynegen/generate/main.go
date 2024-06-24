package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"log"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/mod/module"

	"github.com/BurntSushi/toml"
	"github.com/iancoleman/strcase"
	"github.com/refaktor/rye-front/fynegen/generate/repo"
)

var fset = token.NewFileSet()

func makeMakeRetArgErr(argn int) func(allowedTypes ...string) string {
	return func(allowedTypes ...string) string {
		allowedTypesPfx := make([]string, len(allowedTypes))
		for i := range allowedTypes {
			allowedTypesPfx[i] = "env." + allowedTypes[i]
		}
		return fmt.Sprintf(
			`return evaldo.MakeArgError(ps, %v, []env.Type{%v}, "((RYEGEN:FUNCNAME))")`,
			argn,
			strings.Join(allowedTypesPfx, ", "),
		)
	}
}

type BindingFunc struct {
	Recv      string
	Name      string
	NameIdent *Ident
	Doc       string
	Argsn     int
	Body      string
}

func (bind *BindingFunc) FullName() string {
	if bind.Recv != "" {
		return bind.Recv + "//" + bind.Name
	} else {
		return bind.Name
	}
}

func (bind *BindingFunc) SplitGoNameAndMod() (string, *File) {
	file := bind.NameIdent.File
	var name string
	switch expr := bind.NameIdent.Expr.(type) {
	case *ast.Ident:
		name = expr.Name
	case *ast.SelectorExpr:
		mod, ok := expr.X.(*ast.Ident)
		if !ok {
			panic("expected ast.SelectorExpr.X to be of type *ast.Ident")
		}
		file, ok = file.ImportsByName[mod.Name]
		if !ok {
			panic(fmt.Errorf("module %v imported by %v not found", mod.Name, file.Name))
		}
		name = expr.Sel.Name
	default:
		panic("expected func name identifier to be of type *ast.Ident or *ast.SelectorExpr")
	}
	return name, file
}

func GenerateBinding(ctx *Context, fn *Func, indent int) (*BindingFunc, error) {
	res := &BindingFunc{}
	res.Name = fn.Name.RyeName
	res.NameIdent = &fn.Name
	if fn.Recv != nil {
		res.Recv = fn.Recv.RyeName
	}

	var cb CodeBuilder
	cb.Indent = indent

	params := fn.Params
	if fn.Recv != nil {
		recvName, _ := NewIdent(ctx, nil, &ast.Ident{Name: "__recv"})
		params = append([]NamedIdent{{Name: recvName, Type: *fn.Recv}}, params...)
	}

	if len(params) > 5 {
		return nil, errors.New("can only handle at most 5 parameters")
	}

	res.Doc = FuncGoIdent(fn)
	res.Argsn = len(params)

	for i, param := range params {
		cb.Linef(`var arg%vVal %v`, i, param.Type.GoName)
		ctx.MarkUsed(param.Type)
		if _, found := ConvRyeToGo(
			ctx,
			&cb,
			param.Type,
			fmt.Sprintf(`arg%vVal`, i),
			fmt.Sprintf(`arg%v`, i),
			makeMakeRetArgErr(i),
		); !found {
			return nil, errors.New("unhandled type conversion (rye to go): " + param.Type.GoName)
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
	ctx.MarkUsed(fn.Name)
	if len(fn.Results) > 0 {
		for i, result := range fn.Results {
			cb.Linef(`var res%vObj env.Object`, i)
			if _, found := ConvGoToRye(
				ctx,
				&cb,
				result.Type,
				fmt.Sprintf(`res%vObj`, i),
				fmt.Sprintf(`res%v`, i),
				nil,
			); !found {
				return nil, errors.New("unhandled type conversion (go to rye): " + result.Type.GoName)
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
	res.Body = cb.String()

	return res, nil
}

func GenerateGetterOrSetter(ctx *Context, field NamedIdent, structName Ident, indent int, ptrToStruct, setter bool) (*BindingFunc, error) {
	res := &BindingFunc{}

	if ptrToStruct {
		var err error
		structName, err = NewIdent(ctx, structName.File, &ast.StarExpr{X: structName.Expr})
		if err != nil {
			return nil, err
		}
	}

	res.Recv = structName.RyeName
	if setter {
		res.Name = field.Name.RyeName + "!"
	} else {
		res.Name = field.Name.RyeName + "?"
	}

	var cb CodeBuilder
	cb.Indent = indent

	if setter {
		res.Doc = fmt.Sprintf("Set %v %v value", structName.GoName, field.Name.GoName)
		res.Argsn = 2
	} else {
		res.Doc = fmt.Sprintf("Get %v %v value", structName.GoName, field.Name.GoName)
		res.Argsn = 1
	}

	cb.Linef(`var self %v`, structName.GoName)
	ctx.MarkUsed(structName)
	if _, found := ConvRyeToGo(
		ctx,
		&cb,
		structName,
		`self`,
		`arg0`,
		makeMakeRetArgErr(0),
	); !found {
		return nil, errors.New("unhandled type conversion (go to rye): " + structName.GoName)
	}

	if setter {
		if _, found := ConvRyeToGo(
			ctx,
			&cb,
			field.Type,
			`self.`+field.Name.GoName,
			`arg1`,
			makeMakeRetArgErr(1),
		); !found {
			return nil, errors.New("unhandled type conversion (go to rye): " + structName.GoName)
		}

		cb.Linef(`return arg0`)
	} else {
		cb.Linef(`var resObj env.Object`)
		if _, found := ConvGoToRye(
			ctx,
			&cb,
			field.Type,
			`resObj`,
			`self.`+field.Name.GoName,
			nil,
		); !found {
			return nil, errors.New("unhandled type conversion (go to rye): " + field.Type.GoName)
		}
		cb.Linef(`return resObj`)
	}
	res.Body = cb.String()

	return res, nil
}

// Order of importance (descending):
// - Part of stdlib
// - Prefix of preferPkg
// - Shorter path
// - Smaller string according to strings.Compare
func makeCompareModulePaths(preferPkg string) func(a, b string) int {
	return func(a, b string) int {
		{
			aSp := strings.SplitN(a, "/", 2)
			bSp := strings.SplitN(b, "/", 2)
			if len(aSp) > 0 && len(bSp) > 0 {
				aStd := !strings.Contains(aSp[0], ".")
				bStd := !strings.Contains(bSp[0], ".")
				if aStd && !bStd {
					return -1
				} else if !aStd && bStd {
					return 1
				}
			}
		}
		if preferPkg != "" {
			aPfx := strings.HasPrefix(a, preferPkg)
			bPfx := strings.HasPrefix(b, preferPkg)
			if aPfx && !bPfx {
				return -1
			} else if !aPfx && bPfx {
				return 1
			}
		}
		if len(a) < len(b) {
			return -1
		} else if len(a) > len(b) {
			return 1
		}
		return strings.Compare(a, b)
	}
}

func main() {
	outFile := "../current/fynegen/builtins_fyne.go"

	configPath := "config.toml"
	if _, err := os.Stat(configPath); err != nil {
		if err := os.WriteFile(configPath, []byte(DefaultConfig), 0666); err != nil {
			fmt.Println("create default config:", err)
			os.Exit(1)
		}
		fmt.Println("created default config at", configPath)
		os.Exit(0)
	}
	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		fmt.Println("open config:", err)
		os.Exit(1)
	}

	dstPath := "_srcrepos"

	getRepo := func(pkg, version string) (string, error) {
		have, dir, _, err := repo.Have(dstPath, pkg, version)
		if err != nil {
			return "", err
		}
		if !have {
			log.Printf("downloading %v %v\n", pkg, version)
			_, err := repo.Get(dstPath, pkg, version)
			if err != nil {
				return "", err
			}
		}
		return dir, nil
	}

	srcDir, err := getRepo(cfg.Package, cfg.Version)
	if err != nil {
		fmt.Println("get repo:", err)
		os.Exit(1)
	}

	moduleNames := make(map[string]string) // module path to name
	{
		addPkgNames := func(dir, modulePath string) (string, []module.Version, error) {
			goVer, pkgNms, req, err := ParseDirModules(fset, dir, modulePath)
			if err != nil {
				return "", nil, err
			}
			for mod, name := range pkgNms {
				moduleNames[mod] = name
			}
			return goVer, req, nil
		}
		goVer, req, err := addPkgNames(srcDir, cfg.Package)
		if err != nil {
			fmt.Println("parse modules:", err)
			os.Exit(1)
		}
		req = append(req, module.Version{Path: "std", Version: goVer})
		for _, v := range req {
			dir, err := getRepo(v.Path, v.Version)
			if err != nil {
				fmt.Println("get repo:", err)
				os.Exit(1)
			}
			if _, _, err := addPkgNames(dir, v.Path); err != nil {
				fmt.Println("parse modules:", err)
				os.Exit(1)
			}
		}
	}
	moduleImportNames := make(map[string]string) // module path to name; each name value is unique
	{
		moduleNameKeys := make([]string, 0, len(moduleNames))
		for k := range moduleNames {
			moduleNameKeys = append(moduleNameKeys, k)
		}
		slices.SortFunc(moduleNameKeys, makeCompareModulePaths(cfg.Package))

		moduleNameIdxs := make(map[string]int) // module name to number of occurrences
		for _, mod := range moduleNameKeys {
			name := moduleNames[mod]
			impName := name
			if idx := moduleNameIdxs[name]; idx > 0 {
				impName += "_" + strconv.Itoa(idx)
			}
			moduleImportNames[mod] = impName
			moduleNameIdxs[name]++
		}
	}

	startTime := time.Now()

	pkgs, err := ParseDir(fset, srcDir, cfg.Package)
	if err != nil {
		fmt.Println("parse source:", err)
		os.Exit(1)
	}

	const bindingCodeIndent = 3

	data := &Data{
		Funcs:      make(map[string]*Func),
		Interfaces: make(map[string]*Interface),
		Structs:    make(map[string]*Struct),
		Typedefs:   make(map[string]Ident),
	}
	ctx := &Context{
		Config:      &cfg,
		Data:        data,
		ModuleNames: moduleImportNames,
		UsedImports: make(map[string]struct{}),
		UsedTyps: make(map[string]Ident),
	}

	for _, pkg := range pkgs {
		for name, f := range pkg.Files {
			name = strings.TrimPrefix(name, dstPath+string(filepath.Separator))
			if err := data.AddFile(ctx, f, name, pkg.Path, moduleNames); err != nil {
				fmt.Printf("%v: %v\n", pkg.Name, err)
			}
		}
	}
	if err := data.ResolveInheritancesAndMethods(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bindingFuncs := make(map[string]*BindingFunc)

	for _, iface := range data.Interfaces {
		if IdentIsInternal(ctx, iface.Name) {
			continue
		}
		for _, fn := range iface.Funcs {
			bind, err := GenerateBinding(ctx, fn, bindingCodeIndent)
			if err != nil {
				fmt.Println(fn.String()+":", err)
				continue
			}
			bindingFuncs[bind.FullName()] = bind
		}
	}

	for _, fn := range data.Funcs {
		if IdentIsInternal(ctx, fn.Name) || (fn.Recv != nil && IdentIsInternal(ctx, *fn.Recv)) {
			continue
		}
		bind, err := GenerateBinding(ctx, fn, bindingCodeIndent)
		if err != nil {
			fmt.Println(fn.String()+":", err)
			continue
		}
		bindingFuncs[bind.FullName()] = bind
	}

	for _, struc := range data.Structs {
		if IdentIsInternal(ctx, struc.Name) {
			continue
		}
		for _, f := range struc.Fields {
			for _, ptrToStruct := range []bool{false, true} {
				for _, setter := range []bool{false, true} {
					bind, err := GenerateGetterOrSetter(ctx, f, struc.Name, bindingCodeIndent, ptrToStruct, setter)
					if err != nil {
						s := struc.Name.RyeName + "//" + f.Name.RyeName
						if setter {
							s += "!"
						} else {
							s += "?"
						}
						fmt.Println(s+":", err)
						continue
					}
					bindingFuncs[bind.FullName()] = bind
				}
			}
		}
	}

	// rye ident to list of modules with priority
	bindingFuncPrios := make(map[string][]struct {
		Mod  string
		Prio int // less => higher prio
	})
	addBindingFuncPrio := func(bind *BindingFunc) {
		if bind.NameIdent == nil || bind.Recv != "" {
			return
		}

		name, file := bind.SplitGoNameAndMod()
		if ctx.Config.CutNew {
			name = strcase.ToKebab(strings.TrimPrefix(name, "New"))
			if name == "" {
				name = strcase.ToKebab(ctx.ModuleNames[file.ModulePath])
			}
		}

		ps := bindingFuncPrios[name]
		for _, p := range ps {
			if file.ModulePath == p.Mod {
				return
			}
		}

		prio := math.MaxInt
		for i, v := range ctx.Config.NoPrefix {
			if v == file.ModulePath {
				prio = i
			}
		}
		if prio == math.MaxInt {
			return
		}

		bindingFuncPrios[name] = append(bindingFuncPrios[name], struct {
			Mod  string
			Prio int
		}{
			Mod:  file.ModulePath,
			Prio: prio,
		})
	}
	for _, bind := range bindingFuncs {
		if bind.Recv != "" {
			continue
		}
		addBindingFuncPrio(bind)
	}
	for _, bind := range bindingFuncs {
		if bind.NameIdent == nil {
			continue
		}

		newName, file := bind.SplitGoNameAndMod()

		newNameIsPfx := false
		if ctx.Config.CutNew {
			newName = strings.TrimPrefix(newName, "New")
			if newName == "" {
				newName = strcase.ToKebab(ctx.ModuleNames[file.ModulePath])
				newNameIsPfx = true
			}
		}
		newName = strcase.ToKebab(newName)

		if bind.Recv == "" {
			prios := bindingFuncPrios[newName]
			isHighestPrio := len(prios) > 0 && slices.MinFunc(prios, func(a, b struct {
				Mod  string
				Prio int
			}) int {
				return a.Prio - b.Prio
			}).Mod == file.ModulePath
			if !(isHighestPrio || (newNameIsPfx && len(prios) == 0)) {
				newName = strcase.ToKebab(ctx.ModuleNames[file.ModulePath]) + "-" + newName
			}
		}

		bind.Name = newName
	}

	ctx.UsedImports["github.com/refaktor/rye/env"] = struct{}{}
	ctx.UsedImports["github.com/refaktor/rye/evaldo"] = struct{}{}

	var cb CodeBuilder
	cb.Linef(`//go:build b_fynegen`)
	cb.Linef(``)
	cb.Linef(`// Code generated by generator/generate. DO NOT EDIT.`)
	cb.Linef(``)
	cb.Linef(`package fynegen`)
	cb.Linef(``)
	cb.Linef(`import (`)
	cb.Indent++
	usedImportKeys := make([]string, 0, len(ctx.UsedImports))
	for k := range ctx.UsedImports {
		usedImportKeys = append(usedImportKeys, k)
	}
	slices.Sort(usedImportKeys)
	for _, mod := range usedImportKeys {
		name := moduleNames[mod]
		impName := moduleImportNames[mod]
		if name == impName {
			cb.Linef(`"%v"`, mod)
		} else {
			cb.Linef(`%v "%v"`, impName, mod)
		}
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

	cb.Linef(`var ryeStructNameLookup = map[string]string{`)
	cb.Indent++
	{
		typNames := make(map[string]string, len(data.Structs)*2)
		for _, struc := range data.Structs {
			id := struc.Name
			if !IdentExprIsExported(id.Expr) || IdentIsInternal(ctx, id) {
				continue
			}
			var nameNoMod string
			switch expr := id.Expr.(type) {
			case *ast.Ident:
				nameNoMod = expr.Name
			case *ast.StarExpr:
				id, ok := expr.X.(*ast.Ident)
				if !ok {
					continue
				}
				nameNoMod = "*"+id.Name
			case *ast.SelectorExpr:
				nameNoMod = expr.Sel.Name
			default:
				continue
			}
			typNames[id.File.ModulePath + "." + nameNoMod] = id.RyeName
			typNames[id.File.ModulePath + ".*" + nameNoMod] = "ptr-" + id.RyeName
		}
		strucNameKeys := make([]string, 0, len(typNames))
		for k := range typNames {
			strucNameKeys = append(strucNameKeys, k)
		}
		slices.Sort(strucNameKeys)
		for _, k := range strucNameKeys {
			cb.Linef(`"%v": "%v",`, k, typNames[k])
		}
	}
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

	bindingFuncKeys := make([]string, 0, len(bindingFuncs))
	for k := range bindingFuncs {
		bindingFuncKeys = append(bindingFuncKeys, k)
	}
	slices.Sort(bindingFuncKeys)
	for _, k := range bindingFuncKeys {
		bind := bindingFuncs[k]
		cb.Linef(`"%v": {`, bind.FullName())
		cb.Indent++
		cb.Linef(`Doc: "%v",`, bind.Doc)
		cb.Linef(`Argsn: %v,`, bind.Argsn)
		cb.Linef(`Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {`)
		cb.Indent++
		rep := strings.NewReplacer(`((RYEGEN:FUNCNAME))`, bind.FullName())
		cb.Write(rep.Replace(bind.Body))
		cb.Indent--
		cb.Linef(`},`)
		cb.Indent--
		cb.Linef(`},`)
	}

	cb.Indent--
	cb.Linef(`}`)

	code, err := format.Source([]byte(cb.String()))
	if err != nil {
		fmt.Println("gofmt:", err)
		os.Exit(1)
	}
	//code := []byte(cb.String())

	log.Printf("Generated bindings containing %v functions in %v", len(bindingFuncs), time.Since(startTime))

	if err := os.WriteFile(outFile, code, 0666); err != nil {
		panic(err)
	}
	log.Printf("Wrote bindings to %v", outFile)
}
