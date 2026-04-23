package shell

import (
	"fmt"
	"strings"
)

// SingleQuoteEscape escapes a value for use inside single quotes in bash/zsh.
// A literal ' becomes '\'' (end quote, escaped quote, start quote).
func SingleQuoteEscape(value string) string {
	return strings.ReplaceAll(value, "'", `'\''`)
}

// FormatExport returns a shell export statement with the value safely single-quoted.
func FormatExport(key, value string) string {
	return fmt.Sprintf("export %s='%s'", key, SingleQuoteEscape(value))
}
