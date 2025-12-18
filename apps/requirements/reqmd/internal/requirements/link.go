package requirements

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Link struct {
	ReqId string // Value like "F13"
	Kind  string // what kind of requirement this is, like "F"
	Num   uint   // The specific id of this type, like 13
	Match string // The regex matched text for this link like "[f13](link)"
}

func newLink(match string) (link Link, err error) {

	reqId, kind, num, err := parseLinkReqId(match)
	if err != nil {
		return Link{}, err
	}

	link = Link{
		ReqId: reqId,
		Kind:  kind,
		Num:   num,
		Match: match,
	}

	return link, nil
}

func parseLinkReqId(linkText string) (reqId, kind string, num uint, err error) {

	matches := _linkRegexp.FindStringSubmatch(linkText)

	// The first entry in the match is the whole string.
	// Then there are the right number of submatches.
	// Otherwise this is not a link.
	if len(matches) < 1 {
		return "", "", 0, errors.WithStack(errors.Errorf(`Not a link: '%s'`, linkText))
	}

	// Get parts, reach into the indexes that are populated based on parenthesis in the regex.
	// fmt.Print(helper.JsonPretty(matches)) // To see the structure of the match.
	reqId = matches[1]      // an ref id like: "F13"
	kind = matches[2]       // an ref id like: "F"
	numString := matches[3] // an ref id like: 13
	linkId := matches[5]    // when the format is [F13][F13] and a ref id like: "F13"

	// Turn num into an integer.
	numInt, err := strconv.Atoi(numString)
	if err != nil {
		return "", "", 0, errors.WithStack(err)
	}
	num = uint(numInt)

	// Clean up.
	reqId = strings.ToUpper(reqId)
	kind = strings.ToUpper(kind)
	linkId = strings.ToUpper(linkId)

	// There may be no link.
	if linkId != "" {
		if reqId != linkId {
			return "", "", 0, errors.WithStack(errors.Errorf(`Link malformed: '%s'`, linkText))
		}
	}

	return reqId, kind, num, nil
}

func findLinks(text string) (links []Link, err error) {

	matches := _linkRegexp.FindAllStringSubmatch(text, -1)

	// Each top-level find is a match to a link.
	for _, match := range matches {

		// The first element is the full matched text.
		link, err := newLink(match[0])
		if err != nil {
			return nil, err
		}

		links = append(links, link)
	}

	return links, nil
}
