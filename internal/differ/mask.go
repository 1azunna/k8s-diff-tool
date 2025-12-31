package differ

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// MaskConfig defines which fields to mask for a specific Kind.
type MaskConfig struct {
	// RootKeys is a list of top-level keys within the resource that contain sensitive data.
	// Examples: []string{"data", "stringData", "binaryData"}
	RootKeys []string
}

// DefaultMaskingRules returns the hardcoded defaults for Secrets and ConfigMaps.
func DefaultMaskingRules() map[string]MaskConfig {
	return map[string]MaskConfig{
		"secret": {
			RootKeys: []string{"data", "stringData"},
		},
		"configmap": {
			RootKeys: []string{"data", "binaryData"},
		},
	}
}

// maskSensitiveData operates on the documents in-place, masking sensitive fields
// based on the provided configuration.
func maskSensitiveData(docs []interface{}, rules map[string]MaskConfig) {
	for _, doc := range docs {
		m, ok := doc.(map[string]interface{})
		if !ok {
			// Handle generic map if present
			if _, ok := doc.(map[interface{}]interface{}); ok {
				// Convert to string map wrapper just for checking kind?
				// Actually we need to traverse it. simpler to expect map[string]interface{}.
				// If we encounter generic map, we can try to mask it too if we find Kind.
				// But deeply nested masking in generic map is tricky.
				// Let's stick to standard behavior: our loader usually returns map[string]interface{}.
				continue
			}
			continue
		}

		kindVal, ok := m["kind"]
		if !ok {
			continue
		}
		kindStr, ok := kindVal.(string)
		if !ok {
			continue
		}
		
		normalizedKind := strings.ToLower(kindStr)
		
		// Check if we have a rule for this Kind
		if config, exists := rules[normalizedKind]; exists {
			for _, rootKey := range config.RootKeys {
				maskMap(m, rootKey)
			}
		}
	}
}

// maskMap replaces values in the map found at parent[key] with masked strings.
func maskMap(parent map[string]interface{}, key string) {
	v, ok := parent[key]
	if !ok {
		return
	}

	if dataMap, ok := v.(map[string]interface{}); ok {
		for k, val := range dataMap {
			parent[key].(map[string]interface{})[k] = generateMask(fmt.Sprintf("%v", val))
		}
		return
	}
	
	if dataMapGeneric, ok := v.(map[interface{}]interface{}); ok {
		for k, val := range dataMapGeneric {
			dataMapGeneric[k] = generateMask(fmt.Sprintf("%v", val))
		}
	}
}

// generateMask returns a string of equal length to input.
// It uses a hash suffix to preserve uniqueness (so changes are detected)
// while masking the content.
func generateMask(original string) string {
	length := len(original)
	if length == 0 {
		return ""
	}
	
	// Create a hash of the content
	hash := sha256.Sum256([]byte(original))
	hexHash := hex.EncodeToString(hash[:])

	// If short, just use hash characters (up to length)
	if length <= 8 {
		return hexHash[:length]
	}

	// If longer, prefix with stars and suffix with last 8 chars of hash
	// This ensures:
	// 1. Length is preserved.
	// 2. Diff engine sees different values for different inputs.
	// 3. Sensitive prefix is fully hidden.
	prefixLen := length - 8
	return strings.Repeat("*", prefixLen) + hexHash[:8]
}
