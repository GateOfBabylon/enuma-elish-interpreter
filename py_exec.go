package main

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func MainPyExecutor() {
	// Create an Executor struct with values
	executor := &types.Executor{
		Name: "confirm-or-reject-status",
		Type: "py",
		Universe: types.Universe{
			World:  "docker.io/my-docker-image:latest",
			Secret: "my_docker_registry_credentials",
		},
		Env: map[string]interface{}{
			"status": "pending",
		},
		Tasks: []types.Task{
			{
				Name:       "check-status",
				Export:     "${{status}}",
				ScriptPath: "./scripts/check_status.py",
			},
			{
				Name:       "wait-for-ready",
				ScriptPath: "./scripts/wait_for_ready.py ${{status}}",
			},
			{
				Name: "confirm-or-reject",
				Tries: []types.Try{
					{
						If: "${{status == 'ready'}}",
						Task: &types.Task{
							ScriptPath: "./scripts/confirm_resource.py",
						},
					},
					{
						If: "${{status == 'pending'}}",
						Task: &types.Task{
							ScriptPath: "./scripts/queue_resource_for_review.py",
						},
					},
					{
						If: "${{status == 'error'}}",
						Task: &types.Task{
							ScriptPath: "./scripts/report_resource_error.py",
						},
					},
				},
				Else: &types.Task{
					ScriptPath: "./scripts/reject_resource.py",
				},
			},
		},
	}

	// Marshal the Executor struct to YAML
	data, err := yaml.Marshal(&types.Run{Executor: executor})
	if err != nil {
		log.Fatalf("Failed to marshal struct to YAML: %v", err)
	}

	// Save the YAML data to a file named example.ea
	err = os.WriteFile("example-py.ea", data, 0644)
	if err != nil {
		log.Fatalf("Failed to write YAML to file: %v", err)
	}

	fmt.Println("Executor data has been saved to example-py.ea")
}
