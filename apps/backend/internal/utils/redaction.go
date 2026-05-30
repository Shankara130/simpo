package utils

import (
	"strings"
)

// Redactor handles sensitive data redaction from logs and other outputs
type Redactor struct {
	patterns []string
}

// NewRedactor creates a new Redactor with default patterns
func NewRedactor() *Redactor {
	return &Redactor{
		patterns: []string{
			"password",
			"passwd",
			"token",
			"api_key",
			"apikey",
			"secret",
			"authorization",
			"cookie",
			"session",
		},
	}
}

// AddPattern adds a custom redaction pattern
func (r *Redactor) AddPattern(pattern string) {
	r.patterns = append(r.patterns, strings.ToLower(pattern))
}

// shouldRedactKey checks if a key should be redacted based on patterns
func (r *Redactor) shouldRedactKey(key string) bool {
	lowerKey := strings.ToLower(key)
	for _, pattern := range r.patterns {
		if strings.Contains(lowerKey, pattern) {
			return true
		}
	}
	return false
}

// Redact recursively redacts sensitive data in a map
func (r *Redactor) Redact(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		if r.shouldRedactKey(key) {
			result[key] = "****"
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			result[key] = r.Redact(nestedMap)
		} else {
			result[key] = value
		}
	}
	return result
}

// RedactString redacts sensitive data from strings
func (r *Redactor) RedactString(s string) string {
	lower := strings.ToLower(s)

	// Special case for Bearer authorization
	if strings.Contains(lower, "authorization: bearer") {
		return "Authorization: Bearer ****"
	}

	// Check for common patterns and redact values
	for _, pattern := range r.patterns {
		if strings.Contains(lower, pattern) {
			// Find the value after the pattern (e.g., "password=secret" -> "password=****")
			parts := strings.Split(s, "=")
			if len(parts) > 1 {
				// Redact everything after the first '='
				return parts[0] + "=****"
			}

			// For JSON-like strings, try to redact values
			if strings.Contains(s, "{") || strings.Contains(s, ":") {
				return r.redactJSONLike(s)
			}
		}
	}

	return s
}

// redactJSONLike attempts to redact values in JSON-like strings
func (r *Redactor) redactJSONLike(s string) string {
	result := s
	// Simple pattern: replace values after known keys with ****
	for _, pattern := range r.patterns {
		// Look for "key":"value" or "key":'value' patterns
		search := `"` + pattern + `"`
		if idx := strings.Index(strings.ToLower(result), search); idx >= 0 {
			// Found the key, now find the value
			// Look for the next : and either " or '
			valueStart := strings.Index(result[idx:], ":")
			if valueStart >= 0 {
				valueStart += idx
				// Find the quote after the colon
				for i := valueStart + 1; i < len(result); i++ {
					if result[i] == '"' || result[i] == '\'' {
						// Found the start of the value, now find the end
						quoteChar := result[i]
						for j := i + 1; j < len(result); j++ {
							if result[j] == quoteChar {
								// Replace the value with ****
								result = result[:i+1] + "****" + result[j:]
								break
							}
						}
						break
					}
				}
			}
		}
	}
	return result
}

// RedactHeader redacts sensitive data from HTTP headers
func (r *Redactor) RedactHeader(name, value string) string {
	if r.shouldRedactKey(name) {
		if strings.ToLower(name) == "authorization" && strings.HasPrefix(value, "Bearer ") {
			return "Bearer ****"
		}
		return "****"
	}
	return value
}

// RedactQueryString redacts sensitive query parameters
func (r *Redactor) RedactQueryString(query string) string {
	if query == "" {
		return ""
	}

	params := strings.Split(query, "&")
	var redacted []string

	for _, param := range params {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			if r.shouldRedactKey(key) {
				redacted = append(redacted, key+"=****")
			} else {
				redacted = append(redacted, param)
			}
		} else {
			redacted = append(redacted, param)
		}
	}

	return strings.Join(redacted, "&")
}

// RedactMap recursively redacts sensitive data from any map structure
func (r *Redactor) RedactMap(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		return r.Redact(v)
	case []interface{}:
		for i, item := range v {
			v[i] = r.RedactMap(item)
		}
		return v
	default:
		return data
	}
}
