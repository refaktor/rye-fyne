package current

import (
	"github.com/refaktor/rye-front/current/ebitengine"
	"github.com/refaktor/rye-front/current/fyne_io_fyne_v2"
	"github.com/refaktor/rye-front/current/webview"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

var Builtins_current = map[string]*env.Builtin{}

func RegisterBuiltins(ps *env.ProgramState) {
	evaldo.RegisterBuiltins2(Builtins_current, ps, "current")
	evaldo.RegisterBuiltinsInContext(ebitengine.Builtins_ebitengine, ps, "ebitengine")
	evaldo.RegisterBuiltinsInContext(fyne_io_fyne_v2.Builtins, ps, "fyne")
	evaldo.RegisterBuiltinsInContext(webview.Builtins_webview, ps, "webview")
}
