package generate

import (
	"io"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
)

// ContentWriter is an interface for writing generated content.
type ContentWriter interface {
	WriteMarkdown(filename string, content []byte) error
	WriteSVG(filename string, content []byte) error
	WriteCSS(content []byte) error
}

// GenerateMdToWriter generates markdown documentation to a ContentWriter.
func GenerateMdToWriter(parsedModel core.Model, writer ContentWriter) error { //nolint:revive // public API name
	// Create the flattened requirements from the model.
	reqs := req_flat.NewRequirements(parsedModel)

	// Prepare the convenience structures inside.
	reqs.PrepLookups()

	// Generate files to writer.
	return generateFilesToWriter(reqs, writer)
}

// generateFilesToWriter generates all files and writes them to the ContentWriter.
func generateFilesToWriter(reqs *req_flat.Requirements, writer ContentWriter) error {
	// Write CSS
	if err := writer.WriteCSS([]byte(_MD_CSS)); err != nil {
		return err
	}

	// Write support images
	if err := writeSupportImagesToWriter(writer); err != nil {
		return err
	}

	// Generate model files
	if err := generateModelFilesToWriter(reqs, writer); err != nil {
		return err
	}

	// Generate actor files
	if err := generateActorFilesToWriter(reqs, writer); err != nil {
		return err
	}

	// Generate domain files
	if err := generateDomainFilesToWriter(reqs, writer); err != nil {
		return err
	}

	// Generate subdomain files
	if err := generateSubdomainFilesToWriter(reqs, writer); err != nil {
		return err
	}

	// Generate class files
	if err := generateClassFilesToWriter(reqs, writer); err != nil {
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
	_, _ = w.Write([]byte(_MD_CSS))
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
func generateModelFilesToWriter(reqs *req_flat.Requirements, writer ContentWriter) error {
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

	// Generate domains Mermaid diagram.
	domainsDiagram, err := generateDomainsMermaidContents(reqs, domains, associations)
	if err != nil {
		return err
	}

	// Generate model summary markdown with embedded diagram.
	mdContents, err := generateModelMdContents(reqs, model, actors, domains, domainsDiagram)
	if err != nil {
		return err
	}
	if err := writer.WriteMarkdown("model.md", []byte(mdContents)); err != nil {
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
func generateDomainFilesToWriter(reqs *req_flat.Requirements, writer ContentWriter) error {
	domainLookup, _ := reqs.DomainLookup()

	for _, domain := range domainLookup {
		diagrams, err := buildDomainDiagrams(reqs, domain)
		if err != nil {
			return err
		}

		// Generate domain markdown page with embedded diagrams.
		modelFilename := convertKeyToFilename("domain", domain.Key.String(), "", ".md")
		mdContents, err := generateDomainMdContents(reqs, reqs.Model, domain, diagrams)
		if err != nil {
			return err
		}
		if err := writer.WriteMarkdown(modelFilename, []byte(mdContents)); err != nil {
			return err
		}
	}

	return nil
}

// buildDomainDiagrams generates the Mermaid diagrams needed for a domain page.
func buildDomainDiagrams(reqs *req_flat.Requirements, domain model_domain.Domain) (domainDiagrams, error) {
	hasMultipleSubdomains := len(domain.Subdomains) > 1

	if hasMultipleSubdomains {
		subdomainsDiagram, err := generateSubdomainsMermaidContents(reqs, domain)
		if err != nil {
			return domainDiagrams{}, err
		}
		return domainDiagrams{SubdomainsDiagram: subdomainsDiagram}, nil
	}

	// Single subdomain: generate classes and use cases diagrams.
	classesDiagram, err := buildClassesDiagram(reqs, gatherDomainClasses(domain))
	if err != nil {
		return domainDiagrams{}, err
	}

	useCasesDiagram, err := buildUseCasesDiagram(reqs, domain)
	if err != nil {
		return domainDiagrams{}, err
	}

	return domainDiagrams{ClassesDiagram: classesDiagram, UseCasesDiagram: useCasesDiagram}, nil
}

// gatherDomainClasses collects all classes from all subdomains of a domain.
func gatherDomainClasses(domain model_domain.Domain) []model_class.Class {
	var classes []model_class.Class
	for _, subdomain := range domain.Subdomains {
		for _, class := range subdomain.Classes {
			classes = append(classes, class)
		}
	}
	return classes
}

// buildClassesDiagram generates a Mermaid class diagram for a set of classes.
func buildClassesDiagram(reqs *req_flat.Requirements, classes []model_class.Class) (string, error) {
	generalizations, allClasses, associations := reqs.RegardingClasses(classes)
	return generateClassesMermaidContents(reqs, generalizations, allClasses, associations)
}

// buildUseCasesDiagram generates a Mermaid use case diagram for a domain.
func buildUseCasesDiagram(reqs *req_flat.Requirements, domain model_domain.Domain) (string, error) {
	var useCases []model_use_case.UseCase
	for _, subdomain := range domain.Subdomains {
		for _, useCase := range subdomain.UseCases {
			useCases = append(useCases, useCase)
		}
	}
	relevantUseCases, relevantActors, err := reqs.RegardingUseCases(useCases)
	if err != nil {
		return "", err
	}
	return generateUseCasesMermaidContents(reqs, domain, relevantUseCases, relevantActors)
}

// generateSubdomainFilesToWriter generates subdomain files to a ContentWriter.
func generateSubdomainFilesToWriter(reqs *req_flat.Requirements, writer ContentWriter) error {
	domainLookup, _ := reqs.DomainLookup()

	for _, domain := range domainLookup {
		// Skip if only one subdomain.
		if len(domain.Subdomains) <= 1 {
			continue
		}

		for _, subdomain := range domain.Subdomains {
			if err := generateSingleSubdomainFiles(reqs, writer, domain, subdomain); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateSingleSubdomainFiles generates all files for a single subdomain.
func generateSingleSubdomainFiles(reqs *req_flat.Requirements, writer ContentWriter, domain model_domain.Domain, subdomain model_domain.Subdomain) error {
	// Generate classes diagram for subdomain.
	var subdomainClasses []model_class.Class
	for _, class := range subdomain.Classes {
		subdomainClasses = append(subdomainClasses, class)
	}
	classesDiagram, err := buildClassesDiagram(reqs, subdomainClasses)
	if err != nil {
		return err
	}

	// Generate use cases diagram for subdomain.
	useCasesDiagram, err := buildSubdomainUseCasesDiagram(reqs, domain, subdomain)
	if err != nil {
		return err
	}

	// Generate subdomain markdown page with embedded diagrams.
	subdomainFilename := convertKeyToFilename("subdomain", subdomain.Key.String(), "", ".md")
	mdContents, err := generateSubdomainMdContents(reqs, reqs.Model, domain, subdomain, classesDiagram, useCasesDiagram)
	if err != nil {
		return err
	}
	if err := writer.WriteMarkdown(subdomainFilename, []byte(mdContents)); err != nil {
		return err
	}

	return nil
}

// buildSubdomainUseCasesDiagram generates a Mermaid use case diagram for a subdomain.
func buildSubdomainUseCasesDiagram(reqs *req_flat.Requirements, domain model_domain.Domain, subdomain model_domain.Subdomain) (string, error) {
	var useCases []model_use_case.UseCase
	for _, useCase := range subdomain.UseCases {
		useCases = append(useCases, useCase)
	}
	relevantUseCases, relevantActors, err := reqs.RegardingUseCases(useCases)
	if err != nil {
		return "", err
	}
	return generateUseCasesMermaidContents(reqs, domain, relevantUseCases, relevantActors)
}

// generateClassFilesToWriter generates class files to a ContentWriter.
func generateClassFilesToWriter(reqs *req_flat.Requirements, writer ContentWriter) error {
	classLookup, _ := reqs.ClassLookup()

	for _, class := range classLookup {
		// Generate classes Mermaid diagram.
		generalizations, classes, associations := reqs.RegardingClasses([]model_class.Class{class})
		classesDiagram, err := generateClassesMermaidContents(reqs, generalizations, classes, associations)
		if err != nil {
			return err
		}

		// Generate state machine Mermaid diagram if applicable.
		stateDiagram := ""
		if len(class.States) > 0 {
			stateDiagram, err = generateClassStateMermaidContents(reqs, class)
			if err != nil {
				return err
			}
		}

		// Generate class summary markdown with embedded diagrams.
		classFilename := convertKeyToFilename("class", class.Key.String(), "", ".md")
		classMdContents, err := generateClassMdContents(reqs, class, classesDiagram, stateDiagram)
		if err != nil {
			return err
		}
		if err := writer.WriteMarkdown(classFilename, []byte(classMdContents)); err != nil {
			return err
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
