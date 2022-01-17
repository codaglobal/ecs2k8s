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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate ECS cluster to the k8s cluster.",
	Long: `Migrate ECS cluster to the k8s cluster. For example:	`,
	Run: func(cmd *cobra.Command, args []string) {
		taskDefintion, _ := cmd.Flags().GetString("task")
		fileName, _ := cmd.Flags().GetString("fname")
		rCount, _ := cmd.Flags().GetInt("rcount")

		if fileName == "" {
			fileName = getDefaultFileName()
		}

		if taskDefintion == "" {
			fmt.Println("Task definition required")
		}

		td := getTaskDefiniton(taskDefintion, fileName)
		d := generateDeploymentJSON(td, fileName, rCount)
		createKubeDeployment(d)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().String("task", "", "A valid task definition in ECS")
	migrateCmd.Flags().String("container-name", "", "Name of the container inside the task, if more than one container is specified in that task")
	migrateCmd.PersistentFlags().Int("replicas", 1, "Number of replicas")
}

// Creates a K8s deployment in the local K8s cluster
func createKubeDeployment(deployment *appsv1.Deployment) {
	var kubeconfig string

	home := homedir.HomeDir()
	localKubeconfig := filepath.Join(home, ".kube", "config")
	kubeConfigenv := os.Getenv("kubeconfig")

	if kubeConfigenv != "" {
		_, err := exists(kubeConfigenv)
		if err != nil {
			fmt.Println("Directory provided in kubeconfig does not exist.")
			os.Exit(1)
		}
		kubeconfig = kubeConfigenv
	} else {
		kubeconfig = localKubeconfig
	}

	fmt.Println("Using kubeconfig provided in", kubeconfig)

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	fmt.Print("Proceed with deploying: ", deployment.ObjectMeta.Name, " (yes/no): ")

	deploy := askForConfirmation()
	if !deploy {
		fmt.Println("Operation cancelled by user")
		return
	}
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	fmt.Println("Creating deployment...")

	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})

	if err != nil {
		log.Println("Deployment failed", err)
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

}
