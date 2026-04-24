package shell

import (
	"maps"
	"sort"
	"strings"
)

// UpsertExports edits the RC file directly: for each key in vars, it updates the first
// matching "export KEY=..." line found, or appends a new export if none exists.
// Commented-out lines (starting with #) are ignored.
func UpsertExports(sh DetectedShell, vars map[string]string) error {
	if err := BackupIfNeeded(sh); err != nil {
		return err
	}
	lines, err := readLines(sh.RCFile)
	if err != nil {
		return err
	}

	remaining := make(map[string]string, len(vars))
	maps.Copy(remaining, vars)

	for i, line := range lines {
		for k, v := range remaining {
			if isExportLine(line, k) {
				lines[i] = FormatExport(k, v)
				delete(remaining, k)
				break
			}
		}
	}

	if len(remaining) > 0 {
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
		keys := make([]string, 0, len(remaining))
		for k := range remaining {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			lines = append(lines, FormatExport(k, remaining[k]))
		}
	}

	return writeLines(sh.RCFile, lines)
}

// isExportLine reports whether line is an active (non-commented) export statement for key.
func isExportLine(line, key string) bool {
	trimmed := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(trimmed, "#") {
		return false
	}
	return strings.HasPrefix(trimmed, "export "+key+"=") ||
		strings.HasPrefix(trimmed, "export "+key+" =")
}
