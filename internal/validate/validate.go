package validate

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	profileNameRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	envKeyRe      = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
)

func ProfileName(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if !profileNameRe.MatchString(name) {
		return fmt.Errorf("invalid profile name %q: only letters, numbers, _ and - are allowed", name)
	}
	return nil
}

func EnvKey(key string) error {
	if key == "" {
		return fmt.Errorf("env key cannot be empty")
	}
	if !envKeyRe.MatchString(key) {
		return fmt.Errorf("invalid env key %q: must match ^[A-Z_][A-Z0-9_]*$ (uppercase only)", key)
	}
	return nil
}

func EnvValue(value string) error {
	if strings.Contains(value, "\n") {
		return fmt.Errorf("multiline values are not supported")
	}
	return nil
}

// ParseKVPairs parses a slice of "KEY=VALUE" strings into a map.
func ParseKVPairs(pairs []string) (map[string]string, error) {
	result := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		idx := strings.Index(pair, "=")
		if idx < 0 {
			return nil, fmt.Errorf("malformed argument %q: expected KEY=VALUE format", pair)
		}
		key := pair[:idx]
		value := pair[idx+1:]
		if err := EnvKey(key); err != nil {
			return nil, err
		}
		if err := EnvValue(value); err != nil {
			return nil, fmt.Errorf("invalid value for %s: %w", key, err)
		}
		result[key] = value
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("at least one KEY=VALUE argument is required")
	}
	return result, nil
}

var sensitiveSubstrings = []string{"KEY", "TOKEN", "SECRET", "PASSWORD"}

func IsSensitiveKey(key string) bool {
	upper := strings.ToUpper(key)
	for _, sub := range sensitiveSubstrings {
		if strings.Contains(upper, sub) {
			return true
		}
	}
	return false
}

func MaskValue(_ string) string {
	return "****"
}
