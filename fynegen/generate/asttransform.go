package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

type File struct {
	Name          string
	ModuleName    string
	ModulePath    string
	ImportsByName map[string]*File
	ImportsByPath map[string]*File
}

func (f *File) AddImport(imp *File) {
	f.ImportsByName[imp.ModuleName] = imp
	f.ImportsByPath[imp.ModulePath] = imp
}

func IdentExprIsExported(expr ast.Expr) bool {
	switch expr := expr.(type) {
	case *ast.Ident:
		return token.IsExported(expr.Name)
	case *ast.StarExpr:
		return IdentExprIsExported(expr.X)
	case *ast.SelectorExpr:
		return IdentExprIsExported(expr.Sel)
	case *ast.ArrayType:
		return IdentExprIsExported(expr.Elt)
	case *ast.Ellipsis:
		return IdentExprIsExported(expr.Elt)
	default:
		return false
	}
}

func ModulePathIsInternal(ctx *Context, p string) bool {
	if _, ok := ctx.ModuleNames[p]; !ok {
		return false
	}
	sp := strings.Split(p, "/")
	for _, elem := range sp {
		if elem == "internal" {
			return true
		}
	}
	return false
}

func IdentIsInternal(ctx *Context, id Ident) bool {
	if id.File == nil {
		return false
	}
	for _, imp := range id.UsedImports {
		if ModulePathIsInternal(ctx, imp.ModulePath) {
			return true
		}
	}
	return false
}

type Ident struct {
	Expr        ast.Expr
	GoName      string
	RyeName     string
	IsEllipsis  bool
	File        *File
	UsedImports []*File
}

func identExprToRyeName(ctx *Context, file *File, expr ast.Expr) (string, error) {
	// From https://github.com/refaktor/rye/blob/main/loader/loader.go#L444
	// WORD          <-  LETTER LETTERORNUM* / NORMOPWORDS
	// LETTER        <-  < [a-zA-Z^(` + "`" + `] >
	// LETTERORNUM   <-  < [a-zA-Z0-9-?=.\\!_+<>\]*()] >
	// NORMOPWORDS   <-  < ("_"[<>*+-=/]) >

	switch expr := expr.(type) {
	case *ast.Ident:
		res := strcase.ToKebab(expr.Name)
		if ast.IsExported(expr.Name) && file != nil {
			res = strcase.ToKebab(ctx.ModuleNames[file.ModulePath]) + "-" + res
		}
		return res, nil
	case *ast.StarExpr:
		res, err := identExprToRyeName(ctx, file, expr.X)
		return "ptr-" + res, err
	case *ast.SelectorExpr:
		mod, ok := expr.X.(*ast.Ident)
		if !ok {
			panic("expected ast.SelectorExpr.X to be of type *ast.Ident")
		}
		f, ok := file.ImportsByName[mod.Name]
		if !ok {
			return "", fmt.Errorf("module %v imported by %v not found", mod.Name, file.Name)
		}
		return identExprToRyeName(ctx, f, expr.Sel)
	case *ast.ArrayType:
		res, err := identExprToRyeName(ctx, file, expr.Elt)
		return "arr-" + res, err
	case *ast.Ellipsis:
		res, err := identExprToRyeName(ctx, file, expr.Elt)
		return "arr-" + res, err
	case *ast.FuncType:
		if expr.TypeParams != nil {
			return "", errors.New("generic functions as parameters are unsupported")
		}

		var res strings.Builder

		params, _, err := ParamsToIdents(ctx, file, expr.Params)
		if err != nil {
			return "", err
		}
		res.WriteString("func(")
		for i, v := range params {
			if i != 0 {
				res.WriteString("_")
			}
			res.WriteString(v.Type.RyeName)
		}
		res.WriteString(")")

		if expr.Results != nil {
			results, _, err := ParamsToIdents(ctx, file, expr.Results)
			if err != nil {
				return "", err
			}
			res.WriteString("_(")
			for i, v := range results {
				if i != 0 {
					res.WriteString("_")
				}
				res.WriteString(v.Type.RyeName)
			}
			res.WriteString(")")
		}

		return res.String(), nil
	case *ast.MapType:
		key, err := identExprToRyeName(ctx, file, expr.Key)
		if err != nil {
			return "", err
		}
		val, err := identExprToRyeName(ctx, file, expr.Value)
		if err != nil {
			return "", err
		}
		return "map(" + key + ")" + val, nil
	case *ast.InterfaceType:
		if len(expr.Methods.List) == 0 {
			return "any", nil
		}
		return "", errors.New("non-empty inline interfaces not supported")
	default:
		return "", errors.New("invalid identifier expression type " + reflect.TypeOf(expr).String())
	}
}

