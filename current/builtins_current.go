package current

import (
	"strings"

	"github.com/refaktor/rye-front/current/ebitengine"
	"github.com/refaktor/rye-front/current/fyne"
	"github.com/refaktor/rye-front/current/webview"
	"github.com/refaktor/rye/env"
)

var Builtins_current = map[string]*env.Builtin{}

func RegisterBuiltins(ps *env.ProgramState, builtinNames *map[string]int) {
	RegisterBuiltins2(Builtins_current, ps, "current", builtinNames)
	RegisterBuiltins2(ebitengine.Builtins_ebitengine, ps, "ebitengine", builtinNames)
	RegisterBuiltins2(fyne.Builtins_fyne, ps, "fyne", builtinNames)
	RegisterBuiltins2(webview.Builtins_webview, ps, "webview", builtinNames)
}

// TODO -- move these two into main rye repo and import and call

func RegisterBuiltins2(builtins map[string]*env.Builtin, ps *env.ProgramState, name string, builtinNames *map[string]int) {
	bn := *builtinNames
	bn[name] = len(builtins)
	for k, v := range builtins {
		bu := env.NewBuiltin(v.Fn, v.Argsn, v.AcceptFailure, v.Pure, v.Doc)
		registerBuiltin(ps, k, *bu)
	}
}

func registerBuiltin(ps *env.ProgramState, word string, builtin env.Builtin) {
	// indexWord
	// TODO -- this with string separator is a temporary way of how we define generic builtins
	// in future a map will probably not be a map but an array and builtin will also support the Kind value

	idxk := 0
	if strings.Index(word, "//") > 0 {
		temp := strings.Split(word, "//")
		word = temp[1]
		idxk = ps.Idx.IndexWord(temp[0])
	}
	idxw := ps.Idx.IndexWord(word)
	// set global word with builtin
	if idxk == 0 {
		ps.Ctx.Set(idxw, builtin)
		if builtin.Pure {
			ps.PCtx.Set(idxw, builtin)
		}

	} else {
		ps.Gen.Set(idxk, idxw, builtin)
	}
}
