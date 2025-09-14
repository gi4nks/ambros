//go:build !windows
// +build !windows

package commands

import (
	"os/exec"
	"syscall"
)

// extractExitStatus extracts the exit status code from an error returned by exec.Command.Wait()
// On Unix-like platforms this checks for syscall.WaitStatus.
func extractExitStatus(err error) (int, bool) {
	if err == nil {
		return 0, true
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), true
		}
	}
	return 1, false
}
