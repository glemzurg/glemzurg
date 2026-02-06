package generate

import (
	"io"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
)

// ContentWriter is an interface for writing generated content.
type ContentWriter interface {
	WriteMarkdown(filename string, content []byte) error
	WriteSVG(filename string, content []byte) error
	WriteCSS(content []byte) error
}

// GenerateMdToWriter generates markdown documentation to a ContentWriter.
func GenerateMdToWriter(debug bool, parsedModel req_model.Model, writer ContentWriter) error {
	// Create the flattened requirements from the model.
	reqs := req_flat.NewRequirements(parsedModel)

	// Prepare the convenience structures inside.
	reqs.PrepLookups()

	// Generate files to writer.
	return generateFilesToWriter(debug, reqs, writer)
}

// generateFilesToWriter generates all files and writes them to the ContentWriter.
func generateFilesToWriter(debug bool, reqs *req_flat.Requirements, writer ContentWriter) error {
	// Write CSS
	if err := writer.WriteCSS([]byte(_MD_CSS)); err != nil {
		return err
	}

	// Write support images
	if err := writeSupportImagesToWriter(writer); err != nil {
		return err
	}

	// Generate model files
	if err := generateModelFilesToWriter(debug, reqs, writer); err != nil {
		return err
	}

	// Generate actor files
	if err := generateActorFilesToWriter(reqs, writer); err != nil {
		return err
	}

	// Generate domain files
	if err := generateDomainFilesToWriter(debug, reqs, writer); err != nil {
		return err
	}

	// Generate subdomain files
	if err := generateSubdomainFilesToWriter(debug, reqs, writer); err != nil {
		return err
	}

	// Generate class files
	if err := generateClassFilesToWriter(debug, reqs, writer); err != nil {
		return err
	}

	// Generate use case files
	if err := generateUseCaseFilesToWriter(reqs, writer); err != nil {
		return err
	}

	// Generate scenario files
	if err := generateScenarioFilesToWriter(reqs, writer); err != nil {
		return err
	}

	return nil
}

// WriteCSS writes CSS content to an io.Writer.
func WriteCSS(w io.Writer) {
	w.Write([]byte(_MD_CSS))
}

// writeSupportImagesToWriter writes support images to a ContentWriter.
func writeSupportImagesToWriter(writer ContentWriter) error {
	if err := writer.WriteSVG("person.svg", []byte(_ACTOR_PERSON_SVG)); err != nil {
		return err
	}
	if err := writer.WriteSVG("system.svg", []byte(_ACTOR_SYSTEM_SVG)); err != nil {
		return err
	}
	return nil
}

// generateModelFilesToWriter generates model files to a ContentWriter.
func generateModelFilesToWriter(debug bool, reqs *req_flat.Requirements, writer ContentWriter) error {
	model := reqs.Model

	// Get actors as a sorted slice.
	var actors []model_actor.Actor
	for _, actor := range reqs.Actors {
		actors = append(actors, actor)
	}
	sort.Slice(actors, func(i, j int) bool {
		return actors[i].Key.String() < actors[j].Key.String()
	})

	// Get domains as a sorted slice.
	var domains []model_domain.Domain
	for _, domain := range reqs.Domains {
		domains = append(domains, domain)
	}
	sort.Slice(domains, func(i, j int) bool {
		return domains[i].Key.String() < domains[j].Key.String()
	})

	// Get domain associations as a sorted slice.
	var associations []model_domain.Association
	for _, assoc := range reqs.DomainAssociations {
		associations = append(associations, assoc)
	}
	sort.Slice(associations, func(i, j int) bool {
		return associations[i].Key.String() < associations[j].Key.String()
	})

	// Generate model summary markdown.
	mdContents, err := generateModelMdContents(reqs, model, actors, domains)
	if err != nil {
		return err
	}
	if err := writer.WriteMarkdown("model.md", []byte(mdContents)); err != nil {
		return err
	}

	// Generate domains diagram.
	svgContents, _, err := generateDomainsSvgContents(reqs, domains, associations)
	if err != nil {
		return err
	}
	if err := writer.WriteSVG("domains.svg", []byte(svgContents)); err != nil {
		return err
	}

	return nil
}

