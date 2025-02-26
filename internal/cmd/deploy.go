// Copyright © 2018 The Havener
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/homeport/gonvenience/pkg/v1/bunt"
	"github.com/homeport/gonvenience/pkg/v1/wait"
	"github.com/homeport/havener/pkg/havener"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

const (
	envVarDeployConfig  = "deployment_config"
	envVarDeployTimeout = "deployment_timeout"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:           "deploy",
	Short:         "Deploy to Kubernetes",
	Long:          `Deploy to Kubernetes based on havener configuration`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		havenerConfig := viper.GetString(envVarDeployConfig)

		switch {
		case len(havenerConfig) > 0:
			return DeployViaHavenerConfig(havenerConfig)

		default:
			cmd.Usage()
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.PersistentFlags().String("config", "", "havener configuration file")
	deployCmd.PersistentFlags().Int("timeout", 40, "deployment timeout in minutes")

	viper.AutomaticEnv()
	viper.BindPFlag(envVarDeployConfig, deployCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag(envVarDeployTimeout, deployCmd.PersistentFlags().Lookup("timeout"))
}

// DeployViaHavenerConfig entry function for running a helm install
func DeployViaHavenerConfig(havenerConfig string) error {
	timeoutInMin := viper.GetInt(envVarDeployTimeout)

	cfgdata, err := ioutil.ReadFile(havenerConfig)
	if err != nil {
		return &ErrorWithMsg{"unable to read havener configuration", err}
	}

	var config havener.Config
	if err := yaml.Unmarshal(cfgdata, &config); err != nil {
		return &ErrorWithMsg{"failed to unmarshal havener configuration", err}
	}

	if err := processTask("Predeployment Steps", config.Before); err != nil {
		return &ErrorWithMsg{"failed to evaluate predeployment steps", err}
	}

	for _, release := range config.Releases {
		overrides, err := havener.TraverseStructureAndProcessShellOperators(release.Overrides)
		if err != nil {
			return &ErrorWithMsg{"failed to process overrides section", err}
		}

		if err := processTask("Before Chart "+release.ChartName, release.Before); err != nil {
			return &ErrorWithMsg{"failed to evaluate before release steps", err}
		}

		overridesData, err := yaml.Marshal(overrides)
		if err != nil {
			return &ErrorWithMsg{"failed to marshal overrides structure into bytes", err}
		}

		pi := wait.NewProgressIndicator(fmt.Sprintf("Creating Helm Release for %s", release.ChartName))
		pi.SetTimeout(time.Duration(timeoutInMin) * time.Minute)
		err = havener.DeployHelmRelease(
			release.ChartName,
			release.ChartNamespace,
			release.ChartLocation,
			timeoutInMin,
			overridesData)

		pi.Stop()

		if err != nil {
			return &ErrorWithMsg{"failed to deploy via havener configuration", err}
		}

		message := bunt.Sprintf("Successfully created new helm chart *%s* in namespace *_%s_*.",
			release.ChartName,
			release.ChartNamespace,
		)

		releaseNotes, err := havener.RunHelmBinary("get", "notes", release.ChartName)
		if err != nil {
			return &ErrorWithMsg{"failed to get notes of release", err}
		}

		if releaseNotes != nil {
			message = message + "\n\n" + string(releaseNotes)
		}

		printStatusMessage("Upgrade", message, bunt.Gray)

		if err := processTask("After Chart "+release.ChartName, release.After); err != nil {
			return &ErrorWithMsg{"failed to evaluate after release steps", err}
		}
	}

	if err := processTask("Postdeployment Steps", config.After); err != nil {
		return &ErrorWithMsg{"failed to evaluate postdeployment steps", err}
	}
	return nil
}
