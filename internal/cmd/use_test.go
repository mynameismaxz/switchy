package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintSourceApplyHint_Copied(t *testing.T) {
	var got bytes.Buffer
	var copiedCmd string

	printSourceApplyHint(&got, "/tmp/test/.zshrc", func(cmd string) bool {
		copiedCmd = cmd
		return true
	})

	if copiedCmd != "source /tmp/test/.zshrc" {
		t.Fatalf("copied command = %q, want %q", copiedCmd, "source /tmp/test/.zshrc")
	}
	if !strings.Contains(got.String(), "Copied to clipboard") {
		t.Fatalf("output = %q, want clipboard confirmation", got.String())
	}
}

func TestPrintSourceApplyHint_Fallback(t *testing.T) {
	var got bytes.Buffer

	printSourceApplyHint(&got, "/tmp/test/.bashrc", func(string) bool { return false })

	if got.String() != "To apply now:  source /tmp/test/.bashrc\n" {
		t.Fatalf("output = %q", got.String())
	}
}
