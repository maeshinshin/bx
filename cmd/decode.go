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
	Short:   "Decode a Base64 string or Kubernetes Secret YAML",
	Long: `Decode reads a Base64 encoded string and prints the decoded value to
stdout. With -k/--k8s, it instead reads a Kubernetes Secret YAML manifest and
prints every value under the "data" key as Base64-decoded "key: value" pairs.

Input is taken from the first argument, from a file passed with -f/--file,
or from stdin when no argument or file is given.`,
	Example: `  # Decode a Base64 string
  bx decode "aG9nZQ=="

  # Decode the contents of a file
  bx decode -f encoded.txt

  # Decode data piped from another command
  echo "aG9nZQ==" | bx decode

  # Decode every value in the data field of a Secret YAML
  bx decode -k -f secret.yaml

  # Pipe directly from kubectl
  kubectl get secret my-secret -o yaml | bx decode -k`,
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
