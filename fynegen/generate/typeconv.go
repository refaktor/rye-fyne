package main

import (
	"fmt"
	"go/ast"
	"strings"
)

type Converter struct {
	Name    string
	TryConv func(cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool
}

func ConvRyeToGo(cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) (string, bool) {
	for _, conv := range ConvListRyeToGo {
		if conv.TryConv(cb, typ, inVar, outVar, makeRetArgErr) {
			return conv.Name, true
		}
	}
	return "", false
}

func ConvGoToRye(cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) (string, bool) {
	for _, conv := range convListGoToRye {
		if conv.TryConv(cb, typ, inVar, outVar, makeRetArgErr) {
			return conv.Name, true
		}
	}
	return "", false
}

// If conversion lists are declared directly, the compiler falsely complains of an initialization cycle.
var ConvListRyeToGo []Converter
var ConvListGoToRye []Converter

func init() {
	ConvListRyeToGo = convListRyeToGo
	ConvListGoToRye = convListGoToRye
}

var convListRyeToGo = []Converter{
	{
		Name: "array",
		TryConv: func(cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			var elTyp Ident
			switch t := typ.Expr.(type) {
			case *ast.ArrayType:
				var err error
				elTyp, err = NewIdent(typ.RootPkg, t.Elt)
				if err != nil {
					// TODO
					panic(err)
				}
			case *ast.Ellipsis:
				var err error
				elTyp, err = NewIdent(typ.RootPkg, t.Elt)
				if err != nil {
					// TODO
					panic(err)
				}
			default:
				return false
			}

			cb.Linef(`switch v := %v.(type) {`, inVar)
			cb.Linef(`case env.Block:`)
			cb.Indent++
			cb.Linef(`%v = make(%v, len(v.Series.S))`, outVar, typ.GoName)
			cb.Linef(`for i, it := range v.Series.S {`)
			cb.Indent++
			if _, found := ConvRyeToGo(
				cb,
				elTyp,
				`it`,
				fmt.Sprintf(`%v[i]`, outVar),
				func(...string) string {
					// Force toplevel allowed types
					return makeRetArgErr("BlockType", "NativeType")
				},
			); !found {
				return false
			}
			cb.Indent--
			cb.Linef(`}`)
			cb.Indent--
			cb.Linef(`case env.Native:`)
			cb.Indent++
			cb.Linef(`var ok bool`)
			cb.Linef(`%v, ok = v.Value.(%v)`, outVar, typ.GoName)
			cb.Linef(`if !ok {`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr("BlockType", "NativeType"))
			cb.Indent--
			cb.Linef(`}`)
			cb.Indent--
			cb.Linef(`default:`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr("BlockType", "NativeType"))
			cb.Indent--
			cb.Linef(`}`)

			return true
		},
	},
	{
		Name: "func",
		TryConv: func(cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			var fnParams []Ident
			var fnResults []Ident
			var fnTyp string
			switch t := typ.Expr.(type) {
			case *ast.FuncType:
				var err error
				fnParams, err = ParamsToIdents(typ.RootPkg, t.Params)
				if err != nil {
					// TODO
					panic(err)
				}
				if t.Results != nil {
					fnResults, err = ParamsToIdents(typ.RootPkg, t.Results)
					if err != nil {
						// TODO
						panic(err)
					}
				}
				if len(fnParams) > 4 || len(fnParams) == 3 {
					// TODO
					//panic("cannot have function as argument with more than 4 or exactly 3 parameters")
					return false
				}
				if len(fnResults) > 1 {
					// TODO
					//panic("cannot have function as argument with more than 1 result")
					return false
				}
				var fnTypB strings.Builder
				fnTypB.WriteString("func(")
				for i, param := range fnParams {
					if i != 0 {
						fnTypB.WriteString(", ")
					}
					fnTypB.WriteString(fmt.Sprintf("arg%v %v", i, param.GoName))
				}
				fnTypB.WriteString(")")
				if len(fnResults) > 0 {
					fnTypB.WriteString(" (")
					for i, result := range fnResults {
						if i != 0 {
							fnTypB.WriteString(", ")
						}
						fnTypB.WriteString(result.GoName)
					}
					fnTypB.WriteString(")")
				}
				fnTyp = fnTypB.String()
			default:
				return false
			}

			cb.Linef(`switch fn := %v.(type) {`, inVar)
			cb.Linef(`case env.Function:`)
			cb.Indent++
			cb.Linef(`if fn.Argsn != %v {`, len(fnParams))
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr("FunctionType"))
			cb.Indent--
			cb.Linef(`}`)
			cb.Linef(`%v = %v {`, outVar, fnTyp)
			cb.Indent++
			var argVals strings.Builder
			for i := range fnParams {
				if i != 0 {
					argVals.WriteString(", ")
				}
				argVals.WriteString(fmt.Sprintf("arg%vVal", i))
			}
			if len(fnParams) == 0 {
				argVals.WriteString("nil")
			} else {
				cb.Linef(`var %v env.Object`, argVals.String())
			}
			for i, param := range fnParams {
				if _, found := ConvGoToRye(
					cb,
					param,
					fmt.Sprintf(`arg%v`, i),
					fmt.Sprintf(`arg%vVal`, i),
					nil,
				); !found {
					return false
				}
			}
			var argsSuffix string
			if len(fnParams) > 1 {
				argsSuffix = fmt.Sprintf("Args%v", len(fnParams))
			}
			var toLeftArg string
			if len(fnParams) <= 1 {
				toLeftArg = ", false"
			}
			cb.Linef(`evaldo.CallFunction%v(fn, ps, %v%v, ps.Ctx)`, argsSuffix, argVals.String(), toLeftArg)
			if len(fnResults) > 0 {
				cb.Linef(`var res %v`, fnResults[0].GoName)
				if _, found := ConvRyeToGo(
					cb,
					fnResults[0],
					`ps.Res`,
					`res`,
					func(...string) string {
						// Can't return error from inside function
						return "// TODO: Invalid type"
					},
				); !found {
					return false
				}
				cb.Linef(`return res`)
			}
			cb.Indent--
			cb.Linef(`}`)
			cb.Indent--
			cb.Linef(`default:`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr("FunctionType"))
			cb.Indent--
			cb.Linef(`}`)

			return true
		},
	},
	{
		Name: "builtin",
		TryConv: func(cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			id, ok := typ.Expr.(*ast.Ident)
			if !ok {
				return false
			}

			var ryeObj string
			var ryeObjType string
			if id.Name == "int" || id.Name == "uint" ||
				id.Name == "uint8" || id.Name == "uint16" || id.Name == "uint32" || id.Name == "uint64" ||
				id.Name == "int8" || id.Name == "int16" || id.Name == "int32" || id.Name == "int64" ||
				id.Name == "bool" {
				ryeObj = "Integer"
				ryeObjType = "IntegerType"
			} else if id.Name == "float32" || id.Name == "float64" {
				ryeObj = "Decimal"
				ryeObjType = "DecimalType"
			} else if id.Name == "string" {
				ryeObj = "String"
				ryeObjType = "StringType"
			} else {
				return false
			}

			cb.Linef(`if v, ok := %v.(env.%v); ok {`, inVar, ryeObj)
			cb.Indent++
			if id.Name == "bool" {
				cb.Linef(`%v = v.Value != 0`, outVar)
			} else if id.Name == "error" {
				cb.Linef(`%v = errors.New(v.Print(*ps.Idx))`, outVar)
			} else {
				cb.Linef(`%v = %v(v.Value)`, outVar, id.Name)
			}
			cb.Indent--
			cb.Linef(`} else {`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr(ryeObjType))
			cb.Indent--
			cb.Linef(`}`)

			return true
		},
	},
	{
		Name: "native",
		TryConv: func(cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			cb.Linef(`if v, ok := %v.(env.Native); ok {`, inVar)
			cb.Indent++
			cb.Linef(`%v, ok = v.Value.(%v)`, outVar, typ.GoName)
			cb.Linef(`if !ok {`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr("NativeType"))
			cb.Indent--
			cb.Linef(`}`)
			cb.Indent--
			cb.Linef(`} else {`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr("NativeType"))
			cb.Indent--
			cb.Linef(`}`)

			return true
		},
	},
}