func identExprToGoName(ctx *Context, file *File, expr ast.Expr) (ident string, usedImports []*File, err error) {
	switch expr := expr.(type) {
	case *ast.Ident:
		if file != nil {
			if ast.IsExported(expr.Name) {
				mod, ok := ctx.ModuleNames[file.ModulePath]
				if !ok {
					return "", nil, fmt.Errorf("unknown module path %v", file.ModulePath)
				}
				return mod + "." + expr.Name, []*File{file}, nil
			}
		}
		if ast.IsExported(expr.Name) {
			return expr.Name, []*File{file}, nil
		} else {
			return expr.Name, nil, nil
		}
	case *ast.StarExpr:
		res, imps, err := identExprToGoName(ctx, file, expr.X)
		return "*" + res, imps, err
	case *ast.SelectorExpr:
		mod, ok := expr.X.(*ast.Ident)
		if !ok {
			panic("expected ast.SelectorExpr.X to be of type *ast.Ident")
		}
		f, ok := file.ImportsByName[mod.Name]
		if !ok {
			return "", nil, fmt.Errorf("module %v imported by %v not found", mod.Name, file.Name)
		}
		res, imps, err := identExprToGoName(ctx, f, expr.Sel)
		return res, imps, err
	case *ast.ArrayType:
		res, imps, err := identExprToGoName(ctx, file, expr.Elt)
		return "[]" + res, imps, err
	case *ast.Ellipsis:
		res, imps, err := identExprToGoName(ctx, file, expr.Elt)
		return "[]" + res, imps, err
	case *ast.FuncType:
		if expr.TypeParams != nil {
			return "", nil, errors.New("generic functions as parameters are unsupported")
		}

		var res strings.Builder

		params, imps, err := ParamsToIdents(ctx, file, expr.Params)
		if err != nil {
			return "", nil, err
		}
		res.WriteString("func(")
		for i, v := range params {
			if i != 0 {
				res.WriteString(", ")
			}
			res.WriteString(v.Type.GoName)
		}
		res.WriteString(")")

		if expr.Results != nil {
			results, impsR, err := ParamsToIdents(ctx, file, expr.Results)
			if err != nil {
				return "", nil, err
			}
			imps = append(imps, impsR...)
			res.WriteString(" (")
			for i, v := range results {
				if i != 0 {
					res.WriteString(", ")
				}
				res.WriteString(v.Type.GoName)
			}
			res.WriteString(")")
		}

		return res.String(), imps, nil
	case *ast.MapType:
		key, impsK, err := identExprToGoName(ctx, file, expr.Key)
		if err != nil {
			return "", nil, err
		}
		val, impsV, err := identExprToGoName(ctx, file, expr.Value)
		if err != nil {
			return "", nil, err
		}
		return "map[" + key + "]" + val, append(impsK, impsV...), nil
	case *ast.InterfaceType:
		if len(expr.Methods.List) == 0 {
			return "any", nil, nil
		}
		return "", nil, errors.New("non-empty inline interfaces not supported")
	default:
		return "", nil, errors.New("invalid identifier expression type " + reflect.TypeOf(expr).String())
	}
}

func NewIdent(ctx *Context, file *File, expr ast.Expr) (Ident, error) {
	goName, imps, err := identExprToGoName(ctx, file, expr)
	if err != nil {
		return Ident{}, err
	}
	ryeName, err := identExprToRyeName(ctx, file, expr)
	if err != nil {
		return Ident{}, err
	}
	isEllipsis := false
	if _, ok := expr.(*ast.Ellipsis); ok {
		isEllipsis = true
	}
	return Ident{
		Expr:        expr,
		GoName:      goName,
		RyeName:     ryeName,
		IsEllipsis:  isEllipsis,
		File:        file,
		UsedImports: imps,
	}, nil
}

type Func struct {
	Name    Ident
	Recv    *Ident // non-nil for methods
	Params  []NamedIdent
	Results []NamedIdent
	File    *File
}

