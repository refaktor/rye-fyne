package main

import (
	"fmt"
	"go/format"
	"os"
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

func (w *CodeBuilder) FmtString() (string, error) {
	code := []byte(w.String())
	code, err := format.Source(code)
	if err != nil {
		return "", err
	}
	return string(code), nil
}

func (w *CodeBuilder) Reset() {
	w.b.Reset()
}

func (w *CodeBuilder) SaveToFile(outFile string, goFmt bool) error {
	var code string
	if goFmt {
		var err error
		code, err = w.FmtString()
		if err != nil {
			return err
		}
	} else {
		code = w.String()
	}
	if err := os.WriteFile(outFile, []byte(code), 0666); err != nil {
		return err
	}
	return nil
}
