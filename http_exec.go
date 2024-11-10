package main

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func MainHttpExecutor() {

	executor := &types.Executor{
		Name: "fetch-and-process-items",
		Type: "http",
		Env: map[string]interface{}{
			"api_key": "qwerty",
		},
		Tasks: []types.Task{
			{
				Name:   "fetch-item-list",
				Method: "GET",
				Url:    "http://example.com/api/item",
				Headers: map[string]string{
					"Authorization": "Bearer ${{api_key}}",
				},
				Export: "${{items}}",
			},
			{
				Name: "fetch-additional-data-parallel",
				Parallel: []types.Task{
					{
						Name:    "fetch-item-details",
						Iterate: "${{items}}",
						Method:  "GET",
						Url:     "http://example.com/api/items/${{item.id}}/details",
						Headers: map[string]string{
							"Authorization": "Bearer ${{api_key}}",
						},
						Export: "${{items.details}}",
					},
					{
						Name:    "fetch-item-stats",
						Iterate: "${{items}}",
						Method:  "GET",
						Url:     "http://example.com/api/items/${{item.id}}/stats",
						Headers: map[string]string{
							"Authorization": "Bearer ${{api_key}}",
						},
						Export: "${{items.stats}}",
					},
				},
			},
			{
				Name:    "process-each-item",
				Iterate: "${{items}}",
				Method:  "GET",
				Url:     "http://example.com/api/process/${{item.id}}",
				Headers: map[string]string{
					"Authorization": "Bearer ${{api_key}}",
					"Content-Type":  "application/json",
				},
				Body: "{\"details\":${{item.details}}, \"stats\":${{item.stats}}}",
			},
		},
	}

	// Marshal the Executor struct to YAML
	data, err := yaml.Marshal(&types.Run{Executor: executor})
	if err != nil {
		log.Fatalf("Failed to marshal struct to YAML: %v", err)
	}

	// Save the YAML data to a file named example.ea
	err = os.WriteFile("example-http.ea", data, 0644)
	if err != nil {
		log.Fatalf("Failed to write YAML to file: %v", err)
	}

	fmt.Println("Executor data has been saved to example-http.ea")
}
