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
package ecsCmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/spf13/cobra"

	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
	// "gopkg.in/yaml.v2"
	gyaml "github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate-k8s-spec",
	Short: "Generate the YAML or Helm charts for the tasks",
	Long:  `Generate the YAML or Helm charts for the tasks. For example:`,
	Run: func(cmd *cobra.Command, args []string) {
		taskDefintion, _ := cmd.Flags().GetString("task-definition")
		fileName, _ := cmd.Flags().GetString("file-name")
		rCount, _ := cmd.Flags().GetInt32("replicas")
		yaml, _ := cmd.Flags().GetBool("yaml")
		namespace, _ := cmd.Flags().GetString("namespace")

		if fileName == "" {
			fileName = getDefaultFileName()
		}

		if taskDefintion == "" {
			fmt.Println("Task definition required")
			os.Exit(1)
		}

		if namespace == "" {
			fmt.Println("Namespace required")
			os.Exit(1)
		}

		td := getTaskDefiniton(taskDefintion)
		d := generateDeploymentObject(td, rCount, namespace)

		kubeObjects := []interface{}{
			&d.kubernetesDeployment,
			&d.kubernetesSecrets[0],
		}
		generateK8sSpecFile(kubeObjects, fileName, yaml)
	},
}

func init() {
	ecsCmd.AddCommand(generateCmd)
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

type DeploymentObject struct {
	kubernetesDeployment appsv1.Deployment
	kubernetesSecrets    []apiv1.Secret
}

// Generate K8s deployment object
func generateDeploymentObject(output ecs.DescribeTaskDefinitionOutput, rCount int32, namespace string) DeploymentObject {
	var secrets []apiv1.Secret
	var kubeContainers []apiv1.Container
	var kubeLabels map[string]string = make(map[string]string)

	// Imports tags to labels
	for _, object := range output.Tags {
		key := sanitizeValue(*object.Key, labelSpecialChars, "-")
		value := sanitizeValue(*object.Value, labelSpecialChars, "-")
		kubeLabels[key] = value
	}

	// Imports container definition – Name, Image, Port mapping
	for _, object := range output.TaskDefinition.ContainerDefinitions {
		// K8s object declarations
		var containerPorts []apiv1.ContainerPort
		var envVars []apiv1.EnvVar
		// ECS object
		PortMappings := object.PortMappings
		EnvironmentVars := object.Environment
		Secrets := object.Secrets

		// Port mapping
		for _, object := range PortMappings {
			cp := apiv1.ContainerPort{
				HostPort:      *object.HostPort,
				ContainerPort: *object.ContainerPort,
				Protocol:      apiv1.ProtocolTCP,
			}
			containerPorts = append(containerPorts, cp)
		}

		// Environment variable mapping
		for _, env := range EnvironmentVars {
			ev := apiv1.EnvVar{
				Name:  *env.Name,
				Value: *env.Value,
			}
			envVars = append(envVars, ev)
		}

		// ECS Secrets (Secrets Manager) mounted as Environment variables from Kubernetes Secrets

		// var kubeSecrets []string
		for _, ecsSecret := range Secrets {
			secretData := make(map[string][]byte)
			envVarName := sanitizeValue(*ecsSecret.Name, envSpecialChars, "")

			secretName, secretValue, secretKey := parseSecret(*ecsSecret.ValueFrom)

			// If key is not provided in valueFrom of ECS Secret, we are using environment variable name as key and entire secret is available in the env var
			if secretKey == "" {
				secretKey = envVarName
			}

			secretData[secretKey] = []byte(secretValue)

			sev := apiv1.EnvVar{
				Name: envVarName,
				ValueFrom: &apiv1.EnvVarSource{
					SecretKeyRef: &apiv1.SecretKeySelector{
						LocalObjectReference: apiv1.LocalObjectReference{
							Name: secretName,
						},
						Key: secretKey,
					},
				},
			}

			s := createK8sSecret(secretName, secretData, namespace)

			secrets = append(secrets, s)
			envVars = append(envVars, sev)
		}

		c := apiv1.Container{
			Name:    *object.Name,
			Image:   *object.Image,
			Ports:   containerPorts,
			Command: object.Command,
			Env:     envVars,
		}

		c.Resources = apiv1.ResourceRequirements{
			Limits: apiv1.ResourceList{
				"cpu":    resource.MustParse(fmt.Sprintf("%d%s", object.Cpu, "m")),
				"memory": resource.MustParse(fmt.Sprintf("%d%s", *object.Memory, "Mi")),
			},
		}
		kubeContainers = append(kubeContainers, c)
	}

	//Create deployment object
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      *output.TaskDefinition.Family,
			Namespace: namespace,
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
	return DeploymentObject{
		kubernetesSecrets:    secrets,
		kubernetesDeployment: *deployment,
	}
}

func mergeKubeObjects(kubeObjects []interface{}, yaml bool) string {
	var mergedObject string

	var buffer bytes.Buffer

	if yaml {
		for _, obj := range kubeObjects {
			bytes, _ := json.MarshalIndent(obj, "", "  ")
			y, _ := gyaml.JSONToYAML(bytes)
			buffer.Write(y)
			buffer.WriteString("---\n")
		}
		mergedObject = buffer.String()
	} else {
		// Generate objects as Kind list
		// for _, obj := range kubeObjects {
		// }
	}
	return mergedObject
}

func generateK8sSpecFile(kubeObjects []interface{}, fileName string, yaml bool) {
	if yaml {
		k := mergeKubeObjects(kubeObjects, yaml)
		fileName = fileName + ".yaml"
		fmt.Println("Writing K8s Deployment YAML file to : ", fileName)
		_ = ioutil.WriteFile(fileName, []byte(k), 0644)
	} else {
	}
}

func createK8sSecret(secretName string, data map[string][]byte, namespace string) apiv1.Secret {
	secret := apiv1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: apiv1.SchemeGroupVersion.String(), Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: data,
	}
	return secret
}

func getValueFromSecretsManager(secretId string) string {
	var secretCache, _ = secretcache.New()
	secretValue, _ := secretCache.GetSecretString(secretId)
	return secretValue
}

func parseSecret(secretArn string) (string, string, string) {
	var secretName, secretValue, secretJsonKey string
	s := strings.Split(secretArn, ":")
	secretType := s[2]

	if secretType == "secretsmanager" {
		secretJsonKey = s[7]
		secretName = strings.ToLower(sanitizeValue(s[6], envSpecialChars, "-")) // K8s secret names can be only - lowercase alnum, '-', '.'
		secretValue = getValueFromSecretsManager(strings.Join(s[:7], ":"))
	} else {
		// Support values from AWS Systems Manager
	}

	return secretName, secretValue, secretJsonKey
}

func getK8Spec() {}

func generateTaskDefinition() {}
