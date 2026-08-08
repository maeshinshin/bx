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
		filePath     string
		useDirFile   bool
		useNullStdin bool
		useDirStdin  bool
		useK8sFlag   bool
		k8sFromFile  string
		k8sFromArg   string
		k8sFromStdin string
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
			name:        "9. Successful k8s secret decode from argument",
			args:        []string{"decode", "-k", "data:\n  username: YWRtaW4=\n"},
			useK8sFlag:  true,
			expectedOut: "username: admin\n",
		},
		{
			name:        "10. Successful k8s secret decode from file",
			args:        []string{"decode", "-k"},
			useK8sFlag:  true,
			k8sFromFile: "data:\n  username: YWRtaW4=\n",
			expectedOut: "username: admin\n",
		},
		{
			name:         "11. Successful k8s secret decode from stdin",
			args:         []string{"decode", "-k"},
			useK8sFlag:   true,
			k8sFromStdin: "data:\n  username: YWRtaW4=\n",
			expectedOut:  "username: admin\n",
		},
		{
			name:        "12. k8s decode returns error when YAML is invalid",
			args:        []string{"decode", "-k", "data: : :\n  bad: : :\n"},
			useK8sFlag:  true,
			expectedErr: true,
		},
		{
			name:        "13. Successful URL-safe decode round trip (Pj4_ -> >>?)",
			args:        []string{"decode", "-u", "Pj4_"},
			expectedOut: ">>?\n",
		},
		{
			name:        "14. Successful URL-safe decode of URL alphabet (a-b_ accepted)",
			args:        []string{"decode", "-u", "a-b_"},
			expectedOut: "k\xE6\xFF\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileFlag = ""
			k8sFlag = false
			urlFlag = false
			testArgs := tt.args

			// Prepare a file path arg for k8sFromFile scenarios.
			if tt.useK8sFlag && tt.k8sFromFile != "" {
				tmpfile, err := os.CreateTemp("", "bx-k8s-*")
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(tmpfile.Name())
				if _, err := tmpfile.WriteString(tt.k8sFromFile); err != nil {
					t.Fatal(err)
				}
				tmpfile.Close()
				testArgs = []string{"decode", "-k", "-f", tmpfile.Name()}
				tt.filePath = tmpfile.Name()
			}

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
				} else if tt.k8sFromStdin != "" {
					wIn.WriteString(tt.k8sFromStdin)
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
