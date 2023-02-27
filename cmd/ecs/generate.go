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
	"encoding/base64"
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
		d := generateDeploymentObject(td, rCount, namespace, false)

		var kubeObjects []interface{}

		kubeObjects = append(kubeObjects, &d)
		kubeObjects = append(kubeObjects, secrets)

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

var secrets []apiv1.Secret

// Generate K8s deployment object
func generateDeploymentObject(output ecs.DescribeTaskDefinitionOutput, rCount int32, namespace string, apply bool) appsv1.Deployment {
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
			// secretData := make(map[string][]byte)
			envVarName := sanitizeValue(*ecsSecret.Name, envSpecialChars, "")

			secretName, secretKey, secretValue := parseSecret(*ecsSecret.ValueFrom)

			generateK8sSecret(secretName, secretValue, namespace, apply)

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

	if apply {
		createKubeDeployment(deployment)
	}
	return *deployment
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
		// list := corev1.List{
		// 	TypeMeta: metav1.TypeMeta{
		// 		Kind:       "List",
		// 		APIVersion: "v1",
		// 	},
		// 	ListMeta: metav1.ListMeta{},
		// }

		// for _, obj := range kubeObjects {

		// 	list.Items = append(list.Items, runtime.RawExtension{Object: obj})

		// }
		// fmt.Println(jsonKubeObjects)
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

func generateK8sSecret(secretName string, data map[string][]byte, namespace string, apply bool) {
	// Check if K8s secret exists already and then create
	var secretExists bool = false

	for i := range secrets {
		if secrets[i].ObjectMeta.Name == secretName {
			secretExists = true
			break
		}
	}

	if !secretExists {
		secret := apiv1.Secret{
			TypeMeta: metav1.TypeMeta{APIVersion: apiv1.SchemeGroupVersion.String(), Kind: "Secret"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: namespace,
			},
			Data: data,
		}
		if apply {
			createKubeSecret(&secret)
		}
		secrets = append(secrets, secret)
	}
}

func getValueFromSecretsManager(secretId string) map[string][]byte {
	var jsonMap map[string][]byte
	var transformedMap map[string][]byte = make(map[string][]byte)
	var secretCache, _ = secretcache.New()
	secretValue, _ := secretCache.GetSecretString(secretId)

	if secretValue == "" {
		fmt.Println("Empty value returned for specified secret ID. Check if secret exists in this account.")
	}

	json.Unmarshal([]byte(secretValue), &jsonMap)
	for index, v := range jsonMap {
		s := base64.StdEncoding.EncodeToString(v)
		transformedMap[index] = []byte(s)
	}
	return transformedMap
}

func parseSecret(secretArn string) (string, string, map[string][]byte) {
	var secretName, secretJsonKey string
	var secretValue map[string][]byte
	s := strings.Split(secretArn, ":")

	switch secretType := s[2]; secretType {
	case "secretsmanager":
		secretName = strings.ToLower(sanitizeValue(s[6], envSpecialChars, "-")) // K8s secret names can be only - lowercase alnum, '-', '.'
		secretJsonKey = s[7]
		if secretJsonKey == "" {
			fmt.Println("Secret JSON key is required in K8s spec")
			os.Exit(1)
		}
		secretValue = getValueFromSecretsManager(strings.Join(s[:7], ":"))

		// generateK8sSecret(secretName, secretValue, "default")
	case "systemsmanager":
		// Support for secrets from AWS Systems Manager
	}

	return secretName, secretJsonKey, secretValue
}

func getK8Spec() {}

func generateTaskDefinition() {}
