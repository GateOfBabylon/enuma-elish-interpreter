package main

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/engine"
	"github.com/GateOfBabylon/enuma-elish-interpreter/parser"
)

func main() {
	executor, err := parser.ParseExecutor("debug_dir/example-http.ea")
	if err != nil {
		panic(err)
	}
	fmt.Println("Executor: ", executor)
	err = engine.ExecuteExecutor(executor)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done!")
}
