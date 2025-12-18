package requirements

import (
	"path/filepath"
	"sort"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

type Requirement struct {
	Filename       string          // The file this requirement is in.
	Header         Header          // The id and title.
	Body           string          // The body of the requirement.
	Aspects        []Aspect        // Any aspects found on this requirement.
	Links          map[string]Link // The links in the requirement keyed by the link string, e.g. "F13".
	ReferencedFrom []Requirement   // The reqs that link to this one.
	Incompletes    []Incomplete    // Anything that makes this requirement incomplete.
}

func newRequirement(ref uint, filename, text string) (req Requirement, err error) {

	err = validation.Validate(filename,
		validation.Required, // not empty
	)
	if err != nil {
		return Requirement{}, errors.WithStack(err)
	}

	title, body := splitReq(text)

	header, err := newHeader(ref, title)
	if err != nil {
		return Requirement{}, err
	}

	body, aspects, err := extractAspects(body)
	if err != nil {
		return Requirement{}, err
	}

	body, err = pruneReferenceSection(body)
	if err != nil {
		return Requirement{}, err
	}

	var linkLookup map[string]Link
	links, err := findLinks(body)
	if err != nil {
		return Requirement{}, errors.WithStack(err)
	}
	if len(links) > 0 {
		linkLookup = map[string]Link{}
		for _, link := range links {
			linkLookup[link.ReqId] = link
		}
	}

	req = Requirement{
		Filename: filename,
		Header:   header,
		Body:     body,
		Aspects:  aspects,
		Links:    linkLookup,
	}

	return req, nil
}

//===========================================
// Methods
//===========================================

func (r *Requirement) String(aspects map[string][]string) (value string, err error) {

	var kindAspects []string
	if len(aspects) != 0 {
		kindAspects = aspects[r.Header.Kind]
	}

	aspectBlock, err := generateAspectBlock(kindAspects, r.Aspects)
	if err != nil {
		return "", err
	}

	referencedFromBlock, err := r.generateReferencesFromBlock()
	if err != nil {
		return "", err
	}

	value = r.Header.String()
	if r.Body != "" {
		value += "\n\n" + r.Body
	}
	if aspectBlock != "" {
		value += "\n\n" + aspectBlock
	}
	if referencedFromBlock != "" {
		value += "\n\n" + referencedFromBlock
	}
	value = strings.TrimSpace(value)

	return value, nil
}

func (r *Requirement) Id() (id string) {
	return r.Header.Id()
}

func (r *Requirement) Ref() (ref uint) {
	return r.Header.Ref
}

func (r *Requirement) RefLink(fromFilename string) (refLinkMarkdown string, err error) {

	// There is only a relative path if the filenames are different.
	var relativePath string
	if fromFilename != r.Filename {

		relativePath, err = filepath.Rel(filepath.Dir(fromFilename), filepath.Dir(r.Filename))
		if err != nil {
			return "", errors.WithStack(err)
		}

		// If this is the same path, then blank it out.
		// Otherwise add a filepath divider.
		if relativePath == "." {
			relativePath = "" // No path component to link.
		} else {
			relativePath += string(filepath.Separator)
		}

		// Not the same file so we need to add the filename.
		relativePath += filepath.Base(r.Filename)
	}

	// In the markdown the title is wrapped in single quotes. Strip any single quotes.
	titleWithoutSingleQuotes := strings.ReplaceAll(r.Header.Title, `'`, ``)

	refLinkMarkdown = `[` + r.Id() + `]: ` + relativePath + r.Header.Link() + ` '` + titleWithoutSingleQuotes + `'`

	return refLinkMarkdown, nil
}

func (r *Requirement) ReferencedFromLink(toFilename string) (referencedFromLinkMarkdown string, err error) {

	// There is only a relative path if the filenames are different.
	var relativePath string
	if toFilename != r.Filename {

		relativePath, err = filepath.Rel(filepath.Dir(toFilename), filepath.Dir(r.Filename))
		if err != nil {
			return "", errors.WithStack(err)
		}

		// If this is the same path, then blank it out.
		// Otherwise add a filepath divider.
		if relativePath == "." {
			relativePath = "" // No path component to link.
		} else {
			relativePath += string(filepath.Separator)
		}

		// Not the same file so we need to add the filename.
		relativePath += filepath.Base(r.Filename)
	}

	// In this markdown we cannot have "]" in the title text.
	titleWithoutCloseBrackets := strings.ReplaceAll(r.Header.Title, `]`, ``)

	referencedFromLinkMarkdown = `- [` + r.Id() + `. ` + titleWithoutCloseBrackets + `](` + relativePath + r.Header.Link() + `)`

	return referencedFromLinkMarkdown, nil
}

func (r *Requirement) generateReferencesFromBlock() (referencesBlock string, err error) {
	referencesReqs := r.ReferencedFrom

	// Sort the references.
	sort.Slice(referencesReqs, func(i, j int) bool {
		return lessThan(referencesReqs[i].Header, referencesReqs[j].Header)
	})

	if len(referencesReqs) > 0 {

		// Sort the references.
		sort.Slice(referencesReqs, func(i, j int) bool {
			if referencesReqs[i].Header.Kind == referencesReqs[j].Header.Kind {
				return referencesReqs[i].Header.Kind < referencesReqs[j].Header.Kind
			}
			return referencesReqs[i].Header.Num < referencesReqs[j].Header.Num
		})

		referencesBlock = _ReferenceSectionIntro + "\n"
		for _, referenceReq := range referencesReqs {
			backlinkMd, err := referenceReq.ReferencedFromLink(r.Filename)
			if err != nil {
				return "", err
			}
			referencesBlock += "\n" + backlinkMd
		}
	}

	return referencesBlock, nil
}

func (r *Requirement) generateIncompletes(reqRefs map[string]uint) (err error) {

	switch r.Header.Kind {
	case Stakeholder:
		r.Incompletes, err = r.incompleteStakeholder()
		if err != nil {
			return err
		}
	case Actor:
		r.Incompletes, err = r.incompleteActor()
		if err != nil {
			return err
		}
	case UseCase:
		r.Incompletes, err = r.incompleteUseCase()
		if err != nil {
			return err
		}
	case Functional:
		r.Incompletes, err = r.incompleteFunctional()
		if err != nil {
			return err
		}
	case NonFunctional:
	case NonRequirement:
	default:
		return errors.WithStack(errors.Errorf(`unknown kind: '%s'`, r.Header.Kind))
	}

	// Any requirement may have links to unknown.
	unknownRequirementLInks, err := r.incompleteUnknownRequirements(reqRefs)
	if err != nil {
		return err
	}
	r.Incompletes = append(r.Incompletes, unknownRequirementLInks...)

	// Any requirement may have unanswered questions.
	unasweredIncompletes, err := r.incompleteFromQuestion()
	if err != nil {
		return err
	}
	r.Incompletes = append(r.Incompletes, unasweredIncompletes...)

	return nil
}

func (r *Requirement) incompleteStakeholder() (incompletes []Incomplete, err error) {

	if r.Header.Kind != Stakeholder {
		return nil, errors.Errorf(`invalid kind, not '%s': '%s'`, Stakeholder, r.Header.Kind)
	}

	// Does this stakeholder link to any other requirements?

	links, err := findLinks(r.Body)
	if err != nil {
		return nil, err
	}

	if len(links) == 0 {
		incompletes = append(incompletes, newIncomplete(
			r.Header,
			StakeholderNotLinked,
			"",
		))
	}

	return incompletes, nil
}

func (r *Requirement) incompleteActor() (incompletes []Incomplete, err error) {

	if r.Header.Kind != Actor {
		return nil, errors.Errorf(`invalid kind, not '%s': '%s'`, Actor, r.Header.Kind)
	}

	foundRefUseCase := false
	for _, refReq := range r.ReferencedFrom {
		if refReq.Header.Kind == UseCase {
			foundRefUseCase = true
			break
		}
	}
	if !foundRefUseCase {
		incompletes = append(incompletes, newIncomplete(
			r.Header,
			ActorNotInUseCase,
			"",
		))
	}

	return incompletes, nil
}

func (r *Requirement) incompleteFunctional() (incompletes []Incomplete, err error) {

	if r.Header.Kind != Functional {
		return nil, errors.Errorf(`invalid kind, not '%s': '%s'`, Functional, r.Header.Kind)
	}

	foundRefUseCase := false
	for _, refReq := range r.ReferencedFrom {
		if refReq.Header.Kind == UseCase {
			foundRefUseCase = true
			break
		}
	}
	if !foundRefUseCase {
		incompletes = append(incompletes, newIncomplete(
			r.Header,
			FunctionalNotInUseCase,
			"",
		))
	}

	return incompletes, nil
}

func (r *Requirement) incompleteFromQuestion() (incompletes []Incomplete, err error) {

	// If there is a question mark, then assume it is an unanswered question that must be resolved.
	textlines := strings.Split(r.Body, "\n")
	for _, textline := range textlines {
		if strings.Contains(textline, "?") {
			incompletes = append(incompletes, newIncomplete(
				r.Header,
				UnansweredQuestion,
				textline,
			))
		}
	}

	return incompletes, nil
}

func (r *Requirement) incompleteUnknownRequirements(reqRefs map[string]uint) (incompletes []Incomplete, err error) {

	links, err := findLinks(r.Body)
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		if _, found := reqRefs[link.ReqId]; !found {
			incompletes = append(incompletes, newIncomplete(
				r.Header,
				UnknownRequirementLink,
				link.Match,
			))
		}
	}

	return incompletes, nil
}

//===========================================
// Parsing
//===========================================

func splitReq(recText string) (title, body string) {

	// Add a new line at the end to ensure split works.
	recText = recText + "\n"
	parts := strings.SplitN(recText, "\n", 2)
	title = strings.TrimSpace(parts[0])
	body = strings.TrimSpace(parts[1])
	return title, body
}