// generateActorFilesToWriter generates actor files to a ContentWriter.
func generateActorFilesToWriter(reqs *req_flat.Requirements, writer ContentWriter) error {
	actorLookup := reqs.ActorLookup()

	for _, actor := range actorLookup {
		modelFilename := convertKeyToFilename("actor", actor.Key.String(), "", ".md")
		mdContents, err := generateActorMdContents(reqs, actor)
		if err != nil {
			return err
		}
		if err := writer.WriteMarkdown(modelFilename, []byte(mdContents)); err != nil {
			return err
		}
	}

	return nil
}

// generateDomainFilesToWriter generates domain files to a ContentWriter.
func generateDomainFilesToWriter(debug bool, reqs *req_flat.Requirements, writer ContentWriter) error {
	domainLookup, _ := reqs.DomainLookup()

	for _, domain := range domainLookup {
		hasMultipleSubdomains := len(domain.Subdomains) > 1

		// Generate domain markdown page.
		modelFilename := convertKeyToFilename("domain", domain.Key.String(), "", ".md")
		mdContents, err := generateDomainMdContents(reqs, reqs.Model, domain)
		if err != nil {
			return err
		}
		if err := writer.WriteMarkdown(modelFilename, []byte(mdContents)); err != nil {
			return err
		}

		if hasMultipleSubdomains {
			// Generate subdomains diagram.
			subdomainsFilename := convertKeyToFilename("domain", domain.Key.String(), "subdomains", ".svg")
			svgContents, _, err := generateSubdomainsSvgContents(reqs, domain)
			if err != nil {
				return err
			}
			if err := writer.WriteSVG(subdomainsFilename, []byte(svgContents)); err != nil {
				return err
			}
		} else {
			// Single subdomain: generate use cases and classes diagrams at domain level.
			var domainClasses []model_class.Class
			for _, subdomain := range domain.Subdomains {
				for _, class := range subdomain.Classes {
					domainClasses = append(domainClasses, class)
				}
			}

			// Generate use cases diagram.
			useCasesFilename := convertKeyToFilename("domain", domain.Key.String(), "use-cases", ".svg")
			var domainUseCases []model_use_case.UseCase
			for _, subdomain := range domain.Subdomains {
				for _, useCase := range subdomain.UseCases {
					domainUseCases = append(domainUseCases, useCase)
				}
			}

			relevantUseCases, relevantActors, err := reqs.RegardingUseCases(domainUseCases)
			if err != nil {
				return err
			}
			useCasesSvgContents, _, err := generateUseCasesSvgContents(reqs, domain, relevantUseCases, relevantActors)
			if err != nil {
				return err
			}
			if err := writer.WriteSVG(useCasesFilename, []byte(useCasesSvgContents)); err != nil {
				return err
			}

			// Generate classes diagram.
			generalizations, classes, classAssociations := reqs.RegardingClasses(domainClasses)
			classesFilename := convertKeyToFilename("domain", domain.Key.String(), "classes", ".svg")
			classesSvgContents, _, err := generateClassesSvgContents(reqs, generalizations, classes, classAssociations)
			if err != nil {
				return err
			}
			if err := writer.WriteSVG(classesFilename, []byte(classesSvgContents)); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateSubdomainFilesToWriter generates subdomain files to a ContentWriter.
func generateSubdomainFilesToWriter(debug bool, reqs *req_flat.Requirements, writer ContentWriter) error {
	domainLookup, _ := reqs.DomainLookup()

	for _, domain := range domainLookup {
		// Skip if only one subdomain.
		if len(domain.Subdomains) <= 1 {
			continue
		}

		for _, subdomain := range domain.Subdomains {
			// Generate subdomain markdown page.
			subdomainFilename := convertKeyToFilename("subdomain", subdomain.Key.String(), "", ".md")
			mdContents, err := generateSubdomainMdContents(reqs, reqs.Model, domain, subdomain)
			if err != nil {
				return err
			}
			if err := writer.WriteMarkdown(subdomainFilename, []byte(mdContents)); err != nil {
				return err
			}

			// Gather classes for this subdomain.
			var subdomainClasses []model_class.Class
			for _, class := range subdomain.Classes {
				subdomainClasses = append(subdomainClasses, class)
			}

			// Generate use cases diagram for subdomain.
			useCasesFilename := convertKeyToFilename("subdomain", subdomain.Key.String(), "use-cases", ".svg")
			var subdomainUseCases []model_use_case.UseCase
			for _, useCase := range subdomain.UseCases {
				subdomainUseCases = append(subdomainUseCases, useCase)
			}

			relevantUseCases, relevantActors, err := reqs.RegardingUseCases(subdomainUseCases)
			if err != nil {
				return err
			}
			useCasesSvgContents, _, err := generateUseCasesSvgContents(reqs, domain, relevantUseCases, relevantActors)
			if err != nil {
				return err
			}
			if err := writer.WriteSVG(useCasesFilename, []byte(useCasesSvgContents)); err != nil {
				return err
			}

			// Generate classes diagram.
			generalizations, classes, associations := reqs.RegardingClasses(subdomainClasses)
			classesFilename := convertKeyToFilename("subdomain", subdomain.Key.String(), "classes", ".svg")
			classesSvgContents, _, err := generateClassesSvgContents(reqs, generalizations, classes, associations)
			if err != nil {
				return err
			}
			if err := writer.WriteSVG(classesFilename, []byte(classesSvgContents)); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateClassFilesToWriter generates class files to a ContentWriter.
func generateClassFilesToWriter(debug bool, reqs *req_flat.Requirements, writer ContentWriter) error {
	classLookup, _ := reqs.ClassLookup()

	for _, class := range classLookup {
		// Generate class summary markdown.
		classFilename := convertKeyToFilename("class", class.Key.String(), "", ".md")
		classMdContents, err := generateClassMdContents(reqs, class)
		if err != nil {
			return err
		}
		if err := writer.WriteMarkdown(classFilename, []byte(classMdContents)); err != nil {
			return err
		}

		// Generate classes diagram.
		generalizations, classes, associations := reqs.RegardingClasses([]model_class.Class{class})
		classesSvgFilename := convertKeyToFilename("class", class.Key.String(), "", ".svg")
		classesSvgContents, _, err := generateClassesSvgContents(reqs, generalizations, classes, associations)
		if err != nil {
			return err
		}
		if err := writer.WriteSVG(classesSvgFilename, []byte(classesSvgContents)); err != nil {
			return err
		}

		// State Machine diagram.
		if len(class.States) > 0 {
			statesSvgFilename := convertKeyToFilename("class", class.Key.String(), "states", ".svg")
			statesSvgContents, _, err := generateClassStateSvgContents(reqs, class)
			if err != nil {
				return err
			}
			if err := writer.WriteSVG(statesSvgFilename, []byte(statesSvgContents)); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateUseCaseFilesToWriter generates use case files to a ContentWriter.
func generateUseCaseFilesToWriter(reqs *req_flat.Requirements, writer ContentWriter) error {
	useCaseLookup := reqs.UseCaseLookup()

	for _, useCase := range useCaseLookup {
		modelFilename := convertKeyToFilename("use_case", useCase.Key.String(), "", ".md")
		mdContents, err := generateUseCaseMdContents(reqs, useCase)
		if err != nil {
			return err
		}
		if err := writer.WriteMarkdown(modelFilename, []byte(mdContents)); err != nil {
			return err
		}
	}

	return nil
}

// generateScenarioFilesToWriter generates scenario files to a ContentWriter.
func generateScenarioFilesToWriter(reqs *req_flat.Requirements, writer ContentWriter) error {
	scenarioLookup := reqs.ScenarioLookup()

	for _, scenario := range scenarioLookup {
		svgFilename := convertKeyToFilename("scenario", scenario.Key.String(), "", ".svg")
		svgContents, err := generateScenarioSvgContents(reqs, scenario)
		if err != nil {
			return err
		}
		if err := writer.WriteSVG(svgFilename, []byte(svgContents)); err != nil {
			return err
		}
	}

	return nil
}
