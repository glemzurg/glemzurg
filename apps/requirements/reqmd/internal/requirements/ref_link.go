package requirements

import (
	"strings"

	"github.com/pkg/errors"
)

type RefLink struct {
	ReqId string // Value like "F13"
	Match string // The regex matched text for this link like "[f13]: url 'title'"
}

func newRefLink(match string) (link RefLink, err error) {

	reqId, err := parseRefLinkReqId(match)
	if err != nil {
		return RefLink{}, err
	}

	link = RefLink{
		ReqId: reqId,
		Match: match,
	}

	return link, nil
}

func parseRefLinkReqId(linkText string) (reqId string, err error) {

	matches := _refLinkTextlineRegexp.FindStringSubmatch(linkText)

	// The first entry in the match is the whole string.
	// Then there are the right number of submatches.
	// Otherwise this is not a link.
	if len(matches) < 1 {
		return "", errors.WithStack(errors.Errorf(`Not a reference link: '%s'`, linkText))
	}

	// Get parts, reach into the indexes that are populated based on parenthesis in the regex.
	//fmt.Print(helper.JsonPretty(matches)) // To see the structure of the match.
	reqId = matches[1] // an ref id like: "F13"

	// Clean up.
	reqId = strings.ToUpper(reqId)

	return reqId, nil
}

func findRefLinks(text string) (refLinks []RefLink, err error) {

	// These need to be examined by textline since the regex matches an entire textline.
	adjustedText := strings.ReplaceAll(text, "\r", "\n") // Ensure only one kind of textline delimiter.
	textlines := strings.Split(adjustedText, "\n")

	for _, textline := range textlines {

		// Clean up any new lines.
		textline = strings.Trim(textline, "\n\r")

		match := _refLinkTextlineRegexp.FindString(textline)

		if match != "" {

			// The first element is the full matched text.
			refLink, err := newRefLink(match)
			if err != nil {
				return nil, err
			}

			refLinks = append(refLinks, refLink)
		}

	}

	return refLinks, nil
}