func (fn *Func) String() string {
	var b strings.Builder
	if fn.Recv != nil {
		b.WriteString("(")
		b.WriteString(fn.Recv.GoName)
		//b.WriteString("/")
		//b.WriteString(fn.Recv.RyeName)
		b.WriteString(") ")
	}
	b.WriteString(fn.Name.GoName)
	//b.WriteString("/")
	//b.WriteString(fn.Name.RyeName)
	b.WriteString(" ")
	b.WriteString("(")
	for i, v := range fn.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(v.Type.GoName)
		//b.WriteString("/")
		//b.WriteString(v.RyeName)
	}
	b.WriteString(") -> (")
	for i, v := range fn.Results {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(v.Type.GoName)
		//b.WriteString("/")
		//b.WriteString(v.RyeName)
	}
	b.WriteString(")")
	return b.String()
}

func ParamsToIdents(ctx *Context, file *File, fl *ast.FieldList) (idents []NamedIdent, substImports []*File, err error) {
	var res []NamedIdent
	var substImps []*File
	for i, v := range fl.List {
		typID, err := NewIdent(ctx, file, v.Type)
		if err != nil {
			return nil, nil, err
		}
		substImps = append(substImps, typID.UsedImports...)
		if len(v.Names) > 0 {
			for _, n := range v.Names {
				nameID, err := NewIdent(ctx, nil, n)
				if err != nil {
					return nil, nil, err
				}
				res = append(res, NamedIdent{
					Name: nameID,
					Type: typID,
				})
			}
		} else {
			var shorthand string
			if typID.GoName == "error" && i == len(fl.List)-1 {
				shorthand = "err"
			} else {
				shorthand = strconv.Itoa(i + 1)
			}
			nameID, err := NewIdent(ctx, nil, &ast.Ident{Name: shorthand})
			if err != nil {
				return nil, nil, err
			}
			res = append(res, NamedIdent{
				Name: nameID,
				Type: typID,
			})
		}
	}
	return res, substImps, nil
}

func FuncFromGoFuncDecl(ctx *Context, file *File, fd *ast.FuncDecl) (*Func, error) {
	var err error
	res := &Func{
		File: file,
	}
	if fd.Recv == nil {
		res.Name, err = NewIdent(ctx, file, fd.Name)
		if err != nil {
			return nil, err
		}
	} else {
		res.Name, err = NewIdent(ctx, nil, fd.Name)
		if err != nil {
			return nil, err
		}
		if len(fd.Recv.List) != 1 {
			panic("expected exactly one receiver in method")
		}
		id, err := NewIdent(ctx, file, fd.Recv.List[0].Type)
		if err != nil {
			return nil, err
		}
		res.Recv = &id
	}
	fn := fd.Type
	{
		ids, _, err := ParamsToIdents(ctx, file, fn.Params)
		if err != nil {
			return nil, err
		}
		res.Params = ids
	}
	if fn.Results != nil {
		ids, _, err := ParamsToIdents(ctx, file, fn.Results)
		if err != nil {
			return nil, err
		}
		res.Results = ids
	}
	return res, nil
}

type NamedIdent struct {
	Name Ident
	Type Ident
}

type Struct struct {
	Name     Ident
	Fields   []NamedIdent
	Methods  map[string]*Func
	Inherits []Ident
}

func NewStruct(ctx *Context, file *File, name *ast.Ident, structTyp *ast.StructType) (*Struct, error) {
	res := &Struct{
		Methods: make(map[string]*Func),
	}
	{
		id, err := NewIdent(ctx, file, name)
		if err != nil {
			return nil, err
		}
		res.Name = id
	}
	for _, f := range structTyp.Fields.List {
		if len(f.Names) > 0 {
			typID, err := NewIdent(ctx, file, f.Type)
			if err != nil {
				return nil, err
			}

			for _, name := range f.Names {
				if !name.IsExported() {
					continue
				}
				nameID, err := NewIdent(ctx, nil, name)
				if err != nil {
					return nil, err
				}
				res.Fields = append(res.Fields, NamedIdent{
					Name: nameID,
					Type: typID,
				})
			}
		} else {
			typ := f.Type
			if se, ok := f.Type.(*ast.StarExpr); ok {
				typ = se.X
			}
			if !IdentExprIsExported(typ) {
				continue
			}
			id, err := NewIdent(ctx, file, typ)
			if err != nil {
				return nil, err
			}
			res.Inherits = append(res.Inherits, id)
		}
	}
	return res, nil
}

type Interface struct {
	Name     Ident
	Funcs    []*Func
	Inherits []Ident
}

