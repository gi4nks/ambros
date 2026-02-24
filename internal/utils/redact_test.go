package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedact(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "OpenAI Key",
			input:    "my key is sk-ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
			expected: "my key is [REDACTED]",
		},
		{
			name:     "GitHub Token",
			input:    "token: ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ01234567",
			expected: "token: [REDACTED]",
		},
		{
			name:     "AWS Access Key",
			input:    "AKIAABCDEFGHIJKLMNOP",
			expected: "[REDACTED]",
		},
		{
			name:     "Slack Webhook",
			input:    "https://hooks.slack.com/services/T00000/B00000/XXXXX",
			expected: "[REDACTED]",
		},
		{
			name:     "Password assignment",
			input:    "password: mypassword123",
			expected: "[REDACTED]",
		},
		{
			name:     "No secrets",
			input:    "ls -la /tmp",
			expected: "ls -la /tmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Redact(tt.input))
		})
	}
}

func TestCustomPatterns(t *testing.T) {
	AddCustomPatterns([]string{`MY-SECRET-[0-9]+`})
	
	input := "here is MY-SECRET-12345"
	expected := "here is [REDACTED]"
	
	assert.Equal(t, expected, Redact(input))
}
