package requirements

import (
	"os"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

type fileParsed struct {
	filename string // The filename.
	header   string // The part of the file before any elements of a requirement.
	refs     []uint // The requirements tha make up this file, in the order they're in.

	// What about regenerating file.
	originalContents string // The original content of the file.
}

func parseFile(lastRef uint, filename string) (newLastRef uint, file fileParsed, fileReqs map[uint]Requirement, err error) {

	contents, err := os.ReadFile(filename)
	if err != nil {
		return 0, fileParsed{}, nil, errors.WithStack(err)
	}

	lastRef, file, fileReqs, err = parseFileContents(lastRef, filename, string(contents))
	if err != nil {
		return 0, fileParsed{}, nil, err
	}

	return lastRef, file, fileReqs, nil
}

func parseFileContents(lastRef uint, filename, contents string) (newLastRef uint, file fileParsed, fileReqs map[uint]Requirement, err error) {

	// We want to find all the reference lines at the bottom of the file
	// and remove them before parsing the rest. Otherwise the last requirement
	// will have them as part of its text, which is not right.
	refLinks, err := findRefLinks(contents)
	if err != nil {
		return 0, fileParsed{}, nil, err
	}
	contentsNoReferenceLinks := contents
	for _, refLink := range refLinks {
		contentsNoReferenceLinks = strings.ReplaceAll(contentsNoReferenceLinks, refLink.Match, "")
	}

	// Big sections of this document.
	header, reqTexts := splitFileOnReqs(contentsNoReferenceLinks)

	var refs []uint
	for _, reqText := range reqTexts {
		lastRef++
		req, err := newRequirement(lastRef, filename, reqText)
		if err != nil {
			return 0, fileParsed{}, nil, err
		}
		refs = append(refs, lastRef)
		if len(fileReqs) == 0 {
			fileReqs = map[uint]Requirement{}
		}
		fileReqs[lastRef] = req
	}

	file, err = newFileParsed(filename, contents, header, refs)
	if err != nil {
		return 0, fileParsed{}, nil, err
	}

	return lastRef, file, fileReqs, nil
}

func newFileParsed(filename, originalContents, header string, refs []uint) (file fileParsed, err error) {

	err = validation.Validate(filename,
		validation.Required, // not empty
	)
	if err != nil {
		return fileParsed{}, errors.WithStack(err)
	}

	return fileParsed{
		filename: filename,
		header:   header,
		refs:     refs,
		// Keep track of the original file body.
		originalContents: originalContents,
	}, nil
}

//===========================================
// Methods
//===========================================

func (f *fileParsed) Generate(config Config, reqs map[uint]Requirement) (generated string, err error) {

	// Start with header.
	generated = f.header

	// Add each requirement.
	for _, ref := range f.refs {
		req := reqs[ref]

		value, err := req.String(config.Aspects)
		if err != nil {
			return "", err
		}

		generated = generated + "\n\n" + value
	}

	return strings.TrimSpace(generated), nil
}

//===========================================
// Parse
//===========================================

func splitFileOnReqs(contents string) (header string, reqTexts []string) {

	lines := strings.Split(contents, "\n")

	// Examine lines one at a time.
	hasRequirement := false
	reqText := ""

	for _, line := range lines {

		// Test this line. Requirement?
		if isRequirementHeader(line) {
			hasRequirement = true

			// Save any requirement we just gathered.
			if reqText != "" {
				reqTexts = append(reqTexts, strings.TrimSpace(reqText))
			}
			reqText = "" // Reset it.
		}

		// Are we before requirements?
		if !hasRequirement {

			// Add this line to the header.
			header = header + "\n" + line

		} else {

			reqText = reqText + "\n" + line

		}
	}

	// Do we need to remember the last req?
	if hasRequirement {
		reqTexts = append(reqTexts, strings.TrimSpace(reqText))
	}

	header = strings.TrimSpace(header)

	return header, reqTexts
}
