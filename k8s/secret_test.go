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
package k8s

import (
	"strings"
	"testing"
)

func TestDecodeSecretYAML(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectedErr bool
	}{
		{
			name: "successful decode of single secret data field",
			input: `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
data:
  username: YWRtaW4=
  password: cEBzc3dvcmQ=
`,
			expected: "username: admin\npassword: p@ssword\n",
		},
		{
			name: "data field with invalid base64 keeps the raw value",
			input: `data:
  key: not_valid_base64!!!
`,
			expected: "key: not_valid_base64!!!\n",
		},
		{
			name: "no data field returns empty output",
			input: `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
`,
			expected: "",
		},
		{
			name: "value that is not a mapping is ignored",
			input: `data: not-a-mapping
`,
			expected: "",
		},
		{
			name: "multiple documents are processed",
			input: `data:
  a: YQ==
---
data:
  b: Yg==
`,
			expected: "a: a\nb: b\n",
		},
		{
			name: "non-mapping data values are skipped",
			input: `data:
  username: YWRtaW4=
  list:
    - foo
    - bar
`,
			expected: "username: admin\n",
		},
		{
			name: "non-mapping root document is ignored",
			input: `- foo
- bar
`,
			expected: "",
		},
		{
			name:     "empty input is valid",
			input:    "",
			expected: "",
		},
		{
			name:        "invalid YAML returns parse error",
			input:       "data: : :\n  bad: : :\n",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeSecretYAML([]byte(tt.input))
			if tt.expectedErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.expected {
				t.Errorf("output mismatch:\nexpected:\n%q\nactual:\n%q", tt.expected, string(got))
			}
		})
	}
}

func TestDecodeSecretYAML_EmptyDocBody(t *testing.T) {
	// input that parses but produces no body
	input := "---\n"
	got, err := DecodeSecretYAML([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty output, got %q", string(got))
	}
}

func TestDecodeSecretYAML_TopLevelNonMappingWithDataInside(t *testing.T) {
	// Verify extractSecretData early-returns when the doc body is not a mapping.
	input := "- just\n- a\n- list\n"
	got, err := DecodeSecretYAML([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty output, got %q", string(got))
	}
}

func TestDecodeSecretYAML_KeyNotDataIsIgnored(t *testing.T) {
	input := `stringData:
  username: YWRtaW4=
`
	got, err := DecodeSecretYAML([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty output for non-data key, got %q", string(got))
	}
}

func TestDecodeSecretYAML_DataValueIsSequence(t *testing.T) {
	// value under "data" that is a sequence rather than a mapping -> ignored
	input := `data:
  - foo
  - bar
`
	got, err := DecodeSecretYAML([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty output, got %q", string(got))
	}
}

func TestDecodeSecretYAML_OutputIncludesTrailingNewline(t *testing.T) {
	input := `data:
  username: YWRtaW4=
`
	got, err := DecodeSecretYAML([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(string(got), "\n") {
		t.Errorf("expected trailing newline, got %q", string(got))
	}
}
