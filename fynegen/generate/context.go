package main

type Context struct {
	Config      *Config
	Data        *Data
	ModuleNames map[string]string
	UsedImports map[string]struct{}
	UsedTyps    map[string]Ident
}

func (c *Context) MarkUsed(id Ident) {
	if id.File == nil {
		return
	}
	c.UsedTyps[id.GoName] = id
	for _, imp := range id.UsedImports {
		/*if ModulePathIsInternal(c, imp.ModulePath) {
			panic("cannot use " + id.GoName + " from internal module " + imp.ModulePath)
		}*/
		c.UsedImports[imp.ModulePath] = struct{}{}
	}
}
