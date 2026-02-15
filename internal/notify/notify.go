package notify

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Tmux sends a message to the tmux status line.
func Tmux(msg string) {
	_ = exec.Command("tmux", "display-message", msg).Run()
}

// Desktop sends a native desktop notification.
func Desktop(title, msg string) {
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification %q with title %q`, msg, title)
		_ = exec.Command("osascript", "-e", script).Run()
	case "linux":
		_ = exec.Command("notify-send", title, msg).Run()
	}
}
