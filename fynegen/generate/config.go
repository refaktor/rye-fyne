package main

type Config struct {
	OutDir         string      `toml:"out-dir"`
	Package        string      `toml:"package"`
	Version        string      `toml:"version"`
	CutNew         bool        `toml:"cut-new"`
	BuildFlag      string      `toml:"build-flag"`
	NoPrefix       []string    `toml:"no-prefix,omitempty"`
	CustomPrefixes [][2]string `toml:"custom-prefixes,omitempty"` // {prefix, package}
}

const DefaultConfig = `# Output directory (relative).
out-dir = "../current"
# Go name of package.
package = "github.com/<user>/<repo>"
# Go semantic version of package.
version = "vX.Y.Z"
# Auto-remove "New" part of functions (e.g. widget.NewLabel => widget-label, app.New => app).
cut-new = true
# Build flag used to enable binding. * is replaced by package name.
build-flag = "b_*"

## Descending priority. Packages not listed will always be prefixed.
## In case of conflicting function names, only the function from the
## package with the highest priority is not prefixed.
#no-prefix = [
#  "github.com/<user>/<repo>",
#  "github.com/<user>/<repo>/important",
#]

## Set custom prefix for all symbols in the package (if applicable: see "no-prefix").
#custom-prefixes = [
#  ["my-fyne", "fyne.io/fyne/v2"],
#  ["my-widget", "fyne.io/fyne/v2/widget"],
#]`
