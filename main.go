package main

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/engine"
	. "github.com/GateOfBabylon/enuma-elish-interpreter/logger"
	"github.com/GateOfBabylon/enuma-elish-interpreter/parser"
	"os"
)

func main() {
	logger := GetLogger()

	if len(os.Args) < 2 {
		logger.Log("path to ea file is not specified")
		printLogsAndExit(logger, 1)
	}

	eaFile := os.Args[1]

	executor, err := parser.ParseExecutor(eaFile)
	if err != nil {
		logger.Log("Failed to parse EA file: %v", err)
		printLogsAndExit(logger, 1)
	}

	err = engine.ExecuteExecutor(executor)
	if err != nil {
		logger.Log("Executor execution failed: %v", err)
		printLogsAndExit(logger, 1)
	}

	logger.Log("Execution completed successfully.")
	printLogsAndExit(logger, 0)
}

func printLogsAndExit(logger *Logger, exitCode int) {
	for _, log := range logger.GetLogs() {
		fmt.Println(log)
	}
	os.Exit(exitCode)
}
