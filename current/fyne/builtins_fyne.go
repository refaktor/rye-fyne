//go:build b_fyne

package fyne

// import "C"

import (
	"fmt"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var Builtins_fyne = map[string]*env.Builtin{

	"fyne-app": {
		Argsn: 0,
		Doc:   "Creates a Fyne app native",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			app1 := app.New()
			return *env.NewNative(ps.Idx, app1, "fyne-app")
		},
	},
	"fyne-app//new-window": {
		Argsn: 2,
		Doc:   "Creates new window for and app",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch title := arg1.(type) {
				case env.String:
					wind := win.Value.(fyne.App).NewWindow(title.Value)
					return *env.NewNative(ps.Idx, wind, "fyne-window")
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "fyne-app//new-window")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "fyne-app//new-window")
			}
		},
	},
	"fyne-label": {
		Argsn: 1,
		Doc:   "Creates a Fyne label widget",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch text := arg0.(type) {
			case env.String:
				win := widget.NewLabel(text.Value)
				return *env.NewNative(ps.Idx, win, "fyne-widget")
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "gtk-window//set-title")
			}
		},
	},
	"fyne-entry": {
		Argsn: 0,
		Doc:   "Creates a Fyne entry widget",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			win := widget.NewEntry()
			return *env.NewNative(ps.Idx, win, "fyne-widget")
		},
	},
	"fyne-multiline-entry": {
		Argsn: 0,
		Doc:   "Creates a Fyne multiline entry widget",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			win := widget.NewMultiLineEntry()
			return *env.NewNative(ps.Idx, win, "fyne-widget")
		},
	},
	"fyne-widget//set-text": {
		Argsn: 2,
		Doc:   "Sets text to a widget",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch widg_ := arg0.(type) {
			case env.Native:
				switch text := arg1.(type) {
				case env.String:
					switch widg := widg_.Value.(type) {
					case *widget.Entry:
						widg.SetText(text.Value)
					case *widget.Label:
						widg.SetText(text.Value)
					}
					return arg0
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "fyne-widget//set-text")
				}
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.NativeType}, "fyne-widget//set-text")
			}
		},
	},
	"fyne-widget//get-text": {
		Argsn: 1,
		Doc:   "Gets text from a widget",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch widg := arg0.(type) {
			case env.Native:
				return env.NewString(widg.Value.(*widget.Entry).Text)
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.NativeType}, "fyne-widget//get-text")
			}
		},
	},

	"fyne-container": {
		Argsn: 2,
		Doc:   "Creates Fyne container with widgets",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch layout_ := arg0.(type) {
			case env.Word:
				layout_str := ps.Idx.GetWord(layout_.Index)
				var layout_r fyne.Layout
				switch layout_str {
				case "vbox":
					layout_r = layout.NewVBoxLayout()
				case "hbox":
					layout_r = layout.NewHBoxLayout()
				default:
					return evaldo.MakeError(ps, "Non-existent layout")
				}
				switch bloc := arg1.(type) {
				case env.Block:
					items := make([]fyne.CanvasObject, bloc.Series.Len())

					for i, it := range bloc.Series.S {
						switch nat := it.(type) {
						case env.Native:
							items[i] = nat.Value.(fyne.CanvasObject)
						}
					}
					win := container.New(layout_r, items...)
					return *env.NewNative(ps.Idx, win, "fyne-container")
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.BlockType}, "fyne-container")
				}
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.WordType}, "fyne-container")
			}
		},
	},

	"fyne-button": {
		Argsn: 2,
		Doc:   "Create new Fyne button widget",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch txt := arg0.(type) {
			case env.String:
				switch fn := arg1.(type) {
				case env.Function:
					widg := widget.NewButton(txt.Value, func() {
						evaldo.CallFunction(fn, ps, nil, false, ps.Ctx)
						// return ps.Res
					})
					return *env.NewNative(ps.Idx, widg, "fyne-widget")
				case env.Block:
					widg := widget.NewButton(txt.Value, func() {
						ser := ps.Ser
						ps.Ser = fn.Series
						// fmt.Println("BEFORE")
						r := evaldo.EvalBlockInj(ps, nil, false)
						ps.Ser = ser
						// fmt.Println("AFTER")
						if r.Res != nil && r.Res.Type() == env.ErrorType {
							fmt.Println(r.Res.(*env.Error).Message)
						}
					})
					return *env.NewNative(ps.Idx, widg, "fyne-widget")
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.BlockType, env.FunctionType}, "fyne-button")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "fyne-button")
			}
		},
	},

	"fyne-window//set-content": {
		Argsn: 2,
		Doc:   "Set content of Fyne window",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch widg := arg1.(type) {
				case env.Native:
					win.Value.(fyne.Window).SetContent(widg.Value.(fyne.CanvasObject))
					return win
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.NativeType}, "fyne-window//set-content")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "fyne-window//set-content")
			}
		},
	},

	"fyne-window//show-and-run": {
		Argsn: 1,
		Doc:   "Shows Fyne window and runs event loop",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				win.Value.(fyne.Window).ShowAndRun()
				return win
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "fyne-window//show-and-run")
			}
		},
	},
}
