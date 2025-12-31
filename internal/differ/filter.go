package differ

import (
	"strings"
)

// filterResources filters documents based on inclusion and exclusion criteria.
// It performs a single pass over the documents.
func filterResources(docs []interface{}, includeKinds, excludeKinds []string) []interface{} {
	// If no filters are applied, return original docs
	if len(includeKinds) == 0 && len(excludeKinds) == 0 {
		return docs
	}

	// Prepare sets for O(1) loopups
	includeSet := make(map[string]bool)
	for _, k := range includeKinds {
		includeSet[strings.ToLower(k)] = true
	}

	excludeSet := make(map[string]bool)
	for _, k := range excludeKinds {
		excludeSet[strings.ToLower(k)] = true
	}

	var filtered []interface{}

	for _, doc := range docs {
		// Extract Kind
		var kind string
		foundKind := false

		if m, ok := doc.(map[string]interface{}); ok {
			if kVal, ok := m["kind"]; ok {
				if kStr, ok := kVal.(string); ok {
					kind = strings.ToLower(kStr)
					foundKind = true
				}
			}
		} else if mg, ok := doc.(map[interface{}]interface{}); ok {
			// Robust fallback
			if kVal, ok := mg["kind"]; ok {
				if kStr, ok := kVal.(string); ok {
					kind = strings.ToLower(kStr)
					foundKind = true
				}
			}
		}

		if !foundKind {
			// If we can't identify the kind, we treat it as "unknown".
			// Policy: If inclusion filter is active, drop unknowns?
			// Or keep them?
			// Usually strict filtering implies dropping things that don't match.
			if len(includeSet) > 0 {
				continue
			}
			// If only exclusion is active, and we don't know the kind, we can't exclude it, so keep it.
			filtered = append(filtered, doc)
			continue
		}

		// 1. Check Inclusion
		if len(includeSet) > 0 {
			if !includeSet[kind] {
				// Not in allowlist -> Drop
				continue
			}
		}

		// 2. Check Exclusion
		if len(excludeSet) > 0 {
			if excludeSet[kind] {
				// In denylist -> Drop
				continue
			}
		}

		// Passed both checks
		filtered = append(filtered, doc)
	}

	return filtered
}
