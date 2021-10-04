// Copyright 2019 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"

	"github.com/fairwindsops/goldilocks/pkg/summary"
)

var deployment string
var container string

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "")
	generateCmd.PersistentFlags().StringVarP(&deployment, "deployment", "d", "", "")
	generateCmd.PersistentFlags().StringVarP(&container, "container", "c", "", "")
}

var generateCmd = &cobra.Command{
	Use:       "generate QOS",
	Short:     `Generate the resources YAML`,
	Long:      "Generate the resources YAML for a given container and QoS (burstable/guaranteed).",
	Args:      cobra.ExactValidArgs(1),
	ValidArgs: []string{"burstable", "guaranteed"},
	Run: func(cmd *cobra.Command, args []string) {
		qos := args[0]

		summarizer := summary.NewSummarizer(
			summary.ForNamespace(namespace),
			summary.WithFilter(qos),
		)
		data, err := summarizer.GetSummary()
		if err != nil {
			klog.Fatalf("Error getting summary: %v", err)
		}

		containerSummary := data.Namespaces[namespace].Deployments[deployment].Containers[container]
		if containerSummary.ContainerName != container {
			os.Stderr.WriteString(fmt.Sprintf("Container %s already matches %s recommendations\n", container, qos))
			os.Exit(1)
		}

		var requestResource v1.ResourceList
		var limitResource v1.ResourceList
		if qos == "guaranteed" {
			requestResource = containerSummary.Target
			limitResource = containerSummary.Target
		} else if qos == "burstable" {
			requestResource = containerSummary.LowerBound
			limitResource = containerSummary.UpperBound
		}

		fmt.Printf(`resources:
  requests:
    cpu: %s
    memory: %s
  limits:
    cpu: %s
    memory: %s`,
			requestResource.Cpu(),
			requestResource.Memory(),
			limitResource.Cpu(),
			limitResource.Memory(),
		)

	},
}
