package commands

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall" // Required for extractExitStatus

	"github.com/creack/pty"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
)

// Executor provides core functionality for executing commands,
// capturing output, and handling interactive sessions.
type Executor struct {
	logger *zap.Logger
}

// NewExecutor creates a new Executor instance.
func NewExecutor(logger *zap.Logger) *Executor {
	return &Executor{
		logger: logger,
	}
}

// IsTerminal checks if the current stdin is a TTY.
func (e *Executor) IsTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd())
}

// extractExitStatus extracts the exit status code from an error returned by exec.Command.Wait()
// On Unix-like platforms this checks for syscall.WaitStatus.
func (e *Executor) extractExitStatus(err error) (int, bool) {
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

// notifyWinch registers SIGWINCH into the provided channel on platforms that support it.
func (e *Executor) notifyWinch(sig chan os.Signal) {
	signal.Notify(sig, syscall.SIGWINCH)
}

// interactiveCmds is a small curated set of known interactive programs.
// This list is intentionally small — detection is conservative and only used
// to warn users to prefer --auto (attach a PTY) when appropriate.
var interactiveCmds = map[string]struct{}{
	"ssh":         {},
	"su":          {},
	"sudo":        {},
	"vim":         {},
	"vi":          {},
	"nano":        {},
	"less":        {},
	"more":        {},
	"man":         {},
	"top":         {},
	"htop":        {},
	"screen":      {},
	"tmux":        {},
	"passwd":      {},
	"tail":        {},
	"docker":      {},
	"kubectl":     {},
	"kubectl.exe": {},
	"docker.exe":  {},
}

// isLikelyInteractive inspects a command name and args to heuristically determine
// whether the command is likely to be interactive (requires a TTY/user input).
// This is conservative and only used to show a small warning to the user.
func (e *Executor) isLikelyInteractive(command string, args []string) bool {
	if command == "" {
		return false
	}

	// Use base name if full path is given
	cmd := strings.ToLower(filepath.Base(command))

	if _, ok := interactiveCmds[cmd]; ok {
		// For docker/kubectl we need to inspect args for -it / exec/run
		if cmd == "docker" || cmd == "docker.exe" {
			return e.dockerLooksInteractive(args)
		}
		if cmd == "kubectl" {
			return e.kubectlLooksInteractive(args)
		}
		// tail -f often waits for input/streams output; treat as interactive-like
		if cmd == "tail" {
			return e.argContains(args, "-f") || e.argContains(args, "-F")
		}
		// Generic interactive-like matches (ssh, vi, vim, less, more, top, ...) are true
		return true
	}

	// If command matches docker/kubectl via absolute path e.g. /usr/bin/docker
	if strings.Contains(cmd, "docker") {
		return e.dockerLooksInteractive(args)
	}
	if strings.Contains(cmd, "kubectl") {
		return e.kubectlLooksInteractive(args)
	}

	// Check for interactive flags commonly used in tools like `-it` or `-i` or `-t`
	// However, this should only apply if the command is *already* in the interactive set
	// or if the flag is part of a more specific interactive pattern (like docker/kubectl).
	// A generic `-i` or `-t` for a non-interactive command should not make it interactive.
	// This loop is now redundant as specific interactive commands are handled above.
	// The more robust solution is to define a command as interactive explicitly.
	return false
}

func (e *Executor) dockerLooksInteractive(args []string) bool {
	// docker run -it ... or docker exec -it ...
	isRunOrExec := false
	for _, a := range args {
		if a == "run" || a == "exec" {
			isRunOrExec = true
			continue
		}
		if a == "-it" || a == "-i" || a == "-t" || strings.Contains(a, "-it") {
			return true
		}
	}
	return isRunOrExec && e.argContains(args, "-it")
}

func (e *Executor) kubectlLooksInteractive(args []string) bool {
	// kubectl exec -it or kubectl run -it
	isExecOrRun := false
	for _, a := range args {
		if a == "exec" || a == "run" {
			isExecOrRun = true
			continue
		}
		if a == "-it" || a == "-i" || a == "-t" {
			return true
		}
	}
	return isExecOrRun && e.argContains(args, "-it")
}

func (e *Executor) argContains(args []string, token string) bool {
	for _, a := range args {
		if a == token {
			return true
		}
	}
	return false
}

// ResolveCommandPath resolves the full path of a command using safeexec.SafeResolveCommandPath
// to ensure command safety.
func (e *Executor) ResolveCommandPath(name string) (string, error) {
	return SafeResolveCommandPath(name)
}

// ExecuteCommand runs the command and captures combined stdout/stderr.
// Returns the combined output, error message (if any), success status, and an error if execution failed to start.
func (e *Executor) ExecuteCommand(name string, args []string) (string, string, bool, error) {
	if _, err := e.ResolveCommandPath(name); err != nil {
		// Do not treat a missing executable as a hard error here — keep compatibility
		// with existing tests which expect ExecuteCommand to return (success=false,
		// errorMsg!="", err==nil) for nonexistent commands.
		return "", err.Error(), false, nil
	}
	cmd := exec.Command(name, args...)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()

	success := err == nil
	var errorMsg string

	if err != nil {
		errorMsg = err.Error()
	}

	return string(output), errorMsg, success, nil
}

// ExecuteCommandAuto runs the target command transparently: it attaches the current
// process stdin/stdout/stderr to the child, forwards signals and returns the child's
// exit code. If the command fails to start, an error is returned and exit code will be 1.
func (e *Executor) ExecuteCommandAuto(name string, args []string) (int, error) {
	if _, err := e.ResolveCommandPath(name); err != nil {
		return 1, err
	}
	cmd := exec.Command(name, args...)

	// If stdin is a TTY, allocate a pty so interactive programs work correctly.
	if isatty.IsTerminal(os.Stdin.Fd()) {
		ptmx, err := pty.Start(cmd)
		if err != nil {
			return 1, err
		}
		// Make sure to close the pty when done
		defer func() { _ = ptmx.Close() }()

		// Handle window size changes
		sigwinch := make(chan os.Signal, 1)
		// platform-specific registration: may be a no-op on Windows
		e.notifyWinch(sigwinch)
		go func() {
			for range sigwinch {
				_ = pty.InheritSize(os.Stdin, ptmx)
			}
		}()
		// Ensure initial size is copied
		_ = pty.InheritSize(os.Stdin, ptmx)

		// Copy input and output and wait for them to finish to avoid races
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = io.Copy(ptmx, os.Stdin)
		}()
		go func() {
			defer wg.Done()
			_, _ = io.Copy(os.Stdout, ptmx)
		}()

		// Forward signals to child process
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs)
		done := make(chan struct{})
		go func() {
			for s := range sigs {
				if cmd.Process != nil {
					_ = cmd.Process.Signal(s)
				}
			}
			close(done)
		}()
		// Wait for command to exit
		err = cmd.Wait()

		// Close the PTY to ensure io.Copy goroutines return, then wait for them
		_ = ptmx.Close()
		wg.Wait()

		// Cleanup
		signal.Stop(sigwinch)
		close(sigwinch)
		signal.Stop(sigs)
		close(sigs)
		<-done

		if status, ok := e.extractExitStatus(err); ok {
			return status, nil
		}
		return 1, err
	}

	// Non-tty case: attach stdio directly
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return 1, err
	}

	// Forward signals to child
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs)
	done := make(chan struct{})

	go func() {
		for s := range sigs {
			if cmd.Process != nil {
				_ = cmd.Process.Signal(s)
			}
		}
		close(done)
	}()

	err := cmd.Wait()

	// Stop forwarding
	signal.Stop(sigs)
	close(sigs)
	<-done

	if status, ok := e.extractExitStatus(err); ok {
		return status, nil
	}
	return 1, err
}

