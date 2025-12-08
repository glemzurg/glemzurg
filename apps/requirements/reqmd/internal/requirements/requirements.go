package requirements

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

const (
	_SUMMARY_FILE_NAME = "summary.generated.md"
)

type Requirements struct {
	path         string                     // What is the root path we're working out of.
	filenames    []string                   // The order that files were processed.
	files        map[string]fileParsed      // The files that compose the requirements.
	reqs         map[uint]Requirement       // The complete set of requirements.
	reqRefs      map[string]uint            // This is a mapping of requirment id ("F23") to refs. There could be a collision.
	referencedIn map[string]map[string]bool // The reverse lookup of which requirements mention another.
}

func New(path string) (req Requirements, err error) {

	err = validation.Validate(path,
		validation.Required, // not empty
	)
	if err != nil {
		return Requirements{}, errors.WithStack(err)
	}

	req = Requirements{
		path:    path,
		files:   map[string]fileParsed{},
		reqs:    map[uint]Requirement{},
		reqRefs: map[string]uint{},
	}

	// Populate the data.
	if err = req.populate(); err != nil {
		return Requirements{}, err
	}

	return req, nil
}

func (r *Requirements) populate() (err error) {

	var lastRef uint

	err = filepath.Walk(r.path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {

				// Is this a markdown file?
				if strings.HasSuffix(path, ".md") {

					fmt.Println(path)

					// Parse the file.
					var file fileParsed
					var fileReqs map[uint]Requirement
					lastRef, file, fileReqs, err = parseFile(lastRef, path)
					if err != nil {
						return err
					}

					// Add to requirements.
					r.filenames = append(r.filenames, path)
					r.files[path] = file
					for ref, req := range fileReqs {
						r.reqs[ref] = req
					}
				}
			}
			return nil
		})
	if err != nil {
		return errors.WithStack(err)
	}

	// Setup reference lookup.

	return nil
}

func (r *Requirements) mapReqsToRefs() (err error) {
	reqRefs := map[string]uint{}
	for _, req := range r.reqs {
		if _, found := reqRefs[req.Id()]; found {
			return errors.WithStack(errors.Errorf(`Same requirement found more than once: '%s'`, req.Id()))
		}
		reqRefs[req.Id()] = req.Ref()
	}
	r.reqRefs = reqRefs
	return nil
}

func (r *Requirements) NumberAll() (err error) {

	// Find the high water mark for each kind of requirment.
	highWaterMarks := map[string]uint{}
	for _, req := range r.reqs {
		if highWaterMarks[req.Header.Kind] < req.Header.Num {
			highWaterMarks[req.Header.Kind] = req.Header.Num
		}
	}

	// Loop through the files and give incrementing numbers to each requirement.
	for _, filename := range r.filenames {
		file := r.files[filename]
		for _, ref := range file.refs {
			req := r.reqs[ref]
			if req.Header.Num == 0 {

				// Get the last high water mark.
				lastNum := highWaterMarks[req.Header.Kind]

				// Next num.
				req.Header.Num = lastNum + 1

				// Save the changes.
				r.reqs[ref] = req
				highWaterMarks[req.Header.Kind] = req.Header.Num
			}
		}
	}

	// Recompoute all the references.
	if err = r.mapReqsToRefs(); err != nil {
		return err
	}

	return nil
}

