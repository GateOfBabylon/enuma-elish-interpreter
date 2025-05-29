package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/GateOfBabylon/enuma-elish-interpreter/engine"
	"github.com/GateOfBabylon/enuma-elish-interpreter/logger"
	"github.com/GateOfBabylon/enuma-elish-interpreter/parser"
)

func main() {
	log := logger.GetLogger()

	if len(os.Args) < 2 || os.Args[1] == "--help" {
		printHelp()
		os.Exit(0)
	}

	eaFile := os.Args[1]

	if !strings.HasSuffix(eaFile, ".ea") {
		log.Log("Error: Provided file does not have a .ea extension: %s", eaFile)
		os.Exit(1)
	}

	executor, err := parser.ParseExecutor(eaFile)
	if err != nil {
		log.Log("Failed to parse EA file: %v", err)
		os.Exit(1)
	}

	err = engine.ExecuteExecutor(executor)
	if err != nil {
		log.Log("Executor execution failed: %v", err)
		os.Exit(1)
	}

	log.Log("Execution completed successfully.")
	os.Exit(0)
}

// printHelp displays usage instructions for the interpreter.
func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  enumago <path_to_ea_file>")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --help   Show this help message")
	fmt.Println("")
	fmt.Println("Description:")
	fmt.Println("  This program is an interpreter for EA (Enuma Elish) files. Provide a path to an EA file with a .ea extension to execute it.")
}
