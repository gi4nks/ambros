package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gi4nks/ambros/v3/internal/errors" // New import
)

var validCmdName = regexp.MustCompile(`^[A-Za-z0-9._/\\-]+$`)

// ValidateCommandName ensures the command name does not contain obvious shell
// metacharacters. It allows alphanumerics, dot, underscore, dash and slashes.
func ValidateCommandName(name string) bool {
	if name == "" {
		return false
	}
	return validCmdName.MatchString(name)
}

// defaultPluginsBase returns the default plugins directory under the user's home.
// This is used when a relative path is provided for a plugin executable (e.g. "./run.sh").
func defaultPluginsBase() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// fallback to current dir if home not available
		return filepath.Join(".", ".ambros", "plugins")
	}
	return filepath.Join(home, ".ambros", "plugins")
}

// ensureNoSymlinkInPath returns an error if any path component between base and target
// is a symbolic link. If the target doesn't exist, it still checks components that do.
func ensureNoSymlinkInPath(base, target string) error {
	base = filepath.Clean(base)
	target = filepath.Clean(target)
	// ensure base is a prefix of target
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return errors.NewError(errors.ErrInternalServer, "failed to evaluate relative path", err)
	}
	if strings.HasPrefix(rel, "..") {
		return errors.NewError(errors.ErrInvalidCommand, "target not under base path", nil)
	}

	cur := base
	// if rel == "." then target == base; still check base
	if rel == "." {
		fi, err := os.Lstat(cur)
		if err == nil && (fi.Mode()&os.ModeSymlink) != 0 {
			return errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("symlink detected in path: %s", cur), nil)
		}
		return nil
	}

	parts := strings.Split(rel, string(filepath.Separator))
	for _, p := range parts {
		cur = filepath.Join(cur, p)
		fi, err := os.Lstat(cur)
		if err != nil {
			// if the component doesn't exist yet, skip (we only can check existing components)
			continue
		}
		if (fi.Mode() & os.ModeSymlink) != 0 {
			return errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("symlink detected in path: %s", cur), nil)
		}
	}
	return nil
}

// ResolveCommandPath validates the name and returns an absolute path to the
// executable. It rejects names with shell metacharacters and attempts to
// canonicalize path-style names to a safe plugins directory. For simple names
// it falls back to exec.LookPath.
func SafeResolveCommandPath(name string) (string, error) {
	if name == "" {
		return "", errors.NewError(errors.ErrInvalidCommand, "command name cannot be empty", nil)
	}
	// Reject names containing common shell metacharacters
	if strings.ContainsAny(name, ";&|$<>`*?~(){}[]'\"\\") {
		return "", errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("invalid characters in command name: %s", name), nil)
	}
	if !ValidateCommandName(name) {
		return "", errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("invalid command name: %s", name), nil)
	}

	// If the name contains a path separator, treat it as a path-style invocation.
	// For safety, only allow relative paths and resolve them under the default
	// plugins directory to avoid arbitrary filesystem execution.
	if strings.ContainsAny(name, string(filepath.Separator)) {
		if filepath.IsAbs(name) {
			return "", errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("absolute paths are not allowed: %s", name), nil)
		}
		// Reject traversal segments
		if strings.HasPrefix(name, "..") || strings.Contains(name, string(filepath.Separator)+"..") {
			return "", errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("path traversal not allowed in command name: %s", name), nil)
		}

		pluginsBase := defaultPluginsBase()
		absPath := filepath.Join(pluginsBase, filepath.Clean(name))

		// Ensure path is inside plugins base
		if rel, err := filepath.Rel(pluginsBase, absPath); err != nil || strings.HasPrefix(rel, "..") {
			return "", errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("resolved path outside plugins base: %s", absPath), nil)
		}

		// Check for symlinks in the path components
		if err := ensureNoSymlinkInPath(pluginsBase, absPath); err != nil {
			return "", err
		}

		// Verify the file exists and is not a symlink and is executable
		fi, err := os.Lstat(absPath)
		if err != nil {
			return "", err
		}
		if (fi.Mode() & os.ModeSymlink) != 0 {
			return "", errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("executable path is a symlink: %s", absPath), nil)
		}
		if fi.Mode().Perm()&0100 == 0 {
			return "", errors.NewError(errors.ErrInvalidCommand, fmt.Sprintf("executable does not have execute bit: %s", absPath), nil)
		}

		return absPath, nil
	}

	// Otherwise use PATH lookup
	path, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	return path, nil
}
