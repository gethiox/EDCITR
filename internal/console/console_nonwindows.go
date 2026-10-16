//go:build !windows

package console

import "os/exec"

func init() {
	// Natively handled by Linux/macOS
}

func HideConsoleWindow(cmd *exec.Cmd) {
	// Linux/macOS GUI apps don't spawn child terminals, do nothing
}
