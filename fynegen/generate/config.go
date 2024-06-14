package main

type Config struct {
	Package          string      `toml:"package"`
	Version          string      `toml:"version"`
	CutNew           bool        `toml:"cut-new"`
	AlwaysPrefix     bool        `toml:"always-prefix"`
	PriorityPackages []string    `toml:"priority-packages,omitempty"`
	Prefixes         [][2]string `toml:"prefixes,omitempty"`   // {prefix, package}
	Substitute       [][2]string `toml:"substitute,omitempty"` // {old, new}
}

const DefaultConfig = `# Go name of package.
package = "github.com/<user>/<repo>"
# Go semantic version of package.
version = "vX.Y.Z"
# Auto-remove "New" part of functions (e.g. widget.NewLabel => widget-label, app.New => app).
cut-new = true
# If true, always prefix function with package name.
# If false, only conflicting functions are prefixed (see "priority-packages").
# See "prefixes" for custom prefixes.
always-prefix = false

## Descending priority. Packages not listed are always lower priority.
## In case of conflicting function names, only the function from the
## package with the highest priority is not prefixed. All other functions
## are prefixed.
## See "always-prefix".
#priority-packages = [
#  "github.com/<user>/<repo>",
#  "github.com/<user>/<repo>/important_math",
#  "github.com/<user>/<repo>/not_as_important_math",
#]

## Set prefix for all symbols in the package (if applicable: see "always-prefix").
#prefixes = [
#  ["my-fyne", "fyne.io/fyne/v2"],
#  ["my-widget", "fyne.io/fyne/v2/widget"],
#]

## Any time the first identifier is encountered, replace with the second identifier.
## Useful if argument type is unexported (first identifier), but has an
## exported typedef (second identifier).
#substitute = [
#  [
#    "fyne.io/fyne/v2/internal/widget.ScrollDirection",
#    "fyne.io/fyne/v2/container.ScrollDirection",
#  ],
#]`
