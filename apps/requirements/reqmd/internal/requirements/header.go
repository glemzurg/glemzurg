package requirements

import (
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

const (
	Functional     = "F"
	NonFunctional  = "R"
	Actor          = "A"
	Stakeholder    = "S"
	UseCase        = "U"
	NonRequirement = "N"
)

var _KindSortValue = map[string]int{
	UseCase:        6, // Higher number sorts first.
	Actor:          5,
	Functional:     4,
	NonFunctional:  3,
	Stakeholder:    2,
	NonRequirement: 1,
}

type Header struct {
	Ref uint // The internal-to-parse identity of this requirement internally for a single pass of the tool

	// The content of the title.
	Prefix string // The "###" prefix for this header.
	Kind   string // what kind of requirement this is.
	Num    uint   // The specific id of this type.
	Title  string // The human name for this requirement.
}

func newHeader(ref uint, textline string) (header Header, err error) {

	err = validation.Validate(ref,
		validation.Required, // not empty
	)
	if err != nil {
		return Header{}, errors.WithStack(err)
	}
	err = validation.Validate(textline,
		validation.Required, // not empty
	)
	if err != nil {
		return Header{}, errors.WithStack(err)
	}

	prefix, kind, num, title, err := parseRequirementHeader(textline)
	if err != nil {
		return Header{}, err
	}

	header = Header{
		Prefix: prefix,
		Ref:    ref,
		Kind:   kind,
		Num:    num,
		Title:  title,
	}

	return header, nil
}

//===========================================
// Methods
//===========================================

func (h *Header) Id() (id string) {
	if h.Num == 0 {
		return h.Kind
	}
	return h.Kind + strconv.Itoa(int(h.Num))
}

func (h *Header) IdTitle() (value string) {
	return h.Id() + ". " + h.Title
}

func (h *Header) String() (value string) {
	return h.Prefix + " " + h.IdTitle()
}

func (h *Header) Link() (link string) {

	// Start with the generated header.
	link = h.String()

	// Turn puncs into whitespace.
	link = puncToWhitespace(link)

	// Lower case with no multiple spaces.
	link = strings.ToLower(normalizeWhitespace(link))

	// Remove the header nesting and leading space.
	link = strings.TrimLeft(link, "# ")

	// Final format with hyphens and hash.
	return "#" + strings.ReplaceAll(link, " ", "-")
}

//===========================================
// Parsing
//===========================================

func isRequirementHeader(textline string) (is bool) {
	return _reqHeaderRegexp.MatchString(textline)
}

func parseRequirementHeader(textline string) (prefix, kind string, num uint, title string, err error) {

	matches := _reqHeaderRegexp.FindAllStringSubmatch(textline, -1)

	// The first entry in the match is the whole string.
	// Then there are the right number of submatches.
	// Otherwise this is not a requirment header.
	if len(matches) != 1 || len(matches[0]) < 5 {
		return "", "", 0, "", errors.Errorf(`Not a requirement header: %+v`, textline)
	}

	// Get parts.
	prefix = matches[0][1]
	kind = matches[0][2]
	numRaw := matches[0][3]
	title = matches[0][4]

	// Clean up.
	kind = strings.ToUpper(kind)
	title = normalizeWhitespace(strings.TrimSpace(title))

	// Convert the num into the right format.
	if numRaw != "" {
		numInt, err := strconv.Atoi(numRaw)
		if err != nil {
			panic(err) // the regex should never allow this.
		}
		num = uint(numInt) // The regex should ensure this is always doable.
	}

	return prefix, kind, num, title, nil
}

//===========================================
// Sorting
//===========================================

func lessThan(a, b Header) (less bool) {

	// Sort by kind first.
	if a.Kind != b.Kind {
		kindASortValue := _KindSortValue[a.Kind]
		kindBSortValue := _KindSortValue[b.Kind]

		// Higher sort preference should appear first, so shoudl be "less" as sort algorithms think of it.
		return kindASortValue > kindBSortValue
	}

	// Sort by num next.

	// If they are the same, then return true since that is more efficient for the sort library.
	if a.Num == b.Num {
		return true
	}

	return a.Num < b.Num
}
