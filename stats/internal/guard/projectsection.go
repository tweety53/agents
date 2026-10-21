package guard

import (
	"bytes"
	"os"
	"strings"
)

// projectSection is the one copy of the heading-to-next-heading reader that
// used to live inline in three scripts and drift: its bash twin was deleted
// (KAN-887), and TestProjectSectionParity still runs that twin, as it stood
// at the change's base, against this one.

// projectSection is the bash project_section: the body of
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
