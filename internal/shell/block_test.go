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

func TestWriteInitBlock_ZshHook(t *testing.T) {
	sh, _ := makeTestShell(t, "")
	if err := WriteInitBlock(sh); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	content := string(data)
	want := []string{
		"swy() {",
		`"")  SWITCHY_HOOK=1 command swy "$@" && eval "$(command swy hook-apply)" ;;`,
		`use) shift; SWITCHY_HOOK=1 command swy use "$@" && eval "$(command swy hook-apply)" ;;`,
		`off) shift; eval "$(command swy off "$@")" ;;`,
		"add-zsh-hook precmd swy_hook",
	}
	for _, w := range want {
		if !strings.Contains(content, w) {
			t.Errorf("zsh hook missing %q\ncontent:\n%s", w, content)
		}
	}
	if strings.Contains(content, "PROMPT_COMMAND") {
		t.Error("zsh hook must not reference PROMPT_COMMAND")
	}
}

func TestWriteInitBlock_BashHook(t *testing.T) {
	sh, _ := makeTestShell(t, "")
	sh.Type = ShellBash
	sh.RCFile = filepath.Join(filepath.Dir(sh.RCFile), ".bashrc")
	if err := WriteInitBlock(sh); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	content := string(data)
	want := []string{
		"swy() {",
		`"")  SWITCHY_HOOK=1 command swy "$@" && eval "$(command swy hook-apply)" ;;`,
		`use) shift; SWITCHY_HOOK=1 command swy use "$@" && eval "$(command swy hook-apply)" ;;`,
		`PROMPT_COMMAND="swy_hook${PROMPT_COMMAND:+;$PROMPT_COMMAND}"`,
	}
	for _, w := range want {
		if !strings.Contains(content, w) {
			t.Errorf("bash hook missing %q\ncontent:\n%s", w, content)
		}
	}
	if strings.Contains(content, "add-zsh-hook") {
		t.Error("bash hook must not reference add-zsh-hook")
	}
}

func TestWriteInitBlock_UpgradeOldBlock(t *testing.T) {
	oldBlock := InitStart + "\nswyuse() {\n  eval \"$(swy export \"$1\")\"\n}\n" + InitEnd + "\n"
	sh, _ := makeTestShell(t, "# user content\n"+oldBlock+"# keep me\n")
	if err := WriteInitBlock(sh); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(sh.RCFile)
	content := string(data)
	if strings.Count(content, InitStart) != 1 {
		t.Error("old init block must be replaced, not duplicated")
	}
	if strings.Contains(content, `eval "$(swy export "$1")"`) {
		t.Error("old swyuse body should have been replaced")
	}
	if !strings.Contains(content, "hook-apply") {
		t.Error("new hook content should be present")
	}
	if !strings.Contains(content, "# user content") || !strings.Contains(content, "# keep me") {
		t.Error("content outside the block must be preserved")
	}
}

func TestInitBlockUpToDate(t *testing.T) {
	t.Run("old block", func(t *testing.T) {
		oldBlock := InitStart + "\nswyuse() { echo hi; }\n" + InitEnd + "\n"
		sh, _ := makeTestShell(t, oldBlock)
		upToDate, err := InitBlockUpToDate(sh)
		if err != nil {
			t.Fatal(err)
		}
		if upToDate {
			t.Error("old block should not be up to date")
		}
	})
	t.Run("hook without TUI support", func(t *testing.T) {
		oldHook := strings.Replace(HookScript(ShellZsh),
			`    "")  SWITCHY_HOOK=1 command swy "$@" && eval "$(command swy hook-apply)" ;;`+"\n", "", 1)
		sh, _ := makeTestShell(t, InitStart+"\n"+oldHook+"\n"+InitEnd+"\n")
		upToDate, err := InitBlockUpToDate(sh)
		if err != nil {
			t.Fatal(err)
		}
		if upToDate {
			t.Error("hook without TUI support should be upgraded")
		}
	})
	t.Run("current block", func(t *testing.T) {
		sh, _ := makeTestShell(t, "")
		if err := WriteInitBlock(sh); err != nil {
			t.Fatal(err)
		}
		upToDate, err := InitBlockUpToDate(sh)
		if err != nil {
			t.Fatal(err)
		}
		if !upToDate {
			t.Error("freshly written block should be up to date")
		}
	})
	t.Run("no block", func(t *testing.T) {
		sh, _ := makeTestShell(t, "")
		upToDate, err := InitBlockUpToDate(sh)
		if err != nil {
			t.Fatal(err)
		}
		if upToDate {
			t.Error("no block should not be up to date")
		}
	})
}

func TestHasActivationBlock(t *testing.T) {
	sh, _ := makeTestShell(t, "")
	has, err := HasActivationBlock(sh)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("expected no activation block")
	}
	if err := WriteActivationBlock(sh, []string{"export FOO='bar'"}); err != nil {
		t.Fatal(err)
	}
	has, err = HasActivationBlock(sh)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("expected activation block to be found")
	}
}

func TestHookScript_Printable(t *testing.T) {
	for _, tt := range []ShellType{ShellZsh, ShellBash} {
		script := HookScript(tt)
		if strings.Contains(script, "\n\n") {
			t.Errorf("%s hook script should not contain blank lines", tt)
		}
		if !strings.Contains(script, "hook-apply") {
			t.Errorf("%s hook script missing hook-apply", tt)
		}
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
