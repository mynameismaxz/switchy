package clipboard

import (
	"os/exec"
	"runtime"
)

// Copy writes text to the system clipboard.
// Returns false (silently) if no clipboard tool is available.
func Copy(text string) bool {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if path, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command(path, "-selection", "clipboard")
		} else if path, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command(path, "--clipboard", "--input")
		} else if path, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command(path)
		}
	case "windows":
		cmd = exec.Command("clip")
	}
	if cmd == nil {
		return false
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		return false
	}
	if err := cmd.Start(); err != nil {
		return false
	}
	if _, err := in.Write([]byte(text)); err != nil {
		return false
	}
	in.Close()
	return cmd.Wait() == nil
}
