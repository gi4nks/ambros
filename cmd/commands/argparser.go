package commands

import (
	"strings"

	"github.com/gi4nks/ambros/v3/internal/errors"
)

// splitByUnescapedPipe splits a command string by unquoted/unescaped pipes.
// It respects single and double quotes and backslash escapes.
func splitByUnescapedPipe(s string) ([]string, error) {
	var parts []string
	var cur strings.Builder
	inSingle := false
	inDouble := false
	escape := false

	for _, r := range s {
		if escape {
			cur.WriteRune(r)
			escape = false
			continue
		}
		switch r {
		case '\\':
			escape = true
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				continue
			}
			cur.WriteRune(r)
		case '"':
			if !inSingle {
				inDouble = !inDouble
				continue
			}
			cur.WriteRune(r)
		case '|':
			if !inSingle && !inDouble {
				parts = append(parts, strings.TrimSpace(cur.String()))
				cur.Reset()
				continue
			}
			cur.WriteRune(r)
		default:
			cur.WriteRune(r)
		}
	}

	if escape {
		return nil, errors.NewError(errors.ErrInvalidCommand, "unfinished escape in command", nil)
	}
	if inSingle || inDouble {
		return nil, errors.NewError(errors.ErrInvalidCommand, "unmatched quote in command", nil)
	}

	last := strings.TrimSpace(cur.String())
	if last != "" {
		parts = append(parts, last)
	}
	return parts, nil
}

// shellFields splits a shell-like command into fields respecting quotes and escapes.
func shellFields(s string) ([]string, error) {
	var res []string
	var cur strings.Builder
	inSingle := false
	inDouble := false
	escape := false

	for _, r := range s {
		if escape {
			cur.WriteRune(r)
			escape = false
			continue
		}
		switch r {
		case '\\':
			escape = true
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				continue
			}
			cur.WriteRune(r)
		case '"':
			if !inSingle {
				inDouble = !inDouble
				continue
			}
			cur.WriteRune(r)
		case ' ', '\t':
			if !inSingle && !inDouble {
				if cur.Len() > 0 {
					res = append(res, cur.String())
					cur.Reset()
				}
				continue
			}
			cur.WriteRune(r)
		default:
			cur.WriteRune(r)
		}
	}

	if escape {
		return nil, errors.NewError(errors.ErrInvalidCommand, "unfinished escape in command", nil)
	}
	if inSingle || inDouble {
		return nil, errors.NewError(errors.ErrInvalidCommand, "unmatched quote in command", nil)
	}
	if cur.Len() > 0 {
		res = append(res, cur.String())
	}
	return res, nil
}
