package requirements

import "strings"

const (
	_ReferenceSectionIntro = "Referenced in:"
)

// Remove the reference section from the end of the requirements body.
func pruneReferenceSection(reqBody string) (noReferenceSection string, err error) {
	bodySplits := strings.Split(reqBody, _ReferenceSectionIntro)
	noReferenceSection = strings.TrimSpace(bodySplits[0])
	return noReferenceSection, nil
}
