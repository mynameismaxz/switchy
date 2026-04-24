package shell

import (
	"os"
	"strings"
	"testing"
)

func TestUpsertExports_EmptyFile(t *testing.T) {
	sh, _ := makeTestShell(t, "")
	vars := map[string]string{"FOO": "bar", "BAZ": "qux"}
	if err := UpsertExports(sh, vars); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	content := string(data)
	if !strings.Contains(content, "export FOO='bar'") {
		t.Error("missing FOO export")
	}
	if !strings.Contains(content, "export BAZ='qux'") {
		t.Error("missing BAZ export")
	}
}

func TestUpsertExports_UpdatesExistingLine(t *testing.T) {
	initial := "export FOO='old'\nexport OTHER='keep'\n"
	sh, _ := makeTestShell(t, initial)
	if err := UpsertExports(sh, map[string]string{"FOO": "new"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	content := string(data)
	if strings.Contains(content, "export FOO='old'") {
		t.Error("old value should have been replaced")
	}
	if !strings.Contains(content, "export FOO='new'") {
		t.Error("new value should be present")
	}
	if !strings.Contains(content, "export OTHER='keep'") {
		t.Error("unrelated export must be preserved")
	}
}

func TestUpsertExports_AppendsNewVar(t *testing.T) {
	initial := "export EXISTING='yes'\n"
	sh, _ := makeTestShell(t, initial)
	if err := UpsertExports(sh, map[string]string{"NEW": "value"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	content := string(data)
	if !strings.Contains(content, "export EXISTING='yes'") {
		t.Error("existing line must be preserved")
	}
	if !strings.Contains(content, "export NEW='value'") {
		t.Error("new export should have been appended")
	}
}

func TestUpsertExports_SkipsCommentedLine(t *testing.T) {
	initial := "# export FOO='commented'\n"
	sh, _ := makeTestShell(t, initial)
	if err := UpsertExports(sh, map[string]string{"FOO": "real"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	content := string(data)
	if !strings.Contains(content, "# export FOO='commented'") {
		t.Error("commented line must not be touched")
	}
	if !strings.Contains(content, "export FOO='real'") {
		t.Error("real export should have been appended")
	}
}

func TestUpsertExports_PreservesUserContent(t *testing.T) {
	initial := "alias ll='ls -la'\nexport PATH=/usr/local/bin:$PATH\nexport FOO='old'\n# comment\n"
	sh, _ := makeTestShell(t, initial)
	if err := UpsertExports(sh, map[string]string{"FOO": "new", "BAR": "added"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	content := string(data)
	if !strings.Contains(content, "alias ll='ls -la'") {
		t.Error("alias line must be preserved")
	}
	if !strings.Contains(content, "export PATH=/usr/local/bin:$PATH") {
		t.Error("PATH export must be preserved")
	}
	if !strings.Contains(content, "# comment") {
		t.Error("comment must be preserved")
	}
	if !strings.Contains(content, "export FOO='new'") {
		t.Error("FOO should be updated")
	}
	if !strings.Contains(content, "export BAR='added'") {
		t.Error("BAR should be appended")
	}
}

func TestUpsertExports_Idempotent(t *testing.T) {
	sh, _ := makeTestShell(t, "")
	vars := map[string]string{"FOO": "bar"}
	if err := UpsertExports(sh, vars); err != nil {
		t.Fatal(err)
	}
	if err := UpsertExports(sh, vars); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	count := strings.Count(string(data), "export FOO='bar'")
	if count != 1 {
		t.Errorf("export should appear exactly once, found %d", count)
	}
}

func TestIsExportLine(t *testing.T) {
	tests := []struct {
		line string
		key  string
		want bool
	}{
		{"export FOO='bar'", "FOO", true},
		{"export FOO=bar", "FOO", true},
		{"export FOO =bar", "FOO", true},
		{"# export FOO='bar'", "FOO", false},
		{"  # export FOO='bar'", "FOO", false},
		{"export FOOBAR='baz'", "FOO", false},
		{"export BAR='x'", "FOO", false},
		{"  export FOO='indented'", "FOO", true},
	}
	for _, tc := range tests {
		got := isExportLine(tc.line, tc.key)
		if got != tc.want {
			t.Errorf("isExportLine(%q, %q) = %v, want %v", tc.line, tc.key, got, tc.want)
		}
	}
}
