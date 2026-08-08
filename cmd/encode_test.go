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

func TestEncodeCmd(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		stdin        string
		fileContent  string
		filePath     string
		useDirFile   bool
		useNullStdin bool
		useDirStdin  bool
		expectedOut  string
		expectedErr  bool
	}{
		{
			name:        "1. Successful encode from argument",
			args:        []string{"encode", "hoge"},
			expectedOut: "aG9nZQ==\n",
		},
		{
			name:        "2. Successful encode from stdin (pipe)",
			args:        []string{"encode"},
			stdin:       "hoge",
			expectedOut: "aG9nZQ==\n",
		},
		{
			name:        "3. Successful encode from file",
			args:        []string{"encode"},
			fileContent: "hoge",
			expectedOut: "aG9nZQ==\n",
		},
		{
			name:        "4. Successful encode with newline from argument",
			args:        []string{"encode", "hoge\n"},
			expectedOut: "aG9nZQo=\n",
		},
		{
			name:        "5. Successful encode empty string from argument",
			args:        []string{"encode", ""},
			expectedOut: "\n",
		},
		{
			name:        "6. Failed to open file (non-existent path)",
			args:        []string{"encode", "-f", "/non/existent/path/to/bx-test"},
			expectedErr: true,
		},
		{
			name:        "7. Failed to read from file (directory)",
			args:        []string{"encode"},
			useDirFile:  true,
			expectedErr: true,
		},
		{
			name:         "8. No input provided (stdin is character device)",
			args:         []string{"encode"},
			useNullStdin: true,
			expectedErr:  true,
		},
		{
			name:        "9. Failed to read from stdin (directory used as stdin)",
			args:        []string{"encode"},
			useDirStdin: true,
			expectedErr: true,
		},
		{
			name:        "10. Successful URL-safe encode from argument (>>? -> Pj4_)",
			args:        []string{"encode", "-u", ">>?"},
			expectedOut: "Pj4_\n",
		},
		{
			name:        "11. Successful URL-safe encode of input without special chars (hoge)",
			args:        []string{"encode", "-u", "hoge"},
			expectedOut: "aG9nZQ==\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileFlag = ""
			urlFlag = false
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
				if tt.expectedOut != actualOut {
					t.Errorf("Output mismatch:\nExpected:\n%q\nActual:\n%q", tt.expectedOut, actualOut)
				}
			}
		})
	}
}
