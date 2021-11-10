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

type kubeMetadata struct {
	Name string `json:"name"`
}

type kubeContainer struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	Ports []struct {
		ContainerPort int `json:"containerPort"`
	} `json:"ports"`
}

type Label struct {
	Name string
}
type kubeSpec struct {
	Selector struct {
		MatchLabels struct {
			Labels []struct{} `json:"labels"`
		} `json:"matchLabels"`
	} `json:"selector"`
	Replicas int `json:"replicas"`
	Template struct {
		MetaData struct {
			Label []string `json:"labels"`
		} `json:"metadata"`
		Spec struct {
			Containers []kubeContainer `json:"containers"`
		} `json:"spec"`
	} `json:"Template"`
}

type kubeObject struct {
	ApiVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind, omitempty"`
	MetaData   kubeMetadata `json:"metadata"`
	Spec       kubeSpec     `json:"spec"`
}
