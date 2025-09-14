//go:build windows
// +build windows

package commands

import (
	"os/exec"
)

// extractExitStatus on Windows attempts to derive the exit status but returns (1,false)
// if it cannot determine a platform-specific status. This stub allows the code to compile
// on Windows and cross-compile targets.
func extractExitStatus(err error) (int, bool) {
	if err == nil {
		return 0, true
	}
	if _, ok := err.(*exec.ExitError); ok {
		// On Windows the concrete implementation differs; return a generic non-zero code
		return 1, true
	}
	return 1, false
}
