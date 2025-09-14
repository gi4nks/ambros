//go:build !windows
// +build !windows

package commands

import (
	"os"
	"os/signal"
	"syscall"
)

// notifyWinch registers SIGWINCH into the provided channel on platforms that support it.
func notifyWinch(sig chan os.Signal) {
	signal.Notify(sig, syscall.SIGWINCH)
}
