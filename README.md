# Rye-Fyne: GUI Programming with Rye Language

![Rye-Fyne](https://github.com/refaktor/rye-fyne/blob/main/image.png?raw=true)

Rye-Fyne brings the power of [Fyne](https://fyne.io) GUI toolkit to the [Rye programming language](https://ryelang.org/). Build cross-platform desktop applications with Rye's expressive syntax and Fyne's modern widgets.

### Download

Download pre-built binaries from [GitHub Releases](https://github.com/refaktor/rye-fyne/releases/latest):
* **Linux**: `rye-fyne-linux-amd64.tar.gz`
* **macOS**: `rye-fyne-macos-amd64.tar.gz`
* **Windows**: `rye-fyne.exe`

### Building from Source

You need [Go](https://go.dev/) installed on your system.

```bash
# Clone the repository
git clone https://github.com/refaktor/rye-fyne.git
cd rye-fyne

# for mac and windows
go tool ryegen

# Build the project
go build

# Run the example
./rye-fyne examples/01-hello-world.rye
```

## Example Application

The `example.rye` file demonstrates a comprehensive GUI application with various widgets:

```rye
fyne: import\go "fyne"
app: import\go "fyne/app"
widget: import\go "fyne/widget"
dialog: import\go "fyne/dialog"
container: import\go "fyne/container"
theme: import\go "fyne/theme"

a: app/new
w: a .window "Hello, world!"

; Menu system with file operations and settings
w .set-main-menu fyne/main-menu [
    fyne/menu "File" [
        fyne/menu-item "Open" does {
            dialog/show-file-open fn { r err } {
                either r .is-nil {
                    if not err .is-nil { print err }
                } {
                    print "Opened file: " ++ r .uri .string
                    r .close
                }
            } w
        } |icon! theme/file-icon
    ]
    fyne/menu "Settings" [
        fyne/menu-item "Preferences" does { }
            |icon! theme/settings-icon
    ]
]

; Live clock that updates every second
clock: widget/label ""
go does {
    forever {
        fyne/do does { clock .set-text now .to-string }
        sleep 1 .seconds
    }
}

; Main application layout with various widgets
w .set-content container/border
    container/vbox [
        widget/label "Hello!"
        do {
            se: widget/select-entry [ "A" "B" ]
            se .set-text "A"
            se
        }
        do {
            p: widget/progress-bar-infinite
            p .start
            p
        }
        widget/button "Say hi" does {
            dialog/show-information "Hello!" "You clicked the button" w
        }
        clock
    ]
    nil nil nil
    [ widget/list
        does { 1000 }
        does { widget/label "Hi" }
        fn { id lab } { lab .set-text "Entry no. " ++ id .to-string }
    ]

w .show-and-run
```

This example showcases:
- **Menu System**: File and Settings menus with icons
- **Dialogs**: File open dialog and information dialogs
- **Widgets**: Labels, buttons, select entries, progress bars, lists
- **Layouts**: Border and VBox container layouts
- **Live Updates**: A clock that updates in real-time
- **Event Handling**: Button clicks and menu interactions

## Features

### Available Widgets
- Labels, buttons, entries, and text areas
- Progress bars (determinate and indeterminate)
- Lists, tables, and trees
- Select entries and combo boxes
- Sliders, check boxes, and radio buttons
- And many more Fyne widgets

### Layout Containers
- Border, VBox, HBox layouts
- Grid and form layouts
- Split containers and accordions
- Tabs and scroll containers

### Dialogs and Menus
- File dialogs (open, save)
- Information, confirmation, and error dialogs
- Menu bars with icons and shortcuts
- Context menus

## Examples


## Interactive Development

Start the Rye console for interactive GUI development:

```bash
./rye-fyne
```

```rye
; Quick GUI creation in the console
rye> fyne: import\go "fyne"
rye> app: import\go "fyne/app" 
rye> widget: import\go "fyne/widget"
rye> container: import\go "fyne/container"

rye> a: app/new
rye> w: a .window "Quick Demo"
rye> w .set-content widget/button "Click me!" does { print "Hello from Rye!" }
rye> w .show-and-run
```

## Resources

- **[Rye Language](https://github.com/refaktor/rye)** - The core Rye language
- **[Rye Website](https://ryelang.org/)** - Documentation and tutorials  
- **[Rye Cookbook](https://ryelang.org/cookbook/rye-fyne/examples/)** - GUI examples and recipes
- **[Fyne Documentation](https://fyne.io/)** - Fyne GUI toolkit documentation
- **[Reddit Community](https://reddit.com/r/ryelang/)** - Join the discussion

## What is Rye?

Rye is a high-level, dynamic programming language inspired by Rebol, Factor, and Go. It emphasizes:
- **Expressive Syntax**: Clean, readable code that's easy to write and understand
- **Interactive Development**: REPL-driven programming with live feedback
- **Go Integration**: Easy access to Go libraries and embedding in Go applications
- **Cross-platform**: Build applications that run on Windows, macOS, and Linux

## Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests on GitHub.

## License

This project is open source. See the [LICENSE](LICENSE) file for details.
