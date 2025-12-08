package requirements

import (
	"strings"

	"github.com/pkg/errors"
)

func (r *Requirement) incompleteUseCase() (incompletes []Incomplete, err error) {

	if r.Header.Kind != UseCase {
		return nil, errors.Errorf(`invalid kind, not '%s': '%s'`, UseCase, r.Header.Kind)
	}

	// No bullets discovered?
	bulletsFound := false

	// Examine the requirement body line by line.
	textlines := strings.Split(r.Body, "\n")
	for _, textline := range textlines {
		if isNumericBulletedLine(textline) || isBulletedLine(textline) {
			bulletsFound = true

			// If a line ends with a colon, it's structural and not an action step.
			// This is used for repeating "loops" in a use case.
			endsWithColon := strings.HasSuffix(strings.TrimSpace(textline), ":")

			// The pattern that indicates a step is out of scope for system is "[x]".
			endsWithOutOfScope := strings.HasSuffix(strings.TrimSpace(textline), "[x]")

			if !(endsWithColon || endsWithOutOfScope) {

				links, err := findLinks(textline)
				if err != nil {
					return nil, err
				}

				// Look for functional requirements and use cases.
				functionalOrUseCaseReqFound := false
				for _, link := range links {
					if link.Kind == Functional {
						functionalOrUseCaseReqFound = true
						break
					}
					if link.Kind == UseCase {
						functionalOrUseCaseReqFound = true
						break
					}
				}

				if !functionalOrUseCaseReqFound {
					incompletes = append(incompletes, newIncomplete(
						r.Header,
						UseCaseStepMissingFunctionalRequirementOrUseCase,
						textline,
					))
				}
			}
		}
	}

	if !bulletsFound {
		incompletes = append(incompletes, newIncomplete(r.Header, UseCaseNoBullets, ""))
	}

	// Are there any actors mentioned anywhere in the use case?
	actorsFound := false
	links, err := findLinks(r.Body)
	if err != nil {
		return nil, err
	}
	for _, link := range links {
		if link.Kind == Actor {
			actorsFound = true
			break
		}
	}
	if !actorsFound {
		incompletes = append(incompletes, newIncomplete(r.Header, UseCaseNoActor, ""))
	}

	return incompletes, nil
}

func isBulletedLine(textline string) (is bool) {
	return _markdownBulletLineRegexp.MatchString(textline)
}

func isNumericBulletedLine(textline string) (is bool) {
	return _markdownNumericBulletLineRegexp.MatchString(textline)
}
