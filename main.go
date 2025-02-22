package main

import (
	"github.com/GateOfBabylon/enuma-elish-interpreter/engine"
	"github.com/GateOfBabylon/enuma-elish-interpreter/parser"
)

func main() {
	executor, err := parser.ParseExecutor("debug_dir/example-py.ea")
	if err != nil {
		panic(err)
	}
	err = engine.ExecuteExecutor(executor)
	if err != nil {
		panic(err)
	}
}
