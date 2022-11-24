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
package k8sCmd

import (
	"os"

	root "codaglobal/ecs2k8s/cmd/root"

	"github.com/spf13/cobra"
)

// k8sCmd represents the root k8s command
var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "A set of commands to work with an K8s Cluster.",
	Long:  `A set of commands to work with an K8s cluster. These subcommands can list, generate and migrate a task definition to a K8s cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

func init() {
	root.RootCmd.AddCommand(k8sCmd)
	k8sCmd.PersistentFlags().String("deployment", "", "A valid deployment in K8s")
	k8sCmd.PersistentFlags().String("container-name", "", "Name of the container inside the task, if more than one container is specified in that task")
	k8sCmd.PersistentFlags().StringP("namespace", "n", "", "The Kubernetes namespace in which the deployment needs to be created")
	k8sCmd.PersistentFlags().String("file-name", "", "The file into which K8s spec will be written to, defaults to datetime of spec generation")
	k8sCmd.PersistentFlags().Bool("yaml", false, "Set this flag if spec file needs to generated in YAML, defaults to JSON")
	k8sCmd.PersistentFlags().String("kubeconfig", "", "Config file for the K8s cluster, if parameter is not passed, checks $HOME/.kube directory and then KUBECONFIG environment variable")
	k8sCmd.PersistentFlags().Int32("replicas", 1, "The replica count for the K8s deployment")
}
