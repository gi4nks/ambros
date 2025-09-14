//go:build windows
// +build windows

package commands

import (
	"os"
)

// notifyWinch is a no-op on Windows where SIGWINCH is not available.
func notifyWinch(sig chan os.Signal) {
	// no-op
	_ = sig
}
