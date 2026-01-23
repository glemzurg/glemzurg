package errors

import (
	"embed"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
)

// ErrorDocs contains all embedded error documentation markdown files.
//
//go:embed *.md
var ErrorDocs embed.FS

// LoadErrorDoc loads the error documentation markdown file for the given error code.
// Error files are named with the pattern: {code}_{description}.md (e.g., 1001_model_name_required.md)
// Returns the file content and filename, or an error if no file exists for the given code.
func LoadErrorDoc(code int) (content string, filename string, err error) {
	prefix := strconv.Itoa(code) + "_"

	entries, err := fs.ReadDir(ErrorDocs, ".")
	if err != nil {
		return "", "", fmt.Errorf("failed to read error docs directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".md") {
			data, err := ErrorDocs.ReadFile(name)
			if err != nil {
				return "", "", fmt.Errorf("failed to read error doc %s: %w", name, err)
			}
			return string(data), name, nil
		}
	}

	return "", "", fmt.Errorf("no error documentation found for error code %d", code)
}
