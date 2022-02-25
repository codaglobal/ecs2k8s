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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate-task",
	Short: "Migrate ECS cluster to the k8s cluster.",
	Long: `Migrate ECS cluster to the k8s cluster. For example:	`,
	Run: func(cmd *cobra.Command, args []string) {
		taskDefintion, _ := cmd.Flags().GetString("task-definition")
		rCount, _ := cmd.Flags().GetInt32("replicas")
		namespace, _ := cmd.Flags().GetString("namespace")
		kubeConfigFile, _ := cmd.Flags().GetString("kubeconfig")

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
		createKubeDeployment(d, kubeConfigFile)
	},
}

func init() {
	ecsCmd.AddCommand(migrateCmd)
}

// Creates a K8s deployment in the local K8s cluster
func createKubeDeployment(deployment *appsv1.Deployment, kubeConfigParamter string) {
	var kubeconfig string
	home := homedir.HomeDir()
	localKubeconfig := filepath.Join(home, ".kube", "config")
	kubeConfigenv := os.Getenv("KUBECONFIG")

	// Checks for config file in parameter passed, then in $HOME/.kube directory and then KUBECONFIG environment variable
	if kubeConfigParamter != "" {
		if _, err := os.Stat(kubeConfigParamter); err != nil {
			fmt.Println("No valid kubeconfig found in the specified location,", kubeConfigParamter)
			os.Exit(1)
		}
		kubeconfig = kubeConfigParamter
	} else if _, err := os.Stat(localKubeconfig); err == nil {
		kubeconfig = localKubeconfig
	} else {
		if _, err := os.Stat(kubeConfigenv); err == nil {
			kubeconfig = kubeConfigenv
		} else {
			fmt.Println("Unable to detect kubeconfig in default location.")
			os.Exit(1)
		}
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
	deploymentsClient := clientset.AppsV1().Deployments(deployment.ObjectMeta.Namespace)

	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})

	if err != nil {
		log.Println("Deployment failed", err)
		panic(err)
	}

	fmt.Printf("Submitted new deployment %q.\n", result.GetObjectMeta().GetName())

}
