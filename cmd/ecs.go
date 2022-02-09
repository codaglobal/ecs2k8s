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
	"os"

	"github.com/spf13/cobra"
)

// ecsCmd represents the root ecs command
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
	rootCmd.AddCommand(ecsCmd)
}
