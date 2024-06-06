package main

import (
	"fmt"
	"go/ast"
	"strings"
)

type Converter struct {
	Name    string
	TryConv func(data *Data, cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool
}

func ConvRyeToGo(data *Data, cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) (string, bool) {
	for _, conv := range ConvListRyeToGo {
		if conv.TryConv(data, cb, typ, inVar, outVar, makeRetArgErr) {
			return conv.Name, true
		}
	}
	return "", false
}

func ConvGoToRye(data *Data, cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) (string, bool) {
	for _, conv := range ConvListGoToRye {
		if conv.TryConv(data, cb, typ, inVar, outVar, makeRetArgErr) {
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
		TryConv: func(data *Data, cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			var elTyp Ident
			switch t := typ.Expr.(type) {
			case *ast.ArrayType:
				var err error
				elTyp, err = NewIdent(typ.File, t.Elt)
				if err != nil {
					// TODO
					panic(err)
				}
			case *ast.Ellipsis:
				var err error
				elTyp, err = NewIdent(typ.File, t.Elt)
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
			typ.MarkUsed(data)
			cb.Linef(`for i, it := range v.Series.S {`)
			cb.Indent++
			if _, found := ConvRyeToGo(
				data,
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
			typ.MarkUsed(data)
			cb.Linef(`if !ok {`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr("BlockType", "NativeType"))
			cb.Indent--
			cb.Linef(`}`)
			cb.Indent--
			cb.Linef(`case env.Integer:`)
			cb.Indent++
			cb.Linef(`if v.Value != 0 {`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr("BlockType", "NativeType"))
			cb.Indent--
			cb.Linef(`}`)
			cb.Linef(`%v = nil`, outVar)
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
		Name: "map",
		TryConv: func(data *Data, cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			var kTyp, vTyp Ident
			if t, ok := typ.Expr.(*ast.MapType); ok {
				var err error
				kTyp, err = NewIdent(typ.File, t.Key)
				if err != nil {
					// TODO
					panic(err)
				}
				vTyp, err = NewIdent(typ.File, t.Value)
				if err != nil {
					// TODO
					panic(err)
				}
			} else {
				return false
			}

			allowedTyps := []string{"BlockType", "NativeType"}
			if kTyp.GoName == "string" {
				allowedTyps = append(allowedTyps, "DictType")
			}

			convAndInsert := func(inKeyVar, inValVar string, convKey bool) bool {
				if convKey {
					cb.Linef(`var mapK %v`, kTyp.GoName)
					kTyp.MarkUsed(data)
					if _, found := ConvRyeToGo(
						data,
						cb,
						kTyp,
						inKeyVar,
						`mapK`,
						func(...string) string {
							// Force toplevel allowed types
							return makeRetArgErr(allowedTyps...)
						},
					); !found {
						return false
					}
				} else {
					cb.Linef(`mapK := %v`, inKeyVar)
				}
				cb.Linef(`var mapV %v`, vTyp.GoName)
				vTyp.MarkUsed(data)
				if _, found := ConvRyeToGo(
					data,
					cb,
					vTyp,
					inValVar,
					`mapV`,
					func(...string) string {
						// Force toplevel allowed types
						return makeRetArgErr(allowedTyps...)
					},
				); !found {
					return false
				}
				cb.Linef(`%v[mapK] = mapV`, outVar)
				return true
			}

			cb.Linef(`switch v := %v.(type) {`, inVar)
			cb.Linef(`case env.Block:`)
			cb.Indent++
			cb.Linef(`if len(v.Series.S) %% 2 != 0 {`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr(allowedTyps...))
			cb.Indent--
			cb.Linef(`}`)
			cb.Linef(`%v = make(%v, len(v.Series.S)/2)`, outVar, typ.GoName)
			typ.MarkUsed(data)
			cb.Linef(`for i := 0; i < len(v.Series.S); i += 2 {`)
			cb.Indent++
			if !convAndInsert(`v.Series.S[i+0]`, `v.Series.S[i+1]`, true) {
				return false
			}
			cb.Indent--
			cb.Linef(`}`)
			cb.Indent--
			cb.Linef(`case env.Dict:`)
			cb.Indent++
			cb.Linef(`%v = make(%v, len(v.Data))`, outVar, typ.GoName)
			typ.MarkUsed(data)
			cb.Linef(`for dictK, dictV := range v.Data {`)
			cb.Indent++
			if !convAndInsert(`dictK`, `dictV`, false) {
				return false
			}
			cb.Indent--
			cb.Linef(`}`)
			cb.Indent--
			cb.Linef(`case env.Native:`)
			cb.Indent++
			cb.Linef(`var ok bool`)
			cb.Linef(`%v, ok = v.Value.(%v)`, outVar, typ.GoName)
			typ.MarkUsed(data)
			cb.Linef(`if !ok {`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr(allowedTyps...))
			cb.Indent--
			cb.Linef(`}`)
			cb.Indent--
			cb.Linef(`case env.Integer:`)
			cb.Indent++
			cb.Linef(`if v.Value != 0 {`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr(allowedTyps...))
			cb.Indent--
			cb.Linef(`}`)
			cb.Linef(`%v = nil`, outVar)
			cb.Indent--
			cb.Linef(`default:`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr(allowedTyps...))
			cb.Indent--
			cb.Linef(`}`)

			return true
		},
	},
	{
		Name: "func",
		TryConv: func(data *Data, cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			var fnParams []NamedIdent
			var fnResults []NamedIdent
			var fnTyp string
			switch t := typ.Expr.(type) {
			case *ast.FuncType:
				var err error
				fnParams, err = ParamsToIdents(typ.File, t.Params)
				if err != nil {
					// TODO
					panic(err)
				}
				if t.Results != nil {
					fnResults, err = ParamsToIdents(typ.File, t.Results)
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
					fnTypB.WriteString(fmt.Sprintf("arg%v %v", i, param.Type.GoName))
					param.Type.MarkUsed(data)
				}
				fnTypB.WriteString(")")
				if len(fnResults) > 0 {
					fnTypB.WriteString(" (")
					for i, result := range fnResults {
						if i != 0 {
							fnTypB.WriteString(", ")
						}
						fnTypB.WriteString(result.Type.GoName)
						result.Type.MarkUsed(data)
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
					data,
					cb,
					param.Type,
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
				cb.Linef(`var res %v`, fnResults[0].Type.GoName)
				fnResults[0].Type.MarkUsed(data)
				if _, found := ConvRyeToGo(
					data,
					cb,
					fnResults[0].Type,
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
			cb.Linef(`case env.Integer:`)
			cb.Indent++
			cb.Linef(`if fn.Value != 0 {`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr("BlockType", "NativeType"))
			cb.Indent--
			cb.Linef(`}`)
			cb.Linef(`%v = nil`, outVar)
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
		TryConv: func(data *Data, cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
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
			} else if id.Name == "error" {
				ryeObj = "Error"
				ryeObjType = "ErrorType"
			} else {
				return false
			}

			cb.Linef(`if v, ok := %v.(env.%v); ok {`, inVar, ryeObj)
			cb.Indent++
			if id.Name == "bool" {
				cb.Linef(`%v = v.Value != 0`, outVar)
			} else if id.Name == "error" {
				cb.Linef(`%v = errors.New(v.Print(*ps.Idx))`, outVar)
				data.UsedImports["errors"] = struct{}{}
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
		TryConv: func(data *Data, cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			isNillable := false
			switch typ.Expr.(type) {
			case *ast.StarExpr, *ast.ArrayType:
				isNillable = true
			}
			if _, exists := data.Interfaces[typ.GoName]; exists {
				isNillable = true
			}

			cb.Linef(`switch v := %v.(type) {`, inVar)
			cb.Linef(`case env.Native:`)
			cb.Indent++
			cb.Linef(`var ok bool`)
			cb.Linef(`%v, ok = v.Value.(%v)`, outVar, typ.GoName)
			typ.MarkUsed(data)
			cb.Linef(`if !ok {`)
			cb.Indent++
			cb.Linef(`%v`, makeRetArgErr("NativeType"))
			cb.Indent--
			cb.Linef(`}`)
			cb.Indent--
			if isNillable {
				cb.Linef(`case env.Integer:`)
				cb.Indent++
				cb.Linef(`if v.Value != 0 {`)
				cb.Indent++
				cb.Linef(`%v`, makeRetArgErr("NativeType"))
				cb.Indent--
				cb.Linef(`}`)
				cb.Linef(`%v = nil`, outVar)
				cb.Indent--
			}
			cb.Linef(`default:`)
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
		Name: "array",
		TryConv: func(data *Data, cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			var elTyp Ident
			switch t := typ.Expr.(type) {
			case *ast.ArrayType:
				var err error
				elTyp, err = NewIdent(typ.File, t.Elt)
				if err != nil {
					// TODO
					panic(err)
				}
			case *ast.Ellipsis:
				var err error
				elTyp, err = NewIdent(typ.File, t.Elt)
				if err != nil {
					// TODO
					panic(err)
				}
			default:
				return false
			}

			cb.Linef(`{`)
			cb.Indent++
			cb.Linef(`items := make([]env.Object, len(%v))`, inVar)
			cb.Linef(`for i, it := range %v {`, inVar)
			cb.Indent++
			if _, found := ConvGoToRye(
				data,
				cb,
				elTyp,
				`it`,
				`items[i]`,
				nil,
			); !found {
				return false
			}
			cb.Indent--
			cb.Linef(`}`)
			cb.Linef(`%v = *env.NewBlock(*env.NewTSeries(items))`, outVar)
			cb.Indent--
			cb.Linef(`}`)

			return true
		},
	},
	{
		Name: "map",
		TryConv: func(data *Data, cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			var kTyp, vTyp Ident
			if t, ok := typ.Expr.(*ast.MapType); ok {
				var err error
				kTyp, err = NewIdent(typ.File, t.Key)
				if err != nil {
					// TODO
					panic(err)
				}
				vTyp, err = NewIdent(typ.File, t.Value)
				if err != nil {
					// TODO
					panic(err)
				}
			} else {
				return false
			}

			if kTyp.GoName != "string" {
				return false
			}

			cb.Linef(`{`)
			cb.Indent++
			cb.Linef(`data := make(map[string]any, len(%v))`, inVar)
			cb.Linef(`for mKey, mVal := range %v {`, inVar)
			cb.Indent++
			cb.Linef(`var dVal env.Object`)
			if _, found := ConvGoToRye(
				data,
				cb,
				vTyp,
				`mVal`,
				`dVal`,
				nil,
			); !found {
				return false
			}
			cb.Linef(`data[mKey] = dVal`)
			cb.Indent--
			cb.Linef(`}`)
			cb.Linef(`%v = *env.NewDict(data)`, outVar)
			cb.Indent--
			cb.Linef(`}`)

			return true
		},
	},
	{
		Name: "builtin",
		TryConv: func(data *Data, cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			id, ok := typ.Expr.(*ast.Ident)
			if !ok {
				return false
			}

			var convFmt string
			if id.Name == "int" || id.Name == "uint" ||
				id.Name == "uint8" || id.Name == "uint16" || id.Name == "uint32" || id.Name == "uint64" ||
				id.Name == "int8" || id.Name == "int16" || id.Name == "int32" || id.Name == "int64" {
				convFmt = `*env.NewInteger(int64(%v))`
			} else if id.Name == "bool" {
				convFmt = `*env.NewInteger(boolToInt64(%v))`
			} else if id.Name == "float32" || id.Name == "float64" {
				convFmt = `*env.NewDecimal(float64(%v))`
			} else if id.Name == "string" {
				convFmt = `*env.NewString(%v)`
			} else if id.Name == "error" {
				convFmt = `*env.NewError(%v.Error())`
			} else {
				return false
			}

			cb.Linef(`%v = %v`, outVar, fmt.Sprintf(convFmt, inVar))
			return true
		},
	},
	{
		Name: "native",
		TryConv: func(data *Data, cb *CodeBuilder, typ Ident, inVar, outVar string, makeRetArgErr func(allowedTypes ...string) string) bool {
			cb.Linef(`%v = *env.NewNative(ps.Idx, %v, "%v")`, outVar, inVar, typ.RyeName)
			return true
		},
	},
}
