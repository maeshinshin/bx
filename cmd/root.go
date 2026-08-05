/*
Copyright © 2026 maeshinshin

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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bx",
	Short: "A Base64 encoder/decoder for strings and Kubernetes Secret YAML",
	Long: `bx encodes and decodes Base64 strings, and decodes the data field of
Kubernetes Secret YAML manifests without requiring kubectl.

Examples:
  # Encode a string
  bx encode "hoge"

  # Decode a Base64 string
  bx decode "aG9nZQ=="

  # Decode every value in the data field of a Secret YAML
  bx decode -k -f secret.yaml`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
