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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/spf13/cobra"

	// "gopkg.in/yaml.v2"
	gyaml "github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the YAML or Helm charts for the tasks",
	Long:  `Generate the YAML or Helm charts for the tasks. For example:`,
	Run: func(cmd *cobra.Command, args []string) {
		taskDefintion, _ := cmd.Flags().GetString("task")
		fileName, _ := cmd.Flags().GetString("file-name")
		rCount, _ := cmd.Flags().GetInt32("replicas")
		yaml, _ := cmd.Flags().GetBool("yaml")

		if fileName == "" {
			fileName = getDefaultFileName()
		}

		if taskDefintion == "" {
			fmt.Println("Task definition required")
			return
		}

		td := getTaskDefiniton(taskDefintion)
		d := generateDeploymentObject(td, rCount)
		generateDeploymentFile(d, fileName, yaml)

	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.PersistentFlags().String("task", "", "A valid task definition in ECS")
	generateCmd.PersistentFlags().StringP("file-name", "f", "", "File to write the deployment file into")
	generateCmd.PersistentFlags().Int32("replicas", 1, "Number of replicas")
	generateCmd.Flags().Bool("yaml", false, "Set this flag if YAML needs to be generated")
}

// Fetch Task definition from ECS
func getTaskDefiniton(taskDefinition string) ecs.DescribeTaskDefinitionOutput {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	fmt.Println("Fetching", taskDefinition, "from ECS...")
	if err != nil {
		log.Fatal(err)
	}

	client := ecs.NewFromConfig(cfg)

	output, err := client.DescribeTaskDefinition(context.TODO(), &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: &taskDefinition,
		Include:        []types.TaskDefinitionField{"TAGS"},
	})

	if err != nil {
		log.Fatal(err)
	}

	return *output
}

// Generate K8s deployment object
func generateDeploymentObject(output ecs.DescribeTaskDefinitionOutput, rCount int32) *appsv1.Deployment {
	var kubeContainers []apiv1.Container
	var kubeLabels map[string]string = make(map[string]string)

	// Imports tags to labels
	for _, object := range output.Tags {
		key := sanitizeValue(*object.Key)
		value := sanitizeValue(*object.Value)
		kubeLabels[key] = value
	}

	// Imports container definition – Name, Image, Port mapping
	for _, object := range output.TaskDefinition.ContainerDefinitions {
		var containerPorts []apiv1.ContainerPort

		PortMappings := object.PortMappings
		for _, object := range PortMappings {
			cp := apiv1.ContainerPort{
				HostPort:      *object.HostPort,
				ContainerPort: *object.ContainerPort,
				Protocol:      apiv1.ProtocolTCP,
			}
			containerPorts = append(containerPorts, cp)
		}

		c := apiv1.Container{
			Name:    *object.Name,
			Image:   *object.Image,
			Ports:   containerPorts,
			Command: object.Command,
		}

		c.Resources = apiv1.ResourceRequirements{
			Requests: apiv1.ResourceList{
				"cpu":    resource.MustParse(fmt.Sprintf("%d%s", object.Cpu, "Mi")),
				"memory": resource.MustParse(fmt.Sprintf("%d%s", *object.Memory, "m")),
			},
		}
		kubeContainers = append(kubeContainers, c)
	}

	//Create deployment object
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name: *output.TaskDefinition.Family,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &rCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: kubeLabels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: kubeLabels,
				},
				Spec: apiv1.PodSpec{
					Containers: kubeContainers,
				},
			},
		},
	}
	return deployment
}

func generateDeploymentFile(d *appsv1.Deployment, fileName string, yaml bool) {
	bytes, _ := json.MarshalIndent(d, "", "  ")
	if yaml {
		y, _ := gyaml.JSONToYAML(bytes)
		fileName = fileName + ".yaml"
		fmt.Println("Writing K8s Deployment YAML file to : ", fileName)
		_ = ioutil.WriteFile(fileName, y, 0644)
	} else {
		fileName = fileName + ".json"
		fmt.Println("Writing K8s Deployment JSON file to : ", fileName)
		_ = ioutil.WriteFile(fileName, bytes, 0644)
	}

}
