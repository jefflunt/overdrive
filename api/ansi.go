package api

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	mdhtml "github.com/yuin/goldmark/renderer/html"
)

var (
	ansiRegex = regexp.MustCompile("\x1b\\[([0-9;]*)([mK])")
)

func MarkdownToHtml(text string) template.HTML {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
		goldmark.WithRendererOptions(
			mdhtml.WithHardWraps(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(text), &buf); err != nil {
		return template.HTML(html.EscapeString(text))
	}
	return template.HTML(buf.String())
}

func AnsiToHtml(text string) template.HTML {
	escaped := html.EscapeString(text)

	var result strings.Builder
	openSpans := 0

	lastEnd := 0
	matches := ansiRegex.FindAllStringSubmatchIndex(escaped, -1)

	for _, m := range matches {
		// Append text before the match
		result.WriteString(escaped[lastEnd:m[0]])

		codes := escaped[m[2]:m[3]]
		terminator := escaped[m[4]:m[5]]

		if terminator == "m" {
			if codes == "" || codes == "0" {
				// Reset
				for openSpans > 0 {
					result.WriteString("</span>")
					openSpans--
				}
			} else {
				parts := strings.Split(codes, ";")
				for _, code := range parts {
					switch code {
					case "0":
						for openSpans > 0 {
							result.WriteString("</span>")
							openSpans--
						}
					case "1":
						result.WriteString("<span class=\"ansi-bold\">")
						openSpans++
					case "30", "31", "32", "33", "34", "35", "36", "37":
						result.WriteString(fmt.Sprintf("<span class=\"ansi-fg-%s\">", code))
						openSpans++
					case "90", "91", "92", "93", "94", "95", "96", "97":
						result.WriteString(fmt.Sprintf("<span class=\"ansi-fg-bright-%s\">", code))
						openSpans++
					}
				}
			}
		}
		// "K" (Erase in Line) is ignored as per requirements

		lastEnd = m[1]
	}

	// Append remaining text
	result.WriteString(escaped[lastEnd:])

	// Close any remaining spans
	for openSpans > 0 {
		result.WriteString("</span>")
		openSpans--
	}

	return template.HTML(result.String())
}

func FormatPrompt(p string) string {
	if strings.HasPrefix(p, "/bdoc-update") {
		return "docs update"
	}
	if strings.HasPrefix(p, "/bdoc-revert") {
		return "revert " + strings.TrimSpace(strings.TrimPrefix(p, "/bdoc-revert"))
	}
	if strings.HasPrefix(p, "/bdoc-idea") {
		p = strings.TrimPrefix(p, "/bdoc-idea")
		p = strings.TrimSpace(p)
		return "chat " + p
	}

	prefixes := []string{
		"/bdoc-engineer",
		"/bdoc-quick",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(p, prefix) {
			p = strings.TrimPrefix(p, prefix)
			p = strings.TrimSpace(p)
			break
		}
	}

	return p
}
