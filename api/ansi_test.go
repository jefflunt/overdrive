package api

import (
	"html/template"
	"testing"
)

func TestAnsiToHtml(t *testing.T) {
	tests := []struct {
		input    string
		expected template.HTML
	}{
		{
			input:    "Normal text",
			expected: "Normal text",
		},
		{
			input:    "\x1b[31mRed text\x1b[0m",
			expected: "<span class=\"ansi-fg-31\">Red text</span>",
		},
		{
			input:    "\x1b[1;32mBold Green\x1b[0m",
			expected: "<span class=\"ansi-bold\"><span class=\"ansi-fg-32\">Bold Green</span></span>",
		},
		{
			input:    "\x1b[31mRed \x1b[32mGreen\x1b[0m",
			expected: "<span class=\"ansi-fg-31\">Red <span class=\"ansi-fg-32\">Green</span></span>",
		},
		{
			input:    "Mixed \x1b[Kignore \x1b[31mColor",
			expected: "Mixed ignore <span class=\"ansi-fg-31\">Color</span>",
		},
		{
			input:    "XSS <script>alert(1)</script>",
			expected: "XSS &lt;script&gt;alert(1)&lt;/script&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := AnsiToHtml(tt.input)
			if actual != tt.expected {
				t.Errorf("AnsiToHtml(%q) = %q; want %q", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestFormatPrompt(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Normal prompt",
			expected: "Normal prompt",
		},
		{
			input:    "/bdoc-engineer build a feature",
			expected: "build a feature",
		},
		{
			input:    "/bdoc-quick fix a bug",
			expected: "fix a bug",
		},
		{
			input:    "don't replace /bdoc-quick if not at start",
			expected: "don't replace /bdoc-quick if not at start",
		},
		{
			input:    "/bdoc-quick at start and /bdoc-quick later",
			expected: "at start and /bdoc-quick later",
		},
		{
			input:    "/bdoc-update documentation",
			expected: "docs update",
		},
		{
			input:    "/bdoc-revert abc123def",
			expected: "revert abc123def",
		},
		{
			input:    "/bdoc-idea something",
			expected: "chat something",
		},
		{
			input:    "<script>alert(1)</script>",
			expected: "<script>alert(1)</script>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := FormatPrompt(tt.input)
			if actual != tt.expected {
				t.Errorf("FormatPrompt(%q) = %q; want %q", tt.input, actual, tt.expected)
			}
		})
	}
}
