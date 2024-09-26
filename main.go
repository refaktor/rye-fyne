//go:build linux || darwin || windows
// +build linux darwin windows

package main

import (
	"github.com/refaktor/rye-front/current"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/runner"
	"github.com/jwalton/go-supportscolor"
)

/* type TagType int
type RjType int
type Series []any

type anyword struct {
	kind RjType
	idx  int
}

type node struct {
	kind  RjType
	value any
}

   var CODE []any */

//
// main function. Dispatches to appropriate mode function
//

func main() {

	supportscolor.Stdout()
	runner.DoMain(func(ps *env.ProgramState) {
		current.RegisterBuiltins(ps)
	})

}

/*
func main_() {
	evaldo.ShowResults = true

	code := " "

	if len(os.Args) == 1 {
		main_rye_string(code, false, false)
	} else if len(os.Args) == 3 {
		main_rye_repl(os.Stdin, os.Stdout, false, false)
	} else if len(os.Args) == 2 {
		main_rye_file(os.Args[1], false, false)
	}
}

func main_rye_string(content string, sig bool, subc bool) {
	// info := true
	//util.PrintHeader()
	//defer profile.Start(profile.CPUProfile).Stop()

	block, genv := loader.LoadString(content, sig)
	switch val := block.(type) {
	case env.Block:
		es := env.NewProgramState(block.(env.Block).Series, genv)
		evaldo.RegisterBuiltins(es)
		contrib.RegisterBuiltins(es, &evaldo.BuiltinNames) // TODO -- remove this in next Rye release
		current.RegisterBuiltins(es)

		if subc {
			ctx := es.Ctx
			es.Ctx = env.NewEnv(ctx)
		}

		evaldo.EvalBlock(es)
		evaldo.MaybeDisplayFailureOrError(es, genv)
	case env.Error:
		fmt.Println(val.Message)
	}
}

func main_rye_file(file string, sig bool, subc bool) {
	info := true
	//util.PrintHeader()
	//defer profile.Start(profile.CPUProfile).Stop()

	bcontent, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	content := string(bcontent)

	if info {
		pattern := regexp.MustCompile(`^; (#[^\n]*)`)

		lines := pattern.FindAllStringSubmatch(content, -1)

		for _, line := range lines {
			if line[1] != "" {
				fmt.Println(line[1])
			}
		}
	}

	block, genv := loader.LoadString(content, sig)
	switch val := block.(type) {
	case env.Block:
		es := env.NewProgramState(block.(env.Block).Series, genv)
		evaldo.RegisterBuiltins(es)
		contrib.RegisterBuiltins(es, &evaldo.BuiltinNames) // TODO -- remove this in next Rye release
		current.RegisterBuiltins(es)

		if subc {
			ctx := es.Ctx
			es.Ctx = env.NewEnv(ctx)
		}

		evaldo.EvalBlock(es)
		evaldo.MaybeDisplayFailureOrError(es, genv)
	case env.Error:
		fmt.Println(val.Message)
	}
}

func main_rye_repl(_ io.Reader, _ io.Writer, subc bool, here bool) {
	input := " 123 " // "name: \"Rye\" version: \"0.011 alpha\""
	// userHomeDir, _ := os.UserHomeDir()
	// profile_path := filepath.Join(userHomeDir, ".rye-profile")

	fmt.Println("Welcome to Rye shell. Use ls and ls\\ \"pr\" to list the current context.")

	//if _, err := os.Stat(profile_path); err == nil {
	//content, err := os.ReadFile(profile_path)
	//if err != nil {
	//	log.Fatal(err)
	//}
	// input = string(content)
	//} else {
	//		fmt.Println("There was no profile.")
	//}

	block, genv := loader.LoadString(input, false)
	es := env.NewProgramState(block.(env.Block).Series, genv)
	evaldo.RegisterBuiltins(es)
	contrib.RegisterBuiltins(es, &evaldo.BuiltinNames) // TODO -- remove this in next Rye release
	current.RegisterBuiltins(es)

	evaldo.EvalBlock(es)

	if subc {
		ctx := es.Ctx
		es.Ctx = env.NewEnv(ctx) // make new context with no parent
	}

	if here {
		if _, err := os.Stat(".rye-here"); err == nil {
			content, err := os.ReadFile(".rye-here")
			if err != nil {
				log.Fatal(err)
			}
			inputH := string(content)
			block, genv := loader.LoadString(inputH, false)
			block1 := block.(env.Block)
			es = env.AddToProgramState(es, block1.Series, genv)
			evaldo.EvalBlock(es)
		} else {
			fmt.Println("There was no `here` file.")
		}
	}

	evaldo.DoRyeRepl(es, "do", evaldo.ShowResults)
}
*/
