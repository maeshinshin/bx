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
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// DecodeSecretYAML parses a Kubernetes Secret YAML, extracts the 'data' and
// 'stringData' fields, decodes the Base64 values in 'data', and returns them
// in a "key: value" format. Values in 'stringData' are emitted verbatim.
func DecodeSecretYAML(input []byte) ([]byte, error) {
	file, err := parser.ParseBytes(input, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	var buf bytes.Buffer

	for _, doc := range file.Docs {
		if doc.Body != nil {
			extractSecretData(&buf, doc.Body)
			extractSecretStringData(&buf, doc.Body)
		}
	}

	return buf.Bytes(), nil
}

// extractSecretData walks the YAML AST, finds the "data" map, and writes
// decoded key-value pairs to the buffer.
func extractSecretData(buf *bytes.Buffer, node ast.Node) {
	extractSecretField(buf, node, "data", true)
}

// extractSecretStringData walks the YAML AST, finds the "stringData" map,
// and writes its plain key-value pairs to the buffer.
func extractSecretStringData(buf *bytes.Buffer, node ast.Node) {
	extractSecretField(buf, node, "stringData", false)
}

// extractSecretField walks the YAML AST, finds the named map, and writes
// its key-value pairs to the buffer. When decode is true, each value is
// treated as Base64 and decoded; on failure the raw value is emitted.
func extractSecretField(buf *bytes.Buffer, node ast.Node, key string, decode bool) {
	mapping, ok := node.(*ast.MappingNode)
	if !ok {
		return
	}

	for _, mapVal := range mapping.Values {
		keyNode, keyOk := mapVal.Key.(*ast.StringNode)
		if !keyOk || keyNode.Value != key {
			continue
		}

		valMapping, valOk := mapVal.Value.(*ast.MappingNode)
		if !valOk {
			continue
		}

		for _, dataVal := range valMapping.Values {
			kNode, kOk := dataVal.Key.(*ast.StringNode)
			vNode, vOk := dataVal.Value.(*ast.StringNode)
			if !kOk || !vOk {
				continue
			}
			value := vNode.Value
			if decode {
				if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
					value = string(decoded)
				}
			}
			fmt.Fprintf(buf, "%s: %s\n", kNode.Value, value)
		}
	}
}
