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
	"os"

	root "codaglobal/ecs2k8s/cmd/root"

	"github.com/spf13/cobra"
)

// ecsCmd repr esents the root ecs command
var ecsCmd = &cobra.Command{
	Use:   "ecs",
	Short: "A set of commands to work with an existing AWS ECS cluster.",
	Long:  `A set of commands to work with an existing AWS ECS cluster. These subcommands can list, generate and migrate a task definition to a K8s cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

func init() {
	root.RootCmd.AddCommand(ecsCmd)
	ecsCmd.PersistentFlags().String("task-definition", "", "A valid task definition in ECS")
	ecsCmd.PersistentFlags().String("container-name", "", "Name of the container inside the task, if more than one container is specified in that task")
	ecsCmd.PersistentFlags().StringP("namespace", "n", "", "The Kubernetes namespace in which the deployment needs to be created")
	ecsCmd.PersistentFlags().String("file-name", "", "The file into which K8s spec will be written to, defaults to datetime of spec generation")
	ecsCmd.PersistentFlags().Bool("yaml", false, "Set this flag if spec file needs to generated in YAML, defaults to JSON")
	ecsCmd.PersistentFlags().String("kubeconfig", "", "Config file for the K8s cluster, if parameter is not passed, checks $HOME/.kube directory and then KUBECONFIG environment variable")
	ecsCmd.PersistentFlags().Int32("replicas", 1, "The replica count for the K8s deployment")
}
