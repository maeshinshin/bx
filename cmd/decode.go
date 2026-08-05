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
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var fileFlag string

// decodeCmd represents the decode command
var decodeCmd = &cobra.Command{
	Use:     "decode [base64_string]",
	Aliases: []string{"d"},
	Short:   "Decodes a Base64 encoded string",
	RunE: func(cmd *cobra.Command, args []string) error {
		var inputStr string
		if len(args) > 0 {
			inputStr = args[0]
		} else if fileFlag != "" {
			file, err := os.Open(fileFlag)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()

			bytes, err := io.ReadAll(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			inputStr = string(bytes)
		} else {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return fmt.Errorf("no input provided. Please provide an argument, pipe data, or use the -f flag")
			}

			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			inputStr = string(bytes)
		}
		inputStr = strings.TrimSpace(inputStr)

		decoded, err := base64.StdEncoding.DecodeString(inputStr)
		if err != nil {
			return fmt.Errorf("failed to decode base64 string: %w", err)
		}

		fmt.Println(string(decoded))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(decodeCmd)
	decodeCmd.Flags().StringVarP(&fileFlag, "file", "f", "", "specify the file to read from")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// decodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// decodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