func funcFromInterfaceField(ctx *Context, file *File, ifaceIdent Ident, f *ast.Field) (*Func, error) {
	var err error
	res := &Func{
		File: file,
	}
	if len(f.Names) != 1 {
		panic("expected interface method to have 1 name")
	}
	res.Name, err = NewIdent(ctx, nil, f.Names[0])
	if err != nil {
		return nil, err
	}
	res.Recv = &ifaceIdent
	fn, ok := f.Type.(*ast.FuncType)
	if !ok {
		panic("expected method type to be of type *ast.FuncType")
	}
	{
		ids, _, err := ParamsToIdents(ctx, file, fn.Params)
		if err != nil {
			return nil, err
		}
		res.Params = ids
	}
	if fn.Results != nil {
		ids, _, err := ParamsToIdents(ctx, file, fn.Results)
		if err != nil {
			return nil, err
		}
		res.Results = ids
	}
	return res, nil
}

func NewInterface(ctx *Context, file *File, name *ast.Ident, ifaceTyp *ast.InterfaceType) (*Interface, error) {
	res := &Interface{}
	{
		id, err := NewIdent(ctx, file, name)
		if err != nil {
			return nil, err
		}
		res.Name = id
	}
	for _, f := range ifaceTyp.Methods.List {
		switch ft := f.Type.(type) {
		case *ast.FuncType:
			fn, err := funcFromInterfaceField(ctx, file, res.Name, f)
			if err != nil {
				fmt.Println("i2fs:", res.Name.GoName+":", err)
				continue
			}
			res.Funcs = append(res.Funcs, fn)
		case *ast.Ident:
			id, err := NewIdent(ctx, file, ft)
			if err != nil {
				return nil, err
			}
			res.Inherits = append(res.Inherits, id)
		default:
			var s strings.Builder
			format.Node(&s, fset, f.Type)
			return nil, errors.New("invalid interface field " + s.String())
		}
	}
	return res, nil
}

func FuncGoIdent(fn *Func) string {
	res := fn.Name.GoName
	if fn.Recv != nil {
		_, recvIsPtr := fn.Recv.Expr.(*ast.StarExpr)
		recv := fn.Recv.GoName
		if recvIsPtr {
			recv = "(" + recv + ")"
		}
		res = recv + "." + res
	}
	return res
}

type Data struct {
	Funcs      map[string]*Func
	Interfaces map[string]*Interface
	Structs    map[string]*Struct
	Typedefs   map[string]Ident
	Values     map[string]NamedIdent // consts and vars
}

func (d *Data) AddFile(ctx *Context, f *ast.File, fName string, modulePath string, moduleNames map[string]string) error {
	file := &File{
		Name:          fName,
		ModuleName:    f.Name.Name,
		ModulePath:    modulePath,
		ImportsByName: make(map[string]*File),
		ImportsByPath: make(map[string]*File),
	}

	for _, imp := range f.Imports {
		var name string
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return err
		}
		if imp.Name != nil {
			name = imp.Name.Name
		} else {
			if v, ok := moduleNames[path]; ok {
				name = v
			} else {
				pathElems := strings.Split(path, "/")
				if len(pathElems) == 0 {
					return fmt.Errorf("unable to get module name: invalid import path %v", path)
				}
				if strings.Contains(pathElems[0], ".") {
					// not part of go std, should have been in moduleNames
					return fmt.Errorf("unable to get module name: unknown package %v", path)
				}
				// go std module
				name = pathElems[len(pathElems)-1]
			}
		}
		file.AddImport(&File{
			ModuleName:    name,
			ModulePath:    path,
			ImportsByName: make(map[string]*File),
			ImportsByPath: make(map[string]*File),
		})
	}

	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if !decl.Name.IsExported() {
				continue
			}
			if decl.Recv != nil {
				if len(decl.Recv.List) != 1 {
					panic("expected exactly one receiver in method")
				}
				if !IdentExprIsExported(decl.Recv.List[0].Type) {
					continue
				}
			}
			fn, err := FuncFromGoFuncDecl(ctx, file, decl)
			if err != nil {
				return err
			}
			d.Funcs[FuncGoIdent(fn)] = fn
		case *ast.GenDecl:
			if decl.Tok == token.CONST || decl.Tok == token.VAR {
				var typ *Ident
				for _, spec := range decl.Specs {
					if valSpec, ok := spec.(*ast.ValueSpec); ok {
						if valSpec.Type != nil {
							newTyp, err := NewIdent(ctx, file, valSpec.Type)
							if err != nil {
								return err
							}
							typ = &newTyp
						}
						if typ == nil {
							continue
						}
						for _, specName := range valSpec.Names {
							if !specName.IsExported() {
								continue
							}
							name, err := NewIdent(ctx, file, specName)
							if err != nil {
								return err
							}
							d.Values[name.GoName] = NamedIdent{
								Type: *typ,
								Name: name,
							}
						}
					}
				}
			} else if decl.Tok == token.TYPE {
				if typeSpec, ok := decl.Specs[0].(*ast.TypeSpec); ok {
					if !typeSpec.Name.IsExported() {
						continue
					}
					switch typ := typeSpec.Type.(type) {
					case *ast.InterfaceType:
						iface, err := NewInterface(ctx, file, typeSpec.Name, typ)
						if err != nil {
							return err
						}
						d.Interfaces[iface.Name.GoName] = iface
					case *ast.StructType:
						struc, err := NewStruct(ctx, file, typeSpec.Name, typ)
						if err != nil {
							return err
						}
						d.Structs[struc.Name.GoName] = struc
					default:
						name, err := NewIdent(ctx, file, typeSpec.Name)
						if err != nil {
							return err
						}
						id, err := NewIdent(ctx, file, typ)
						if err != nil {
							return err
						}
						d.Typedefs[name.GoName] = id
					}
				}
			}
		}
	}
	return nil
}

