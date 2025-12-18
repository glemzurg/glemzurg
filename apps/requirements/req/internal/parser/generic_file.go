package parser

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

func init() {

	// Regex syntax used by golang: https://github.com/google/re2/wiki/Syntax

	// A title in the generic file structure.
	_genericFileTitleRegexp = regexp.MustCompile(`^(#+)[[:space:]]*([[:alnum:]].*)`)
}

// Regex match a h1, h2, h3 line, etc.
var _genericFileTitleRegexp *regexp.Regexp

// Separators in the files.
const (
	_UML_MARKER  = "◆" // U+25C6
	_DATA_MARKER = "◇" // U+25C7
)

// All files have similar parts.
// This parser is pretty simple because we want
// richer details in the domain-specific code.
// The order of the structure is always:
//
//  1. optional Markdown, then
//  2. optional UmlComment, then
//  3. optional Data
type File struct {
	Title      string // Extracted from the first non-whitespace line if it has the pattern of: "# title"
	Markdown   string // The beginning of the document.
	UmlComment string // Comment for uml display is started by U+25C6, a black diamond (◆).
	Data       string // Parseable data is started by U+25C7, a white diamond (◇).
}

func newFile(filename, title, markdown, umlComment, data string) (file File, err error) {

	// No title, then use the filename.
	if title == "" {
		title = filename
	}

	return File{title, markdown, umlComment, data}, nil
}

func parseFile(filename, contents string) (file File, err error) {

	// There can only be one of each marker, but they aren't required.
	umlMarkerCount := strings.Count(contents, _UML_MARKER)
	if umlMarkerCount > 1 {
		return File{}, errors.WithStack(errors.Errorf(`More than one uml marker (%s): %+v`, _UML_MARKER, umlMarkerCount))
	}
	dataMarkerCount := strings.Count(contents, _DATA_MARKER)
	if dataMarkerCount > 1 {
		return File{}, errors.WithStack(errors.Errorf(`More than one data marker (%s): %+v`, _DATA_MARKER, dataMarkerCount))
	}

	// The parts of the generic file.
	var title, markdown, umlComment, data string

	// Start with everything assumed to be in the markdown.
	markdown = contents

	// The data happens last, so cut that off from the back of the content first for simplicity.
	if dataMarkerCount > 0 {
		splitMarkdown := strings.Split(markdown, _DATA_MARKER)
		markdown = splitMarkdown[0]
		data = splitMarkdown[1]
	}

	// Next work forward and get the uml comment.
	if umlMarkerCount > 0 {
		splitMarkdown := strings.Split(markdown, _UML_MARKER)
		markdown = splitMarkdown[0]
		umlComment = splitMarkdown[1]
	}

	// Remove leading and trailing whitespace from each.
	markdown = strings.TrimSpace(markdown)
	umlComment = strings.TrimSpace(umlComment)
	data = strings.TrimSpace(data)

	// Get the title from the markdown if ther is one.
	title = extractMarkdownTitle(markdown)

	return newFile(filename, title, markdown, umlComment, data)
}

func extractMarkdownTitle(markdown string) (title string) {

	// Ensure there is no white space leading.
	markdown = strings.TrimSpace(markdown)

	// Get the first line.
	splitMarkdown := strings.Split(markdown, "\n")
	firstLine := splitMarkdown[0]

	// Trim any whitespace.
	firstLine = strings.TrimSpace(firstLine)

	// Get any title there.
	titleBytes := _genericFileTitleRegexp.Find([]byte(firstLine))

	// Any header is a title.
	// Ignore any number of first charcters that are "#".
	title = strings.TrimLeft(string(titleBytes), "#")

	// Trim any white space.
	title = strings.TrimSpace(title)

	return title
}

func generateFileContent(markdown, umlComment, data string) string {
	// Control whitespace.
	markdown = strings.TrimSpace(markdown)
	umlComment = strings.TrimSpace(umlComment)
	data = strings.TrimSpace(data)

	// Title already in the markdown.
	content := markdown
	if umlComment != "" {
		content += "\n\n◆\n\n" + umlComment
	}
	if data != "" {
		content += "\n\n◇\n\n" + data
	}

	// Make content tidy for easy test comparison.
	if content != "" {
		content = strings.TrimSpace(content)
	}

	return content
}