// ExecuteCapture runs the command and always captures combined
// stdout/stderr into a byte buffer and returns (exitCode, output, error).
func (e *Executor) ExecuteCapture(name string, args []string) (int, string, error) {
	e.logger.Debug("ExecuteCapture invoked", zap.String("name", name), zap.Strings("args", args))

	if name == "" {
		return 1, "", errors.NewError(errors.ErrInvalidCommand,
			"No command specified. Use: ambros run [flags] -- <command> [args...]", nil)
	}

	// Resolve command path to avoid shell lookup surprises
	if _, err := e.ResolveCommandPath(name); err != nil {
		return 1, "", err
	}

	// If not a TTY, use CombinedOutput for simplicity
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		cmd := exec.Command(name, args...)
		out, err := cmd.CombinedOutput()
		if err == nil {
			return 0, string(out), nil
		}
		if status, ok := e.extractExitStatus(err); ok {
			return status, string(out), nil
		}
		return 1, string(out), err
	}

	// TTY case: allocate a pty and capture output by copying into a buffer
	cmd := exec.Command(name, args...)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return 1, "", err
	}
	defer func() { _ = ptmx.Close() }()

	// Copy output into buffer while also writing to stdout so user sees it
	var buf strings.Builder
	done := make(chan struct{})
	go func() {
		// Read from PTY and copy to both buffer and stdout
		b := make([]byte, 1024)
		for {
			n, rerr := ptmx.Read(b)
			if n > 0 {
				s := string(b[:n])
				buf.WriteString(s)
				_, _ = os.Stdout.Write(b[:n])
			}
			if rerr != nil {
				break
			}
		}
		close(done)
	}()

	err = cmd.Wait()
	<-done

	if err == nil {
		return 0, buf.String(), nil
	}
	if status, ok := e.extractExitStatus(err); ok {
		return status, buf.String(), nil
	}
	return 1, buf.String(), err
}
