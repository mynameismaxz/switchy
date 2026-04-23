package shell

import (
	"fmt"
	"os"
	"path/filepath"
)

func rcPaths(t ShellType) (rcFile, bakFile string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	switch t {
	case ShellZsh:
		rc := filepath.Join(home, ".zshrc")
		return rc, rc + ".switchy.bak", nil
	case ShellBash:
		rc := filepath.Join(home, ".bashrc")
		return rc, rc + ".switchy.bak", nil
	default:
		return "", "", fmt.Errorf("unsupported shell type: %q", t)
	}
}

// BackupIfNeeded creates a .switchy.bak copy of the RC file if one doesn't already exist.
func BackupIfNeeded(sh DetectedShell) error {
	if _, err := os.Stat(sh.BakFile); err == nil {
		return nil // backup already exists
	}
	data, err := os.ReadFile(sh.RCFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // RC file doesn't exist yet; nothing to back up
		}
		return fmt.Errorf("cannot read %s: %w", sh.RCFile, err)
	}
	if err := os.WriteFile(sh.BakFile, data, 0600); err != nil {
		return fmt.Errorf("cannot create backup %s: %w", sh.BakFile, err)
	}
	return nil
}