var convListGoToRye = []Converter{
	{
		Name: "builtin",
		TryConv: func(cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			id, ok := typ.Expr.(*ast.Ident)
			if !ok {
				return false
			}

			var convFunc string
			var castFunc string
			if id.Name == "int" || id.Name == "uint" ||
				id.Name == "uint8" || id.Name == "uint16" || id.Name == "uint32" || id.Name == "uint64" ||
				id.Name == "int8" || id.Name == "int16" || id.Name == "int32" || id.Name == "int64" {
				convFunc = "*env.NewInteger"
				castFunc = "int64"
			} else if id.Name == "bool" {
				convFunc = "*env.NewInteger"
				castFunc = "boolToInt64"
			} else if id.Name == "float32" || id.Name == "float64" {
				convFunc = "*env.NewDecimal"
				castFunc = "float64"
			} else if id.Name == "string" {
				convFunc = "*env.NewString"
				castFunc = "string"
			} else {
				return false
			}

			cb.Linef(`%v = %v(%v(%v))`, outVar, convFunc, castFunc, inVar)
			return true
		},
	},
	{
		Name: "native",
		TryConv: func(cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			cb.Linef(`%v = *env.NewNative(ps.Idx, %v, "%v")`, outVar, inVar, typ.RyeName)
			return true
		},
	},
}
