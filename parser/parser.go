package parser

import (
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func ParseExecutor(filePath string) (*types.Executor, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var run types.Run
	err = yaml.Unmarshal(data, &run)
	if err != nil {
		log.Fatalf("Failed to unmarshal data: %v", err)
	}
	return run.Executor, nil
}
