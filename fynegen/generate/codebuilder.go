package main

import (
	"fmt"
	"strings"
)

type CodeBuilder struct {
	b strings.Builder

	Indent int
}

func (w *CodeBuilder) Write(s string) {
	w.b.WriteString(s)
}

func (w *CodeBuilder) Linef(format string, args ...any) {
	for i := 0; i < w.Indent; i++ {
		w.b.WriteString("\t")
	}
	w.b.WriteString(fmt.Sprintf(format, args...))
	w.b.WriteString("\n")
}

func (w *CodeBuilder) String() string {
	return w.b.String()
}
