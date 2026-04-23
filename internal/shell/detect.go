package shell

import (
	"fmt"
	"os"
	"strings"
)

type ShellType string

const (
	ShellZsh  ShellType = "zsh"
	ShellBash ShellType = "bash"
)

type DetectedShell struct {
	Type    ShellType
	RCFile  string
	BakFile string
}

// Detect returns the shell to use. override may be "zsh", "bash", or "" for auto-detect.
func Detect(override string) (DetectedShell, error) {
	var shellStr string
	if override != "" {
		shellStr = override
	} else {
		shellStr = os.Getenv("SHELL")
		if shellStr == "" {
			return DetectedShell{}, fmt.Errorf("$SHELL is not set\nTry: swy use <profile> --persistent --shell zsh")
		}
	}

	switch {
	case strings.Contains(shellStr, "zsh"):
		return newDetectedShell(ShellZsh)
	case strings.Contains(shellStr, "bash"):
		return newDetectedShell(ShellBash)
	default:
		return DetectedShell{}, fmt.Errorf(
			"unsupported shell %q\nSupported shells: zsh, bash\nTry: swy use <profile> --persistent --shell zsh",
			shellStr,
		)
	}
}

func newDetectedShell(t ShellType) (DetectedShell, error) {
	rc, bak, err := rcPaths(t)
	if err != nil {
		return DetectedShell{}, err
	}
	return DetectedShell{Type: t, RCFile: rc, BakFile: bak}, nil
}
