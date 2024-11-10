package debug_dir

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func PyObjToEa() {
	// Create an Executor struct with values
	executor := &types.Executor{
		Name: "confirm-or-reject-status",
		Type: "py",
		Universe: &types.Universe{
			World:  "docker.io/my-docker-image:latest",
			Secret: "my_docker_registry_credentials",
		},
		Env: map[string]interface{}{
			"status": "pending",
		},
		Tasks: []types.Task{
			{
				Name:   "check-status",
				Export: "${{status}}",
				PyTaskFields: &types.PyTaskFields{
					ScriptPath: "./scripts/check_status.py",
				},
			},
			{
				Name: "wait-for-ready",
				PyTaskFields: &types.PyTaskFields{
					ScriptPath: "./scripts/wait_for_ready.py ${{status}}",
				},
			},
			{
				Name: "confirm-or-reject",
				ConditionStatements: &types.ConditionStatements{
					Picks: &types.PickStatement{
						IfStatement: []types.IfStatement{
							{
								Try: "${{status == 'ready'}}",
								Task: &types.Task{
									PyTaskFields: &types.PyTaskFields{
										ScriptPath: "./scripts/confirm_resource.py",
									},
								},
							},
							{
								Try: "${{status == 'pending'}}",
								Task: &types.Task{
									PyTaskFields: &types.PyTaskFields{
										ScriptPath: "./scripts/queue_resource_for_review.py",
									},
								},
							},
							{
								Try: "${{status == 'error'}}",
								Task: &types.Task{
									PyTaskFields: &types.PyTaskFields{
										ScriptPath: "./scripts/report_resource_error.py",
									},
								},
							},
						},
						Else: &types.Task{
							PyTaskFields: &types.PyTaskFields{
								ScriptPath: "./scripts/reject_resource.py",
							},
						},
					},
				},
			},
		},
	}

	// Marshal the Executor struct to YAML
	data, err := yaml.Marshal(&types.Run{Executor: executor})
	if err != nil {
		log.Fatalf("Failed to marshal struct to YAML: %v", err)
	}

	// Save the YAML data to a file named example-py.ea
	err = os.WriteFile("debug_dir/example-py.ea", data, 0644)
	if err != nil {
		log.Fatalf("Failed to write YAML to file: %v", err)
	}

	fmt.Println("Executor data has been saved to example-py.ea")
}

func PyEaToObj() {
	data, err := os.ReadFile("debug_dir/example-py.ea")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	run := types.Run{}
	err = yaml.Unmarshal(data, &run)
	if err != nil {
		log.Fatalf("Failed to unmarshal data: %v", err)
	}
	log.Println(run.Executor)
}