// Resolves interface, struct, and method inheritance
func (d *Data) ResolveInheritancesAndMethods(ctx *Context) error {
	var resolveInheritedIfaces func(iface *Interface) error
	resolveInheritedIfaces = func(iface *Interface) error {
		for _, inh := range iface.Inherits {
			inhIface, exists := d.Interfaces[inh.GoName]
			if !exists {
				fmt.Println(errors.New("cannot resolve interface inheritance " + inh.GoName + " in " + iface.Name.GoName + ": does not exist"))
				continue
				//return
			}
			if err := resolveInheritedIfaces(inhIface); err != nil {
				return err
			}
			iface.Funcs = append(iface.Funcs, inhIface.Funcs...)
			iface.Inherits = nil
		}
		return nil
	}
	for _, iface := range d.Interfaces {
		if err := resolveInheritedIfaces(iface); err != nil {
			return err
		}
	}

	for _, fn := range d.Funcs {
		if fn.Recv == nil {
			continue
		}
		var recv Ident
		if expr, ok := fn.Recv.Expr.(*ast.StarExpr); ok {
			var err error
			recv, err = NewIdent(ctx, fn.Recv.File, expr.X)
			if err != nil {
				return err
			}
		} else {
			recv = *fn.Recv
		}
		struc, ok := d.Structs[recv.GoName]
		if !ok {
			fmt.Println(errors.New("function " + FuncGoIdent(fn) + " has unknown receiver struct " + recv.GoName))
			continue
			//return
		}
		struc.Methods[fn.Name.GoName] = fn
	}

	var resolveInheritedStructs func(struc *Struct) error
	resolveInheritedStructs = func(struc *Struct) error {
		for _, inh := range struc.Inherits {
			inhStruc, exists := d.Structs[inh.GoName]
			if !exists {
				fmt.Println(errors.New("cannot resolve struct inheritance " + inh.GoName + " in " + struc.Name.GoName + ": does not exist"))
				continue
				//return
			}
			if err := resolveInheritedStructs(inhStruc); err != nil {
				return err
			}
			struc.Fields = append(struc.Fields, inhStruc.Fields...)
			for name, meth := range inhStruc.Methods {
				if _, exists := struc.Methods[name]; !exists {
					m := &Func{
						Name:    meth.Name,
						Recv:    &struc.Name,
						Params:  slices.Clone(meth.Params),
						Results: slices.Clone(meth.Results),
					}

					if _, ok := meth.Recv.Expr.(*ast.StarExpr); ok {
						recv, err := NewIdent(ctx, struc.Name.File, &ast.StarExpr{X: struc.Name.Expr})
						if err != nil {
							panic(err)
						}
						m.Recv = &recv
					} else {
						m.Recv = &struc.Name
					}
					struc.Methods[name] = m

					d.Funcs[FuncGoIdent(m)] = m
				}
			}
			struc.Inherits = nil
		}
		return nil
	}
	for _, struc := range d.Structs {
		if err := resolveInheritedStructs(struc); err != nil {
			return err
		}
	}
	return nil
}
