package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/maeshinshin/bx/k8s"
)

var fileFlag string
var k8sFlag bool

var decodeCmd = &cobra.Command{
	Use:     "decode [base64_string]",
	Aliases: []string{"d"},
	Short:   "Decodes a Base64 encoded string or Kubernetes Secret YAML",
	RunE: func(cmd *cobra.Command, args []string) error {
		var inputBytes []byte
		var err error

		if len(args) > 0 {
			inputBytes = []byte(args[0])
		} else if fileFlag != "" {
			file, err := os.Open(fileFlag)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer func() { _ = file.Close() }()

			inputBytes, err = io.ReadAll(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
		} else {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return fmt.Errorf("no input provided")
			}

			inputBytes, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
		}

		if k8sFlag {
			output, err := k8s.DecodeSecretYAML(inputBytes)
			if err != nil {
				return err
			}
			fmt.Print(string(output))
			return nil
		}

		inputStr := strings.TrimSpace(string(inputBytes))
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
	decodeCmd.Flags().BoolVarP(&k8sFlag, "k8s", "k", false, "extract and decode data fields from Kubernetes Secret YAML")
}
