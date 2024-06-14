package main

import (
	"strings"
)

type Context struct {
	Config      *Config
	Data        *Data
	ModuleNames map[string]string
	UsedImports map[string]struct{}
}

func (c Context) MarkUsed(id Ident) {
	if id.File == nil {
		return
	}
	for _, imp := range id.UsedImports {
		if _, isExternal := c.ModuleNames[imp.ModulePath]; isExternal {
			sp := strings.Split(imp.ModulePath, "/")
			if len(sp) < 3 {
				panic("malformed module path " + imp.ModulePath)
			}
			for _, elem := range sp[3:] {
				if elem == "internal" {
					panic("cannot use " + id.GoName + " from internal module " + imp.ModulePath)
				}
			}
		}
		c.UsedImports[imp.ModulePath] = struct{}{}
	}
}
