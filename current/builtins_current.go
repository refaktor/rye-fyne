package current

import (
	"github.com/refaktor/rye-front/current/ebitengine"
	"github.com/refaktor/rye-front/current/fyne"
	"github.com/refaktor/rye-front/current/webview"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

var Builtins_current = map[string]*env.Builtin{}

func RegisterBuiltins(ps *env.ProgramState) {
	evaldo.RegisterBuiltinsInContext(Builtins_current, ps, "current")
	evaldo.RegisterBuiltinsInContext(ebitengine.Builtins_ebitengine, ps, "ebitengine")
	evaldo.RegisterBuiltinsInContext(fyne.Builtins_fyne, ps, "fyne")
	evaldo.RegisterBuiltinsInContext(webview.Builtins_webview, ps, "webview")
}
