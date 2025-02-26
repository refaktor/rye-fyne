// Add your custom builtins to this file.

package fyne_io_fyne_v2

import (
	"strings"

	"github.com/refaktor/rye/env"
)

var builtinsCustom = map[string]*env.Builtin{
	"nil": {
		Doc: "nil value for go types",
		Fn: func(ps *env.ProgramState, arg0, arg1, arg2, arg3, arg4 env.Object) env.Object {
			return *env.NewInteger(0)
		},
	},
	"kind": {
		Doc: "underlying kind of a go native",
		Fn: func(ps *env.ProgramState, arg0, arg1, arg2, arg3, arg4 env.Object) env.Object {
			nat, ok := arg0.(env.Native)
			if !ok {
				ps.FailureFlag = true
				return env.NewError("kind: arg0: expected native")
			}
			s := ps.Idx.GetWord(nat.Kind.Index)
			s = s[3 : len(s)-1]            // remove surrounding "Go()"
			s = strings.TrimPrefix(s, "*") // remove potential pointer "*"
			return *env.NewString(s)
		},
	},
	// Add your custom builtins here:
}
