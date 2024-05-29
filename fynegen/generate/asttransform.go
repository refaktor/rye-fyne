package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
)

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

type Ident struct {
	Expr       ast.Expr
	GoName     string
	RyeName    string
	IsEllipsis bool
	RootPkg    string
}

func identExprToRyeName(rootPkg string, expr ast.Expr) (string, error) {
	switch expr := expr.(type) {
	case *ast.Ident:
		res := expr.Name
		if ast.IsExported(expr.Name) {
			res = strcase.ToKebab(strings.TrimPrefix(expr.Name, "New"))
			if rootPkg != "" {
				if res == "" {
					// e.g. app.New => app
					res = strcase.ToKebab(rootPkg)
				} else {
					res = strcase.ToKebab(rootPkg) + "-" + res
				}
			}
		}
		return res, nil
	case *ast.StarExpr:
		res, err := identExprToRyeName(rootPkg, expr.X)
		return res + "-ptr", err
	case *ast.SelectorExpr:
		pkg, ok := expr.X.(*ast.Ident)
		if !ok {
			panic("expected ast.SelectorExpr.X to be of type *ast.Ident")
		}
		return identExprToRyeName(pkg.Name, expr.Sel)
	case *ast.ArrayType:
		res, err := identExprToRyeName(rootPkg, expr.Elt)
		return res + "-arr", err
	case *ast.Ellipsis:
		res, err := identExprToRyeName(rootPkg, expr.Elt)
		return res + "-arr", err
	default:
		return "", errors.New("invalid identifier expression type " + reflect.TypeOf(expr).String())
	}
}

func identExprToGoName(rootPkg string, expr ast.Expr) (string, error) {
	switch expr := expr.(type) {
	case *ast.Ident:
		if ast.IsExported(expr.Name) {
			if rootPkg != "" {
				return rootPkg + "." + expr.Name, nil
			}
		}
		return expr.Name, nil
	case *ast.StarExpr:
		res, err := identExprToGoName(rootPkg, expr.X)
		return "*" + res, err
	case *ast.SelectorExpr:
		pkg, ok := expr.X.(*ast.Ident)
		if !ok {
			panic("expected ast.SelectorExpr.X to be of type *ast.Ident")
		}
		return identExprToGoName(pkg.Name, expr.Sel)
	case *ast.ArrayType:
		res, err := identExprToGoName(rootPkg, expr.Elt)
		return "[]" + res, err
	case *ast.Ellipsis:
		res, err := identExprToGoName(rootPkg, expr.Elt)
		return "[]" + res, err
	default:
		return "", errors.New("invalid identifier expression type " + reflect.TypeOf(expr).String())
	}
}

func NewIdent(rootPkg string, expr ast.Expr) (Ident, error) {
	goName, err := identExprToGoName(rootPkg, expr)
	if err != nil {
		return Ident{}, err
	}
	ryeName, err := identExprToRyeName(rootPkg, expr)
	if err != nil {
		return Ident{}, err
	}
	isEllipsis := false
	if _, ok := expr.(*ast.Ellipsis); ok {
		isEllipsis = true
	}
	return Ident{
		Expr:       expr,
		GoName:     goName,
		RyeName:    ryeName,
		IsEllipsis: isEllipsis,
		RootPkg:    rootPkg,
	}, nil
}

type Func struct {
	Name    Ident
	Recv    *Ident // non-nil for methods
	Params  []Ident
	Results []Ident
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
		b.WriteString(v.GoName)
		//b.WriteString("/")
		//b.WriteString(v.RyeName)
	}
	b.WriteString(") -> (")
	for i, v := range fn.Results {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(v.GoName)
		//b.WriteString("/")
		//b.WriteString(v.RyeName)
	}
	b.WriteString(")")
	return b.String()
}

func paramsToIdents(rootPkg string, fl *ast.FieldList) ([]Ident, error) {
	var res []Ident
	for _, v := range fl.List {
		id, err := NewIdent(rootPkg, v.Type)
		if err != nil {
			return nil, err
		}
		{
			n := 1
			// e.g. func Max(x, y float32)
			if len(v.Names) > 1 {
				n = len(v.Names)
			}
			for i := 0; i < n; i++ {
				res = append(res, id)
			}
		}
	}
	return res, nil
}

