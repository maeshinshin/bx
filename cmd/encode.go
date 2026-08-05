package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var encodeCmd = &cobra.Command{
	Use:     "encode [string]",
	Aliases: []string{"e"},
	Short:   "Encodes a string to Base64",
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
			defer file.Close()

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

		encoded := base64.StdEncoding.EncodeToString(inputBytes)
		fmt.Println(encoded)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(encodeCmd)
	encodeCmd.Flags().StringVarP(&fileFlag, "file", "f", "", "specify the file to read from")
}