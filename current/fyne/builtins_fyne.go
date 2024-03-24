//go:build b_fyne

package fyne

// import "C"

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var Builtins_fyne = map[string]*env.Builtin{

	"app": {
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
	"label": {
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
	"entry": {
		Argsn: 0,
		Doc:   "Creates a Fyne entry widget",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			win := widget.NewEntry()
			return *env.NewNative(ps.Idx, win, "fyne-widget")
		},
	},
	"password-entry": {
		Argsn: 1,
		Doc:   "Creates a Fyne entry password widget",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			win := widget.NewPasswordEntry()
			win.Validator = func(s string) error {
				if evaldo.CallFunction(arg0.(env.Function), ps, *env.NewString(s), false, ps.Ctx).Res.(env.Integer).Value == 0 {
					return errors.New("Text is not in correct form")
				}
				return nil
			}
			return *env.NewNative(ps.Idx, win, "fyne-widget")
		},
	},
	"multiline-entry": {
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
				switch widg.Value.(type) {
				case *widget.Entry:
					return *env.NewString(widg.Value.(*widget.Entry).Text)
				case *widget.Check:
					return *env.NewString(strconv.FormatBool(widg.Value.(*widget.Check).Checked))
				case *widget.Select:
					return *env.NewString(widg.Value.(*widget.Select).Selected)
				case *widget.RadioGroup:
					return *env.NewString(widg.Value.(*widget.RadioGroup).Selected)
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.NativeType}, "fyne-widget//get-text")
				}
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.NativeType}, "fyne-widget//get-text")
			}
		},
	},
	"checkbox": {
		Argsn: 1,
		Doc:   "Creates a Fyne check widget",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var label string
			switch arg0.(type) {
			case env.String:
				label = arg0.(env.String).Value
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "fyne-check")
			}
			win := widget.NewCheck(label, nil)
			return *env.NewNative(ps.Idx, win, "fyne-widget")
		},
	},
	"selectbox": {
		Argsn: 1,
		Doc:   "Creates a Fyne select widget",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var options []string
			switch arg0.(type) {
			case env.Block:
				{
					for _, v := range arg0.(env.Block).Series.S {
						switch v.(type) {
						case env.String:
							options = append(options, fmt.Sprintf("%v", v.(env.String).Value))
						}
					}
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.BlockType}, "fyne-check")
			}
			win := widget.NewSelect(options, nil)
			return *env.NewNative(ps.Idx, win, "fyne-widget")
		},
	},
	"radiogroup": {
		Argsn: 1,
		Doc:   "Creates a Fyne radio group widget",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var options []string
			switch arg0.(type) {
			case env.Block:
				{
					for _, v := range arg0.(env.Block).Series.S {
						switch v.(type) {
						case env.String:
							options = append(options, fmt.Sprintf("%v", v.(env.String).Value))
						}
					}
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.BlockType}, "fyne-check")
			}
			win := widget.NewRadioGroup(options, nil)
			return *env.NewNative(ps.Idx, win, "fyne-widget")
		},
	},

	"spacer": {
		Argsn: 0,
		Doc:   "Creates Fyne layout spacer",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			win := layout.NewSpacer()
			return *env.NewNative(ps.Idx, win, "fyne-layout-spacer")

		},
	},

	"container": {
		Argsn: 1,
		Doc:   "Creates Fyne container",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				items := []fyne.CanvasObject{}
				for _, it := range bloc.Series.S {
					switch nat := it.(type) {
					case env.Native:
						items = append(items, nat.Value.(fyne.CanvasObject))
					}
				}
				win := container.NewWithoutLayout(items...)
				return *env.NewNative(ps.Idx, win, "fyne-container")
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.WordType}, "fyne-container")
			}
		},
	},

	"container-vbox": {
		Argsn: 1,
		Doc:   "Creates Fyne vbox container",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch bloc := arg0.(type) {
			case env.Block:
				items := []fyne.CanvasObject{}
				for _, it := range bloc.Series.S {
					switch nat := it.(type) {
					case env.Native:
						items = append(items, nat.Value.(fyne.CanvasObject))
					}
				}
				win := container.NewVBox(items...)
				return *env.NewNative(ps.Idx, win, "fyne-container-vbox")
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.WordType}, "fyne-container-vbox")
			}
		},
	},

	"container-hbox": {
		Argsn: 1,
		Doc:   "Creates Fyne hbox container",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				items := []fyne.CanvasObject{}
				for _, it := range bloc.Series.S {
					switch nat := it.(type) {
					case env.Native:
						items = append(items, nat.Value.(fyne.CanvasObject))
					}
				}
				win := container.NewHBox(items...)
				return *env.NewNative(ps.Idx, win, "fyne-container-hbox")
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.WordType}, "fyne-container-hbox")
			}
		},
	},

	"container-grid-rows": {
		Argsn: 2,
		Doc:   "Creates Fyne grid with rows container",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch rows := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					items := []fyne.CanvasObject{}
					for _, it := range bloc.Series.S {
						switch nat := it.(type) {
						case env.Native:
							items = append(items, nat.Value.(fyne.CanvasObject))
						}
					}
					win := container.NewGridWithRows(int(rows.Value), items...)
					return *env.NewNative(ps.Idx, win, "fyne-container-grid-rows")
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.WordType}, "fyne-container-grid-rows")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.IntegerType}, "fyne-container-grid-rows")
			}
		},
	},

	"container-grid-cols": {
		Argsn: 2,
		Doc:   "Creates Fyne grid with cols container",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cols := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					items := []fyne.CanvasObject{}
					for _, it := range bloc.Series.S {
						switch nat := it.(type) {
						case env.Native:
							items = append(items, nat.Value.(fyne.CanvasObject))
						}
					}
					win := container.NewGridWithColumns(int(cols.Value), items...)
					return *env.NewNative(ps.Idx, win, "fyne-container-grid-cols")
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.WordType}, "fyne-container-grid-cols")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.IntegerType}, "fyne-container-grid-cols")
			}
		},
	},

	"container-center": {
		Argsn: 1,
		Doc:   "Creates Fyne center layout container",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				items := []fyne.CanvasObject{}
				for _, it := range bloc.Series.S {
					switch nat := it.(type) {
					case env.Native:
						items = append(items, nat.Value.(fyne.CanvasObject))
					}
				}
				win := container.NewCenter(items...)
				return *env.NewNative(ps.Idx, win, "fyne-container-center")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.WordType}, "fyne-container-center")
			}
		},
	},

	"button": {
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
						fmt.Println("Before click")
						evaldo.EvalBlockInj(ps, nil, false)
						fmt.Println("After click")
						fmt.Println(ps.Res)
						if ps.Res != nil && ps.Res.Type() == env.ErrorType {
							fmt.Println(ps.Res.(*env.Error).Message)
						}
						ps.Ser = ser
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
					return arg0
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.NativeType}, "fyne-window//set-content")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "fyne-window//set-content")
			}
		},
	},

	"fyne-window//resize": {
		Argsn: 3,
		Doc:   "Resize Fyne window",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				var width, height float32
				switch widthArg := arg1.(type) {
				case env.Integer:
					width = float32(widthArg.Value)
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "fyne-window//resize")
				}
				switch heightArg := arg2.(type) {
				case env.Integer:
					height = float32(heightArg.Value)
				default:
					return evaldo.MakeArgError(ps, 3, []env.Type{env.IntegerType}, "fyne-window//resize")
				}
				win.Value.(fyne.Window).Resize(fyne.NewSize(width, height))
				return win
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "fyne-window//resize")
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
