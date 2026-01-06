package differ

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gookit/color"
	"github.com/pmezard/go-difflib/difflib"
	"gopkg.in/yaml.v3"
)

// Options holds configuration for the diff operation.
type Options struct {
	// SecureMode enables masking of sensitive data in Secrets and ConfigMaps.
	SecureMode bool
	// IncludeKinds filters resources to only include specific Kinds (case-insensitive).
	// If empty, all resources are included.
	IncludeKinds []string
	// ExcludeKinds filters resources to exclude specific Kinds (case-insensitive).
	// If empty, no resources are excluded.
	ExcludeKinds []string
}

// Diff compares two YAML byte slices and returns a human-readable diff.
func Diff(fileA, fileB []byte, opts Options) (string, error) {
	docsA, err := decodeDocs(fileA)
	if err != nil {
		return "", fmt.Errorf("failed to decode first file: %w", err)
	}

	docsB, err := decodeDocs(fileB)
	if err != nil {
		return "", fmt.Errorf("failed to decode second file: %w", err)
	}

	// Filter resources based on IncludeKinds and ExcludeKinds
	if len(opts.IncludeKinds) > 0 || len(opts.ExcludeKinds) > 0 {
		docsA = filterResources(docsA, opts.IncludeKinds, opts.ExcludeKinds)
		docsB = filterResources(docsB, opts.IncludeKinds, opts.ExcludeKinds)
	}

	// Mask Sensitive Data
	if opts.SecureMode {
		maskSensitiveData(docsA, DefaultMaskingRules())
		maskSensitiveData(docsB, DefaultMaskingRules())
	}

	// Normal Mode & Secure Mode (now that data is safe): Normalize and Diff
	// Note: We masked the data IN PLACE in the map structures for SecureMode.
	// So we can proceed to diff normally. The "Same Length Hash Suffix" strategy
	// ensures changes are detected by the diff engine.

	yamlA, err := marshalDocs(docsA)
	if err != nil {
		return "", fmt.Errorf("failed to normalize first file: %w", err)
	}

	yamlB, err := marshalDocs(docsB)
	if err != nil {
		return "", fmt.Errorf("failed to normalize second file: %w", err)
	}

	// Compute Raw Diff
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(yamlA),
		B:        difflib.SplitLines(yamlB),
		FromFile: "Original",
		ToFile:   "Modified",
		Context:  3,
	}

	text, _ := difflib.GetUnifiedDiffString(diff)
	if text == "" {
		return "# No Changes", nil
	}

	// Colorize
	return colorizeDiff(text), nil
}

// decodeDocs parses a byte slice that may contain multiple YAML documents.
func decodeDocs(data []byte) ([]interface{}, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var docs []interface{}

	for {
		var doc interface{}
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// marshalDocs encodes a slice of documents back into a single YAML string.
func marshalDocs(docs []interface{}) (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	for _, doc := range docs {
		if err := enc.Encode(doc); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

// colorizeDiff adds ANSI color codes to the diff output using gookit/color.
// It parses the unified diff text and applies colors line by line.
func colorizeDiff(text string) string {
	lines := strings.Split(text, "\n")
	var colored []string
	for _, line := range lines {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			// File headers - keeping them plain or bold
			colored = append(colored, color.Bold.Sprint(line))
		} else if strings.HasPrefix(line, "@@") {
			// Chunk headers
			colored = append(colored, color.Cyan.Sprint(line))
		} else if strings.HasPrefix(line, "+") {
			// Green for additions
			colored = append(colored, color.Green.Sprint(line))
		} else if strings.HasPrefix(line, "-") {
			// Red for deletions
			colored = append(colored, color.Red.Sprint(line))
		} else {
			// Context lines
			colored = append(colored, line)
		}
	}
	return strings.Join(colored, "\n")
}
