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
	"bytes"
	"io"
	"os"
	"testing"
)

func TestDecodeCmd(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		stdin        string
		fileContent  string
		useDirFile   bool
		useNullStdin bool
		useDirStdin  bool
		expectedOut  string
		expectedErr  bool
	}{
		{
			name:        "1. Successful decode from argument",
			args:        []string{"decode", "aG9nZQ=="},
			expectedOut: "hoge\n",
		},
		{
			name:        "2. Successful decode from stdin (pipe)",
			args:        []string{"decode"},
			stdin:       "aG9nZQ==",
			expectedOut: "hoge\n",
		},
		{
			name:        "3. Successful decode from file",
			args:        []string{"decode"},
			fileContent: "aG9nZQ==",
			expectedOut: "hoge\n",
		},
		{
			name:        "4. Invalid Base64 string (should return error)",
			args:        []string{"decode", "invalid_base64!!!"},
			expectedErr: true,
		},
		{
			name:        "5. Failed to open file (non-existent path)",
			args:        []string{"decode", "-f", "/non/existent/path/to/bx-test"},
			expectedErr: true,
		},
		{
			name:        "6. Failed to read from file (directory)",
			args:        []string{"decode"},
			useDirFile:  true,
			expectedErr: true,
		},
		{
			name:         "7. No input provided (stdin is character device)",
			args:         []string{"decode"},
			useNullStdin: true,
			expectedErr:  true,
		},
		{
			name:        "8. Failed to read from stdin (directory used as stdin)",
			args:        []string{"decode"},
			useDirStdin: true,
			expectedErr: true,
		},
		{
			name:        "9. [Color Check] Intentionally failing test",
			args:        []string{"decode", "aG9nZQ=="},
			expectedOut: "hoge\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileFlag = ""
			testArgs := tt.args

			if tt.fileContent != "" {
				tmpfile, err := os.CreateTemp("", "bx-test-*")
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(tmpfile.Name())
				tmpfile.WriteString(tt.fileContent)
				tmpfile.Close()
				testArgs = append(testArgs, "-f", tmpfile.Name())
			} else if tt.useDirFile {
				tmpdir, err := os.MkdirTemp("", "bx-test-dir-*")
				if err != nil {
					t.Fatal(err)
				}
				defer os.RemoveAll(tmpdir)
				testArgs = append(testArgs, "-f", tmpdir)
			}

			oldStdout := os.Stdout
			rOut, wOut, _ := os.Pipe()
			os.Stdout = wOut

			oldStdin := os.Stdin
			switch {
			case tt.useNullStdin:
				nullFile, err := os.Open(os.DevNull)
				if err != nil {
					t.Fatal(err)
				}
				os.Stdin = nullFile
				defer nullFile.Close()
			case tt.useDirStdin:
				tmpdir, err := os.MkdirTemp("", "bx-test-stdin-*")
				if err != nil {
					t.Fatal(err)
				}
				defer os.RemoveAll(tmpdir)
				dirFile, err := os.Open(tmpdir)
				if err != nil {
					t.Fatal(err)
				}
				os.Stdin = dirFile
				defer dirFile.Close()
			default:
				rIn, wIn, _ := os.Pipe()
				os.Stdin = rIn
				if tt.stdin != "" {
					wIn.WriteString(tt.stdin)
				}
				wIn.Close()
			}

			rootCmd.SetArgs(testArgs)
			err := rootCmd.Execute()

			wOut.Close()
			os.Stdout = oldStdout
			os.Stdin = oldStdin

			var outBuf bytes.Buffer
			io.Copy(&outBuf, rOut)
			actualOut := outBuf.String()

			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error occurred: %v", err)
				}
				assertColorDiff(t, tt.expectedOut, actualOut)
			}
		})
	}
}

func assertColorDiff(t *testing.T, expected, actual string) {
	t.Helper()

	if expected != actual {
		green := "\x1b[32m"
		red := "\x1b[31m"
		reset := "\x1b[0m"

		t.Errorf("Output mismatch:\n%s[Expected]%s\n%q\n%s[Actual]%s\n%q\n",
			green, reset, expected,
			red, reset, actual,
		)
	}
}
