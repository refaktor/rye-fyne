package main

import (
	"bufio"
	"bytes"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"unicode"
)

type BindingList struct {
	Enabled map[string]bool
}

func NewBindingList() *BindingList {
	return &BindingList{
		Enabled: make(map[string]bool),
	}
}

func LoadBindingListFromFile(filename string) (*BindingList, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	res := NewBindingList()

	var inSection, inEnabledSection bool
	sc := bufio.NewScanner(f)
	for lineNum := 1; sc.Scan(); lineNum++ {
		makeErr := func(format string, a ...any) error {
			return fmt.Errorf("%v: line %v: %v", filename, lineNum, fmt.Errorf(format, a...))
		}

		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inSection = true
			if line == "[enabled]" {
				inEnabledSection = true
			} else if line == "[disabled]" {
				inEnabledSection = false
			} else {
				return nil, makeErr("invalid section name %v", line)
			}
		}
		if !inSection {
			return nil, makeErr("expected binding name \"%v\" to be under a section ([enabled] or [disabled])", line)
		}
		var name string
		{
			idx := strings.IndexFunc(line, func(r rune) bool {
				return unicode.IsSpace(r)
			})
			if idx == -1 {
				idx = len(line)
			}
			if idx == 0 {
				panic("expected line to be nonempty")
			}
			name = line[:idx]
		}
		res.Enabled[name] = inEnabledSection
	}
	return res, nil
}

func (bl *BindingList) SaveToFile(filename string, bindingFuncsToDocstrs map[string]string) error {
	isEnabled := maps.Clone(bl.Enabled)
	for name := range bindingFuncsToDocstrs {
		if _, ok := isEnabled[name]; !ok {
			isEnabled[name] = true
		}
	}

	var enabledBindings []string
	var disabledBindings []string
	for name, enabled := range isEnabled {
		if enabled {
			enabledBindings = append(enabledBindings, name)
		} else {
			disabledBindings = append(disabledBindings, name)
		}
	}
	slices.Sort(enabledBindings)
	slices.Sort(disabledBindings)

	var res bytes.Buffer
	fmt.Fprintln(&res, "# This file contains a list of bindings, which can be enabled/disabled by placing them under the according section.")
	fmt.Fprintln(&res, "# Re-run `go generate ./...` to update and sort the list.")
	fmt.Fprintln(&res)
	writeBindings := func(bs []string) {
		maxLen := 0
		for _, name := range bs {
			if len(name) > maxLen {
				maxLen = len(name)
			}
		}
		for _, name := range bs {
			if docstr, ok := bindingFuncsToDocstrs[name]; ok {
				fmt.Fprintf(
					&res,
					"%v %v\"%v\"\n",
					name,
					strings.Repeat(" ", maxLen-len(name)),
					docstr,
				)
			}
		}
	}
	fmt.Fprintln(&res, "[enabled]")
	writeBindings(enabledBindings)
	fmt.Fprintln(&res)
	fmt.Fprintln(&res, "[disabled]")
	writeBindings(disabledBindings)

	if err := os.WriteFile(filename, res.Bytes(), 0666); err != nil {
		return err
	}
	return nil
}
