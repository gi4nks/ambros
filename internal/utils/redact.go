package utils

import (
	"regexp"
	"sync"
)

var (
	once     sync.Once
	patterns []*regexp.Regexp

	// Default patterns for common secrets
	defaultPatterns = map[string]string{
		"OpenAI":            `sk-[A-Za-z0-9]{32,}`,
		"GitHub":            `gh[p|o|u|s]_[A-Za-z0-9_]{30,}`,
		"AWS Access Key":    `AKIA[0-9A-Z]{16}`,
		"AWS Secret Key":    `[0-9a-zA-Z/+]{40}`,
		"Slack Webhook":     `https://hooks.slack.com/services/T[A-Z0-9]+/B[A-Z0-9]+/[A-Za-z0-9]+`,
		"Google Cloud API":  `AIza[0-9A-Za-z-_]{35}`,
		"Generic Password":  `(?i)(password|passwd|pwd)\s*[:=]\s*[^\s]+`,
		"Private Key":       `(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`,
		"Generic Auth":      `(?i)Bearer\s+[A-Za-z0-9\-._~+/]+=*`,
	}
)

func init() {
	for _, p := range defaultPatterns {
		patterns = append(patterns, regexp.MustCompile(p))
	}
}

// AddCustomPatterns adds user-defined regex patterns to the redactor
func AddCustomPatterns(customRegexes []string) {
	for _, r := range customRegexes {
		re, err := regexp.Compile(r)
		if err == nil {
			patterns = append(patterns, re)
		}
	}
}

// Redact replaces sensitive information in the input string with [REDACTED]
func Redact(input string) string {
	if input == "" {
		return ""
	}

	out := input
	for _, re := range patterns {
		out = re.ReplaceAllString(out, "[REDACTED]")
	}
	return out
}
