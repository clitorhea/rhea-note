package tui

import (
	"bytes"
	"encoding/json"
	"go/format"
	"regexp"
	"strings"
)

var (
	jsonBlockRegex = regexp.MustCompile("(?s)```json\n(.*?)\n```")
	goBlockRegex   = regexp.MustCompile("(?s)```go\n(.*?)\n```")
)

// autoFormatNote intelligently formats raw JSON documents or specific code blocks (like JSON and Go) embedded in Markdown.
func autoFormatNote(content string) string {
	// 1. If the ENTIRE note is just raw JSON, format it completely!
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(trimmed), "", "  "); err == nil {
			return buf.String()
		}
	}

	// 2. Format JSON blocks within markdown
	content = jsonBlockRegex.ReplaceAllStringFunc(content, func(m string) string {
		match := jsonBlockRegex.FindStringSubmatch(m)
		if len(match) > 1 {
			raw := match[1]
			var buf bytes.Buffer
			if err := json.Indent(&buf, []byte(raw), "", "  "); err == nil {
				return "```json\n" + buf.String() + "\n```"
			}
		}
		return m
	})

	// 3. Format Go blocks within markdown
	content = goBlockRegex.ReplaceAllStringFunc(content, func(m string) string {
		match := goBlockRegex.FindStringSubmatch(m)
		if len(match) > 1 {
			raw := match[1]
			formatted, err := format.Source([]byte(raw))
			if err == nil {
				// format.Source might add trailing newlines, so we trim it
				return "```go\n" + strings.TrimSpace(string(formatted)) + "\n```"
			}
		}
		return m
	})

	return content
}