func (r *Requirements) UpdateFiles(config Config) (err error) {

	// Loop through files, and determine if they need to be updated.
	for filename, file := range r.files {

		// What should the file contents be?
		newContents, err := file.Generate(config, r.reqs)
		if err != nil {
			return err
		}

		newContents, err = r.replaceAllLinks(filename, newContents)
		if err != nil {
			return err
		}

		// Any change to content?
		if newContents != file.originalContents {

			// What permissions are used by this file?
			fileInfo, err := os.Stat(filename)
			if err != nil {
				return errors.WithStack(err)
			}

			// Report what we're doing.
			fmt.Printf("  updating '%s'\n", filename)

			// Make the write.
			if err = os.WriteFile(filename, []byte(newContents), fileInfo.Mode().Perm()); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

func (r *Requirements) replaceAllLinks(filename, fileContents string) (updatedContents string, err error) {

	// Find every link in the file contents.
	links, err := findLinks(fileContents)
	if err != nil {
		return "", err
	}

	// Create a list of reference links for the bottom fo the file.
	var refLinkUrls []string

	// For every link found, do a find replace with the proper link text.
	for _, link := range links {

		// Determine what this link markdown should be.
		var refLinkMarkdown string

		// Get the requirement we're linking too.
		reqId := link.ReqId

		// The link is not guaranteed to go anywhere.
		ref, found := r.reqRefs[reqId]
		if found {
			req := r.reqs[ref]

			// What is the relative path link to this requirement.
			refLinkUrl, err := req.RefLink(filename)
			if err != nil {
				return "", err
			}
			refLinkUrls = append(refLinkUrls, refLinkUrl)

			// The link markdown is the bracketed id with the reference link label.
			refLinkMarkdown = "[" + reqId + "][" + reqId + "]"

		} else {

			// This link is for a non existent requirement.
			// Prune it back to be just the brackets alone.
			refLinkMarkdown = "[" + reqId + "][]"
		}

		// Is this a different link markdown then is there?
		if refLinkMarkdown != link.Match {

			// This link changed.
			// Do a find and replace
			// Links are not necessarily unique,
			// so it may have been already replaced.
			fileContents = strings.ReplaceAll(fileContents, link.Match, refLinkMarkdown)
		}
	}

	// The content was built from requirements without reference links.
	if len(refLinkUrls) > 0 {

		// Give a little whitespace.
		fileContents += "\n"

		// Re-add them.
		// Only add each link once.
		refLinkAdded := map[string]bool{}

		sort.Strings(refLinkUrls)
		for _, refLinkUrl := range refLinkUrls {
			if !refLinkAdded[refLinkUrl] {
				fileContents += "\n" + refLinkUrl
				refLinkAdded[refLinkUrl] = true
			}
		}
	}

	return fileContents, nil
}

func (r *Requirements) generateReferencedIn() {

	referencedIn := map[string]map[string]bool{}

	// Go through each requirement and gather the reverse lookup wiring them together.
	for _, req := range r.reqs {

		// What links are here?
		for _, link := range req.Links {

			// Pull out the reverse lookup.
			reverseLookup := referencedIn[link.ReqId]
			if reverseLookup == nil {
				reverseLookup = map[string]bool{}
			}

			reverseLookup[req.Id()] = true

			// Save the reverse lookup.
			referencedIn[link.ReqId] = reverseLookup
		}
	}

	r.referencedIn = referencedIn
}

func (r *Requirements) GenerateReferencedInLists() (err error) {

	// Create reverse lookup to links.
	r.generateReferencedIn()

	// Walk every requuirement and add a generated "reference in" block.
	for ref, req := range r.reqs {

		// Do other requirements reference this one with a link.
		referencesFrom := r.referencedIn[req.Id()]
		if len(referencesFrom) > 0 {

			// Create a list of requirements that mention this requirement.
			var referencesReqs []Requirement
			for fromId := range referencesFrom {
				fromReq := r.reqs[r.reqRefs[fromId]]
				referencesReqs = append(referencesReqs, fromReq)
			}

			// Add the references from other requirements to this one.
			req.ReferencedFrom = referencesReqs

			// Save the updated requirement docks.
			r.reqs[ref] = req
		}
	}

	return nil
}

func (r *Requirements) GenerateIncompletes() (err error) {

	// Walk every requuirement and add a generated "reference in" block.
	for ref, req := range r.reqs {

		if err = req.generateIncompletes(r.reqRefs); err != nil {
			return err
		}

		// Save the updated requirement docks.
		r.reqs[ref] = req
	}

	return nil
}

func (r *Requirements) WriteSummaryFile(path string) (err error) {

	// A fixed filename.
	filename := filepath.Join(path, _SUMMARY_FILE_NAME)

	// What are the contents of the file.

	contents := `# Summary

## Incomplete
`

	// Each kind of requirement has different rules about what makes it incomplete.

	// Get all the requirements and sort them.
	var sortedReqs []Requirement
	for _, req := range r.reqs {
		sortedReqs = append(sortedReqs, req)
	}
	sort.Slice(sortedReqs, func(i, j int) bool {
		return lessThan(sortedReqs[i].Header, sortedReqs[j].Header)
	})

	// Report any that are incomplete.
	for _, req := range sortedReqs {
		if len(req.Incompletes) > 0 {
			contents += "\n- " + req.Header.IdTitle()
			for _, incomplete := range req.Incompletes {
				contents += "\n\t- " + incomplete.String()
			}
		}
	}

	// Report what we're doing.
	fmt.Printf("  updating '%s'\n", filename)

	// Make the write.
	if err = os.WriteFile(filename, []byte(contents), 0644); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
