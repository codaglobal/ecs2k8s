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
	"io/ioutil"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the YAML or Helm charts for the tasks",
	Long: `Generate the YAML or Helm charts for the tasks. For example:

	ecs2k8s generate --task <task name>

	Flags:
		--options [ YAML | HELM ]
		--task < Name of the task definition >
		--container-name < Name of the container inside the task, if more than one container is specified in that task > (optional field)
		--install < Generate the local copy and also install the same into the k8s cluster > (optional field)  
`,
	Run: func(cmd *cobra.Command, args []string) {
		taskDefintion, _ := cmd.Flags().GetString("task")
		fileName, _ := cmd.Flags().GetString("fname")

		if fileName == "" {
			fileName = getDefaultFileName()
		}

		if taskDefintion != "" {
			td := getTaskDefiniton(taskDefintion, fileName)
			generateDeploymentYAML(td, fileName)
		} else {
			fmt.Println("Task definition required")
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.PersistentFlags().String("task", "td", "A valid task definition in ECS")
}

func getTaskDefiniton(taskDefinition string, fileName string) ecs.DescribeTaskDefinitionOutput {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	client := ecs.NewFromConfig(cfg)

	output, err := client.DescribeTaskDefinition(context.TODO(), &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: &taskDefinition,
	})

	if err != nil {
		log.Fatal(err)
	}

	return *output
}

func generateDeploymentYAML(output ecs.DescribeTaskDefinitionOutput, fileName string) {

	var kubeContainers []kubeContainer

	for _, object := range output.TaskDefinition.ContainerDefinitions {
		c := kubeContainer{
			Name:  *object.Name,
			Image: *object.Image,
		}
		kubeContainers = append(kubeContainers, c)
	}

	data := kubeObject{
		ApiVersion: "apps/v1",
		Kind:       "Deployment",
		MetaData: kubeMetadata{
			Name: *output.TaskDefinition.Family,
		},
	}

	data.Spec.Template.Spec.Containers = kubeContainers

	file, _ := yaml.Marshal(data)

	fileName = fileName + ".yaml"

	fmt.Println("Writing K8s Deployment file to : ", fileName)
	_ = ioutil.WriteFile(fileName, file, 0644)
}

func getDefaultFileName() string {
	const layout = "2006-01-02"
	t := time.Now()
	return "k8s-deployment-" + t.Format(layout)
}