func FuncFromGoFuncDecl(rootPkg string, fd *ast.FuncDecl) (*Func, error) {
	var err error
	res := &Func{}
	if fd.Recv == nil {
		res.Name, err = NewIdent(rootPkg, fd.Name)
		if err != nil {
			return nil, err
		}
	} else {
		res.Name, err = NewIdent("", fd.Name)
		if err != nil {
			return nil, err
		}
		if len(fd.Recv.List) != 1 {
			panic("expected exactly one receiver in method")
		}
		id, err := NewIdent(rootPkg, fd.Recv.List[0].Type)
		if err != nil {
			return nil, err
		}
		res.Recv = &id
	}
	fn := fd.Type
	{
		ids, err := paramsToIdents(rootPkg, fn.Params)
		if err != nil {
			return nil, err
		}
		res.Params = ids
	}
	if fn.Results != nil {
		ids, err := paramsToIdents(rootPkg, fn.Results)
		if err != nil {
			return nil, err
		}
		res.Results = ids
	}
	return res, nil
}

type Struct struct {
	Name     Ident
	Fields   []Ident
	Inherits []Ident
}

type Interface struct {
	Name     Ident
	Funcs    []*Func
	Inherits []Ident
}

func funcFromInterfaceField(rootPkg string, ifaceIdent Ident, f *ast.Field) (*Func, error) {
	var err error
	res := &Func{}
	if len(f.Names) != 1 {
		panic("expected method to have 1 name")
	}
	// interface field is method => not scoped => no namespace / root pkg
	res.Name, err = NewIdent("", f.Names[0])
	if err != nil {
		return nil, err
	}
	res.Recv = &ifaceIdent
	fn, ok := f.Type.(*ast.FuncType)
	if !ok {
		panic("expected method type to be of type *ast.FuncType")
	}
	{
		ids, err := paramsToIdents(rootPkg, fn.Params)
		if err != nil {
			return nil, err
		}
		res.Params = ids
	}
	if fn.Results != nil {
		ids, err := paramsToIdents(rootPkg, fn.Results)
		if err != nil {
			return nil, err
		}
		res.Results = ids
	}
	return res, nil
}

func NewInterface(rootPkg string, name *ast.Ident, ifaceTyp *ast.InterfaceType) (*Interface, error) {
	res := &Interface{}
	{
		id, err := NewIdent(rootPkg, name)
		if err != nil {
			return nil, err
		}
		res.Name = id
	}
	for _, f := range ifaceTyp.Methods.List {
		switch ft := f.Type.(type) {
		case *ast.FuncType:
			fn, err := funcFromInterfaceField(rootPkg, res.Name, f)
			if err != nil {
				fmt.Println("i2fs:", err)
				continue
			}
			res.Funcs = append(res.Funcs, fn)
		case *ast.Ident:
			id, err := NewIdent(rootPkg, ft)
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

func FuncRyeIdent(fn *Func) string {
	res := fn.Name.RyeName
	if fn.Recv != nil {
		res = fn.Recv.RyeName + "//" + res
	}
	return res
}

type Data struct {
	Funcs      map[string]*Func
	Interfaces map[string]*Interface
	Structs    map[string]*Struct // TODO
}

func NewData() *Data {
	return &Data{
		Funcs:      make(map[string]*Func),
		Interfaces: make(map[string]*Interface),
		Structs:    make(map[string]*Struct),
	}
}

func (d *Data) AddFile(f *ast.File) error {
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
			fn, err := FuncFromGoFuncDecl(f.Name.Name, decl)
			if err != nil {
				return err
			}
			d.Funcs[FuncGoIdent(fn)] = fn
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				if typeSpec, ok := decl.Specs[0].(*ast.TypeSpec); ok {
					if !typeSpec.Name.IsExported() {
						continue
					}
					if ifaceTyp, ok := typeSpec.Type.(*ast.InterfaceType); ok {
						/*if err := ast.Print(fset, it.Methods.List); err != nil {
							panic(err)
						}*/
						iface, err := NewInterface(f.Name.Name, typeSpec.Name, ifaceTyp)
						if err != nil {
							return err
						}
						d.Interfaces[iface.Name.GoName] = iface
					}
				}
			}
		}
	}
	return nil
}

// Resolves interface and struct inheritance
func (d *Data) ResolveInheritances() error {
	var resolveInheritedIfaces func(iface *Interface) error
	resolveInheritedIfaces = func(iface *Interface) error {
		for _, inh := range iface.Inherits {
			inhIface, exists := d.Interfaces[inh.GoName]
			if !exists {
				return errors.New("cannot resolve interface inheritance " + inh.GoName + " in " + iface.Name.GoName + ": does not exist")
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
	return nil
}
