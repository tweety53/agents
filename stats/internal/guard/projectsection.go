package guard

import (
	"bytes"
	"os"
	"strings"
)

// projectSection is the Go twin of scripts/lib/project-section.sh, which stays
// the source of truth while project-get.sh still sources it; its header
// carries the reasoning, and TestProjectSectionParity fails when the two
// print or return differently for the same file and key.

// projectSection is lib/project-section.sh's project_section: the body of
// "## <key>" up to the next "## " heading, a leading UTF-8 BOM dropped,
// blank lines trimmed at both ends, trailing newlines stripped.
func projectSection(file, key string) string {
	content, _ := os.ReadFile(file)
	content = bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))
	var body []string
	grab := false
	for _, line := range lines(content) {
		if line == "## "+key {
			grab = true
			continue
		}
		if strings.HasPrefix(line, "## ") && grab {
			break
		}
		if grab {
			body = append(body, line)
		}
	}
	for len(body) > 0 && strings.Trim(body[0], gdcSpace) == "" {
		body = body[1:]
	}
	for len(body) > 0 && strings.Trim(body[len(body)-1], gdcSpace) == "" {
		body = body[:len(body)-1]
	}
	return strings.TrimRight(strings.Join(body, "\n"), "\n")
}
