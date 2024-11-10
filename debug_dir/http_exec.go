package debug_dir

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func HttpObjToEa() {
	executor := &types.Executor{
		Name: "fetch-and-process-items",
		Type: "http",
		Env: map[string]interface{}{
			"api_key": "qwerty",
		},
		Tasks: []types.Task{
			{
				Name:   "fetch-item-list",
				Export: "${{items}}",
				HttpTaskFields: &types.HttpTaskFields{
					Method: "GET",
					Url:    "http://example.com/api/item",
					Headers: map[string]string{
						"Authorization": "Bearer ${{api_key}}",
					},
				},
			},
			{
				Name: "fetch-additional-data-parallel",
				ConditionStatements: &types.ConditionStatements{
					Parallel: &[]types.Task{
						{
							Name: "fetch-item-details",
							ConditionStatements: &types.ConditionStatements{
								Iterate: "${{items}}",
							},
							Export: "${{items.details}}",
							HttpTaskFields: &types.HttpTaskFields{
								Method: "GET",
								Url:    "http://example.com/api/items/${{item.id}}/details",
								Headers: map[string]string{
									"Authorization": "Bearer ${{api_key}}",
								},
							},
						},
						{
							Name: "fetch-item-stats",
							ConditionStatements: &types.ConditionStatements{
								Iterate: "${{items}}",
							},
							Export: "${{items.stats}}",
							HttpTaskFields: &types.HttpTaskFields{
								Method: "GET",
								Url:    "http://example.com/api/items/${{item.id}}/stats",
								Headers: map[string]string{
									"Authorization": "Bearer ${{api_key}}",
								},
							},
						},
					},
				},
			},
			{
				Name: "process-each-item",
				ConditionStatements: &types.ConditionStatements{
					Iterate: "${{items}}",
				},
				HttpTaskFields: &types.HttpTaskFields{
					Method: "GET",
					Url:    "http://example.com/api/process/${{item.id}}",
					Headers: map[string]string{
						"Authorization": "Bearer ${{api_key}}",
						"Content-Type":  "application/json",
					},
					Body: "{\"details\":${{item.details}}, \"stats\":${{item.stats}}}",
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
	err = os.WriteFile("debug_dir/example-http.ea", data, 0644)
	if err != nil {
		log.Fatalf("Failed to write YAML to file: %v", err)
	}

	fmt.Println("Executor data has been saved to example-http.ea")
}

func HttpEaToObj() {
	data, err := os.ReadFile("debug_dir/example-http.ea")
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
