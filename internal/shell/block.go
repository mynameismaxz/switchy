package shell

import (
	"fmt"
	"os"
	"strings"
)

const (
	ActivationStart = "# >>> switchy start >>>"
	ActivationEnd   = "# <<< switchy end <<<"
	InitStart       = "# >>> switchy init start >>>"
	InitEnd         = "# <<< switchy init end <<<"
)

// WriteActivationBlock writes the activation managed block into the RC file.
// exports is a slice of "export KEY='value'" lines.
func WriteActivationBlock(sh DetectedShell, exports []string) error {
	if err := BackupIfNeeded(sh); err != nil {
		return err
	}
	lines, err := readLines(sh.RCFile)
	if err != nil {
		return err
	}
	block := buildBlock(ActivationStart, ActivationEnd, exports)
	lines = replaceOrAppendBlock(lines, ActivationStart, ActivationEnd, block)
	return writeLines(sh.RCFile, lines)
}

// WriteInitBlock writes the swyuse() helper block into the RC file.
func WriteInitBlock(sh DetectedShell) error {
	if err := BackupIfNeeded(sh); err != nil {
		return err
	}
	lines, err := readLines(sh.RCFile)
	if err != nil {
		return err
	}
	helper := []string{
		`swyuse() {`,
		`  eval "$(swy export "$1")"`,
		`}`,
	}
	block := buildBlock(InitStart, InitEnd, helper)
	lines = replaceOrAppendBlock(lines, InitStart, InitEnd, block)
	return writeLines(sh.RCFile, lines)
}

// HasInitBlock reports whether the RC file already contains the init managed block.
func HasInitBlock(sh DetectedShell) (bool, error) {
	lines, err := readLines(sh.RCFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	start, _ := findBlock(lines, InitStart, InitEnd)
	return start >= 0, nil
}

// RemoveActivationBlock removes the activation block from the RC file.
func RemoveActivationBlock(sh DetectedShell) error {
	lines, err := readLines(sh.RCFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	start, end := findBlock(lines, ActivationStart, ActivationEnd)
	if start < 0 {
		return nil
	}
	lines = append(lines[:start], lines[end+1:]...)
	return writeLines(sh.RCFile, lines)
}

// buildBlock wraps content lines with start/end markers.
func buildBlock(start, end string, content []string) []string {
	result := make([]string, 0, len(content)+2)
	result = append(result, start)
	result = append(result, content...)
	result = append(result, end)
	return result
}

// findBlock returns the start and end line indices (inclusive) of the managed block.
// Returns (-1, -1) if not found.
func findBlock(lines []string, startMarker, endMarker string) (start, end int) {
	start = -1
	for i, line := range lines {
		if strings.TrimRight(line, "\r") == startMarker {
			start = i
		}
		if start >= 0 && strings.TrimRight(line, "\r") == endMarker {
			return start, i
		}
	}
	return -1, -1
}

// replaceOrAppendBlock replaces an existing managed block or appends a new one.
func replaceOrAppendBlock(lines []string, startMarker, endMarker string, block []string) []string {
	start, end := findBlock(lines, startMarker, endMarker)
	if start >= 0 {
		// Replace lines[start..end] with block
		result := make([]string, 0, len(lines)-( end-start+1)+len(block))
		result = append(result, lines[:start]...)
		result = append(result, block...)
		result = append(result, lines[end+1:]...)
		return result
	}
	// Append: add blank line separator if file is non-empty
	if len(lines) > 0 && lines[len(lines)-1] != "" {
		lines = append(lines, "")
	}
	return append(lines, block...)
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is derived from os.UserHomeDir() + hardcoded rc filename, never from user input
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	text := string(data)
	// Split preserving line endings stripped; we rejoin with \n on write.
	// Handle both \n and \r\n.
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	// strings.Split on "a\nb\n" gives ["a","b",""] — preserve trailing newline semantics.
	return lines, nil
}

func writeLines(path string, lines []string) error {
	content := strings.Join(lines, "\n")
	return atomicWriteFile(path, []byte(content), 0644)
}

// atomicWriteFile writes data to path via a temp file + rename.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".switchy.tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("cannot finalize write to %s: %w", path, err)
	}
	return nil
}
