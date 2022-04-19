/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/spf13/cobra"
)

// listTasksCmd represents the listTasks command
var listTasksCmd = &cobra.Command{
	Use:   "list-tasks",
	Short: "Lists all the ECS tasks that are in active state",
	Long:  `Lists all the ECS tasks that are in active state`,
	Run: func(cmd *cobra.Command, args []string) {
		outFamilies := getTaskDefinitonFamilies()
		printList(outFamilies)
	},
}

func init() {
	ecsCmd.AddCommand(listTasksCmd)

}

// Prints list of task definitons
func printList(output *ecs.ListTaskDefinitionFamiliesOutput) {
	for _, object := range output.Families {
		fmt.Println(object)
	}
}

// Gets active task definiton families from ECS
func getTaskDefinitonFamilies() (resp *ecs.ListTaskDefinitionFamiliesOutput) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	client := ecs.NewFromConfig(cfg)

	// Get active Taskdefinition families
	output, err := client.ListTaskDefinitionFamilies(context.TODO(), &ecs.ListTaskDefinitionFamiliesInput{
		Status: "ACTIVE",
	})

	if err != nil {
		log.Fatal(err)
	}

	return output
}
