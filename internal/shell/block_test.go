package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeTestShell(t *testing.T, rcContent string) (DetectedShell, string) {
	t.Helper()
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	if rcContent != "" {
		if err := os.WriteFile(rc, []byte(rcContent), 0644); err != nil {
			t.Fatal(err)
		}
	}
	sh := DetectedShell{
		Type:    ShellZsh,
		RCFile:  rc,
		BakFile: rc + ".switchy.bak",
	}
	return sh, dir
}

func TestWriteActivationBlock_EmptyFile(t *testing.T) {
	sh, _ := makeTestShell(t, "")
	exports := []string{"export FOO='bar'", "export SWITCHY_PROFILE='local'"}
	if err := WriteActivationBlock(sh, exports); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	content := string(data)
	if !strings.Contains(content, ActivationStart) {
		t.Error("missing start marker")
	}
	if !strings.Contains(content, "export FOO='bar'") {
		t.Error("missing export line")
	}
	if !strings.Contains(content, ActivationEnd) {
		t.Error("missing end marker")
	}
}

func TestWriteActivationBlock_ReplacesExistingBlock(t *testing.T) {
	initial := "# user content\n" +
		ActivationStart + "\n" +
		"export OLD='value'\n" +
		ActivationEnd + "\n" +
		"# more user content\n"
	sh, _ := makeTestShell(t, initial)

	exports := []string{"export NEW='value'"}
	if err := WriteActivationBlock(sh, exports); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	content := string(data)

	if strings.Contains(content, "export OLD='value'") {
		t.Error("old export should have been replaced")
	}
	if !strings.Contains(content, "export NEW='value'") {
		t.Error("new export should be present")
	}
	if !strings.Contains(content, "# user content") {
		t.Error("user content before block must be preserved")
	}
	if !strings.Contains(content, "# more user content") {
		t.Error("user content after block must be preserved")
	}
}

func TestWriteActivationBlock_PreservesOutsideContent(t *testing.T) {
	before := "export PATH=/usr/local/bin:$PATH\nalias ll='ls -la'\n"
	after := "# my custom stuff\nexport EDITOR=vim\n"
	initial := before + ActivationStart + "\nexport OLD='x'\n" + ActivationEnd + "\n" + after
	sh, _ := makeTestShell(t, initial)

	if err := WriteActivationBlock(sh, []string{"export NEW='y'"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	content := string(data)

	for _, line := range strings.Split(before, "\n") {
		if line != "" && !strings.Contains(content, line) {
			t.Errorf("line %q before block was lost", line)
		}
	}
	for _, line := range strings.Split(after, "\n") {
		if line != "" && !strings.Contains(content, line) {
			t.Errorf("line %q after block was lost", line)
		}
	}
}

func TestWriteActivationBlock_Idempotent(t *testing.T) {
	sh, _ := makeTestShell(t, "")
	exports := []string{"export FOO='bar'"}
	if err := WriteActivationBlock(sh, exports); err != nil {
		t.Fatal(err)
	}
	if err := WriteActivationBlock(sh, exports); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	// Ensure the block doesn't appear twice
	count := strings.Count(string(data), ActivationStart)
	if count != 1 {
		t.Errorf("block should appear exactly once, found %d times", count)
	}
}

func TestHasInitBlock_FalseWhenAbsent(t *testing.T) {
	sh, _ := makeTestShell(t, "# no init here\n")
	has, err := HasInitBlock(sh)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("expected no init block")
	}
}

func TestHasInitBlock_TrueWhenPresent(t *testing.T) {
	content := InitStart + "\nswyuse() { echo hi; }\n" + InitEnd + "\n"
	sh, _ := makeTestShell(t, content)
	has, err := HasInitBlock(sh)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("expected init block to be found")
	}
}

func TestBackupIfNeeded_CreatesBackup(t *testing.T) {
	sh, _ := makeTestShell(t, "export ORIGINAL=1\n")
	if err := BackupIfNeeded(sh); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(sh.BakFile); os.IsNotExist(err) {
		t.Error("backup file should have been created")
	}
	// Second call should not overwrite
	if err := os.WriteFile(sh.RCFile, []byte("modified"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := BackupIfNeeded(sh); err != nil {
		t.Fatal(err)
	}
	bak, _ := os.ReadFile(sh.BakFile)
	if string(bak) != "export ORIGINAL=1\n" {
		t.Error("backup should not have been overwritten on second call")
	}
}
