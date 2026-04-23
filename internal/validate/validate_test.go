package validate

import (
	"testing"
)

func TestProfileName(t *testing.T) {
	valid := []string{"local", "staging", "prod", "my-env", "env_1", "A", "a1-b_2"}
	for _, name := range valid {
		if err := ProfileName(name); err != nil {
			t.Errorf("ProfileName(%q) should be valid, got: %v", name, err)
		}
	}

	invalid := []string{"", "my env", "env@prod", "env.prod", "env/prod"}
	for _, name := range invalid {
		if err := ProfileName(name); err == nil {
			t.Errorf("ProfileName(%q) should be invalid", name)
		}
	}
}

func TestEnvKey(t *testing.T) {
	valid := []string{"ANTHROPIC_BASE_URL", "API_KEY", "DEBUG", "_PRIVATE", "FOO123"}
	for _, key := range valid {
		if err := EnvKey(key); err != nil {
			t.Errorf("EnvKey(%q) should be valid, got: %v", key, err)
		}
	}

	invalid := []string{"", "lowercase", "123_FOO", "FOO-BAR", "FOO BAR", "foo"}
	for _, key := range invalid {
		if err := EnvKey(key); err == nil {
			t.Errorf("EnvKey(%q) should be invalid", key)
		}
	}
}

func TestEnvValue(t *testing.T) {
	if err := EnvValue("hello world"); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
	if err := EnvValue(""); err != nil {
		t.Errorf("empty value should be allowed, got: %v", err)
	}
	if err := EnvValue("line1\nline2"); err == nil {
		t.Error("multiline value should be rejected")
	}
}

func TestParseKVPairs(t *testing.T) {
	pairs, err := ParseKVPairs([]string{"FOO=bar", "BAR=hello world", "BAZ="})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pairs["FOO"] != "bar" {
		t.Errorf("expected bar, got %q", pairs["FOO"])
	}
	if pairs["BAR"] != "hello world" {
		t.Errorf("expected 'hello world', got %q", pairs["BAR"])
	}
	if pairs["BAZ"] != "" {
		t.Errorf("expected empty, got %q", pairs["BAZ"])
	}

	// VALUE containing '=' should work (split on first = only)
	pairs2, err := ParseKVPairs([]string{"URL=http://foo.com/a=b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pairs2["URL"] != "http://foo.com/a=b" {
		t.Errorf("expected 'http://foo.com/a=b', got %q", pairs2["URL"])
	}

	// Missing = separator
	if _, err := ParseKVPairs([]string{"NOKV"}); err == nil {
		t.Error("missing = should be rejected")
	}

	// Lowercase key
	if _, err := ParseKVPairs([]string{"lower=val"}); err == nil {
		t.Error("lowercase key should be rejected")
	}

	// Empty input
	if _, err := ParseKVPairs([]string{}); err == nil {
		t.Error("empty pairs should be rejected")
	}
}

func TestIsSensitiveKey(t *testing.T) {
	sensitive := []string{"API_KEY", "ACCESS_TOKEN", "SECRET_VALUE", "DB_PASSWORD", "PRIVATE_KEY"}
	for _, k := range sensitive {
		if !IsSensitiveKey(k) {
			t.Errorf("expected %q to be sensitive", k)
		}
	}
	notSensitive := []string{"ANTHROPIC_BASE_URL", "DEBUG", "PORT", "HOST"}
	for _, k := range notSensitive {
		if IsSensitiveKey(k) {
			t.Errorf("expected %q to NOT be sensitive", k)
		}
	}
}
