package parser_human

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"

	"github.com/pkg/errors"
)

// _DEFAULT_SUBDOMAIN_NAME is the name used for the default subdomain created for each domain.
const _DEFAULT_SUBDOMAIN_NAME = "default"

func Parse(modelPath string) (model core.Model, err error) {
	log.Printf("Parse files in '%s'", modelPath)

	// First, gather every thing we expect we need to parse, and what kind of entity it is.

	toParseFiles, err := findFilesToParse(modelPath)
	if err != nil {
		return core.Model{}, errors.WithStack(err)
	}

	for _, toParseFile := range toParseFiles {
		log.Println("   walk:", toParseFile.String())
	}

	model, err = parseForDatabase(modelPath, toParseFiles)
	if err != nil {
		return core.Model{}, errors.WithStack(err)
	}

	// Verify the model is well-formed after the parse.
	if err = model.Validate(); err != nil {
		return core.Model{}, errors.WithStack(err)
	}

	return model, nil
}

func findFilesToParse(modelPath string) (toParseFiles []fileToParse, err error) {
	// We need to walk from the root to deeper into the file tree to find everything.
	root, err := filepath.Abs(modelPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		// There was an error walking to this path.
		if err != nil {
			return errors.WithStack(err)
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return errors.WithStack(err)
		}

		// Everything that goes in the database is a file, not a directory.
		if d.IsDir() {
			return nil
		}

		toParseFile, err := newFileToParse(modelPath, relPath, path)
		if err != nil {
			return err
		}

		toParseFiles = append(toParseFiles, toParseFile)

		return nil
	})
	if err != nil {
		return nil, err
	}

	sortFilesToParse(toParseFiles)

	return toParseFiles, nil
}

// parseContext holds shared state used while parsing files into a model.
type parseContext struct {
	domainKeysBySubKey   map[string]identity.Key
	subdomainKeysByPath  map[string]identity.Key
	allClassAssociations map[identity.Key]model_class.Association
}

func parseForDatabase(modelKey string, filesToParse []fileToParse) (model core.Model, err error) {
	// Ensure are sorted in the order by which they may need information from prior objects.
	sortFilesToParse(filesToParse)

	ctx := &parseContext{
		domainKeysBySubKey:   map[string]identity.Key{},
		subdomainKeysByPath:  map[string]identity.Key{},
		allClassAssociations: map[identity.Key]model_class.Association{},
	}

	// Now, parse each file according to its type.
	for _, toParseFile := range filesToParse {
		contentBytes, err := os.ReadFile(toParseFile.PathAbs)
		if err != nil {
			return core.Model{}, errors.WithStack(err)
		}
		contents := string(contentBytes)

		switch toParseFile.FileType {
		case _EXT_MODEL:
			model, err = parseModelFile(modelKey, toParseFile, contents)
		case _EXT_ACTOR:
			err = parseActorFile(&model, toParseFile, contents)
		case _EXT_GENERALIZATION:
			err = parseGeneralizationFile(&model, ctx, toParseFile, contents)
		case _EXT_DOMAIN:
			err = parseDomainFile(&model, ctx, toParseFile, contents)
		case _EXT_SUBDOMAIN:
			err = parseSubdomainFile(&model, ctx, toParseFile, contents)
		case _EXT_CLASS:
			err = parseClassFile(&model, ctx, toParseFile, contents)
		case _EXT_USE_CASE:
			err = parseUseCaseFile(&model, ctx, toParseFile, contents)
		default:
			err = errors.WithStack(errors.Errorf(`unknown filetype: '%s'`, toParseFile.FileType))
		}
		if err != nil {
			return core.Model{}, err
		}
	}

	// Remove empty default subdomains that have no content.
	removeEmptyDefaultSubdomains(&model)

	// Distribute class associations to the correct level (model, domain, subdomain).
	if len(ctx.allClassAssociations) > 0 {
		if err := model.SetClassAssociations(ctx.allClassAssociations); err != nil {
			return core.Model{}, errors.Wrap(err, "failed to set class associations")
		}
	}

	// Phase 2: Re-create all ExpressionSpecs with full lowering context so that
	// Expression trees are populated via constructors.
	if err := convert.LowerAllExpressions(&model); err != nil {
		return core.Model{}, errors.Wrap(err, "failed to lower expressions")
	}

	return model, nil
}

// parseModelFile handles parsing a .model file and initializing model maps.
func parseModelFile(modelKey string, toParseFile fileToParse, contents string) (core.Model, error) {
	model, err := parseModel(modelKey, toParseFile.PathRel, contents)
	if err != nil {
		return core.Model{}, err
	}
	model.Actors = make(map[identity.Key]model_actor.Actor)
	model.ActorGeneralizations = make(map[identity.Key]model_actor.Generalization)
	model.Domains = make(map[identity.Key]model_domain.Domain)
	model.DomainAssociations = make(map[identity.Key]model_domain.Association)
	model.ClassAssociations = make(map[identity.Key]model_class.Association)
	return model, nil
}

// parseActorFile handles parsing an .actor file.
func parseActorFile(model *core.Model, toParseFile fileToParse, contents string) error {
	actor, err := parseActor(toParseFile.Actor, toParseFile.PathRel, contents)
	if err != nil {
		return err
	}
	model.Actors[actor.Key] = actor
	return nil
}

// parseGeneralizationFile handles parsing a .generalization file (actor, use case, or class).
func parseGeneralizationFile(model *core.Model, ctx *parseContext, toParseFile fileToParse, contents string) error {
	switch {
	case toParseFile.Domain == "":
		return parseActorGeneralizationFile(model, toParseFile, contents)
	case isUnderUseCases(toParseFile.PathRel):
		return parseUseCaseGeneralizationFile(model, ctx, toParseFile, contents)
	default:
		return parseClassGeneralizationFile(model, ctx, toParseFile, contents)
	}
}

// parseActorGeneralizationFile handles parsing an actor generalization.
func parseActorGeneralizationFile(model *core.Model, toParseFile fileToParse, contents string) error {
	actorGen, err := parseActorGeneralization(toParseFile.Generalization, toParseFile.PathRel, contents)
	if err != nil {
		return err
	}
	if model.ActorGeneralizations == nil {
		model.ActorGeneralizations = make(map[identity.Key]model_actor.Generalization)
	}
	model.ActorGeneralizations[actorGen.Key] = actorGen
	return nil
}

// lookupSubdomain resolves a subdomain from domain/subdomain names, returning the domain, subdomain, and their keys.
func lookupSubdomain(model *core.Model, ctx *parseContext, toParseFile fileToParse, entityDesc string) (model_domain.Domain, model_domain.Subdomain, identity.Key, identity.Key, error) {
	domainKey, ok := ctx.domainKeysBySubKey[toParseFile.Domain]
	if !ok {
		return model_domain.Domain{}, model_domain.Subdomain{}, identity.Key{}, identity.Key{}, errors.Errorf("domain '%s' not found for %s", toParseFile.Domain, entityDesc)
	}
	domain := model.Domains[domainKey]

	subdomainName := toParseFile.Subdomain
	if subdomainName == "" {
		subdomainName = _DEFAULT_SUBDOMAIN_NAME
	}
	subdomainKey, ok := ctx.subdomainKeysByPath[toParseFile.Domain+"/"+subdomainName]
	if !ok {
		return model_domain.Domain{}, model_domain.Subdomain{}, identity.Key{}, identity.Key{}, errors.Errorf("subdomain '%s' not found in domain '%s' for %s", subdomainName, toParseFile.Domain, entityDesc)
	}
	subdomain, ok := domain.Subdomains[subdomainKey]
	if !ok {
		return model_domain.Domain{}, model_domain.Subdomain{}, identity.Key{}, identity.Key{}, errors.Errorf("subdomain '%s' not found in domain '%s' for %s", subdomainName, toParseFile.Domain, entityDesc)
	}
	return domain, subdomain, domainKey, subdomainKey, nil
}

// updateSubdomain writes a modified subdomain back into the model.
func updateSubdomain(model *core.Model, domainKey, subdomainKey identity.Key, domain model_domain.Domain, subdomain model_domain.Subdomain) {
	domain.Subdomains[subdomainKey] = subdomain
	model.Domains[domainKey] = domain
}

// parseUseCaseGeneralizationFile handles parsing a use case generalization.
func parseUseCaseGeneralizationFile(model *core.Model, ctx *parseContext, toParseFile fileToParse, contents string) error {
	entityDesc := "use case generalization '" + toParseFile.Generalization + "'"
	domain, subdomain, domainKey, subdomainKey, err := lookupSubdomain(model, ctx, toParseFile, entityDesc)
	if err != nil {
		return err
	}

	generalization, err := parseUseCaseGeneralization(subdomainKey, toParseFile.Generalization, toParseFile.PathRel, contents)
	if err != nil {
		return err
	}
	if subdomain.UseCaseGeneralizations == nil {
		subdomain.UseCaseGeneralizations = make(map[identity.Key]model_use_case.Generalization)
	}
	subdomain.UseCaseGeneralizations[generalization.Key] = generalization
	updateSubdomain(model, domainKey, subdomainKey, domain, subdomain)
	return nil
}

// parseClassGeneralizationFile handles parsing a class generalization.
func parseClassGeneralizationFile(model *core.Model, ctx *parseContext, toParseFile fileToParse, contents string) error {
	entityDesc := "generalization '" + toParseFile.Generalization + "'"
	domain, subdomain, domainKey, subdomainKey, err := lookupSubdomain(model, ctx, toParseFile, entityDesc)
	if err != nil {
		return err
	}

	generalization, err := parseClassGeneralization(subdomainKey, toParseFile.Generalization, toParseFile.PathRel, contents)
	if err != nil {
		return err
	}
	if subdomain.Generalizations == nil {
		subdomain.Generalizations = make(map[identity.Key]model_class.Generalization)
	}
	subdomain.Generalizations[generalization.Key] = generalization
	updateSubdomain(model, domainKey, subdomainKey, domain, subdomain)
	return nil
}

// parseDomainFile handles parsing a .domain file.
func parseDomainFile(model *core.Model, ctx *parseContext, toParseFile fileToParse, contents string) error {
	domain, associations, err := parseDomain(toParseFile.Domain, toParseFile.PathRel, contents)
	if err != nil {
		return err
	}

	for _, assoc := range associations {
		model.DomainAssociations[assoc.Key] = assoc
	}

	defaultSubdomainKey, err := identity.NewSubdomainKey(domain.Key, _DEFAULT_SUBDOMAIN_NAME)
	if err != nil {
		return errors.WithStack(err)
	}
	subdomain, err := model_domain.NewSubdomain(defaultSubdomainKey, "Default", "", "")
	if err != nil {
		return err
	}
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		defaultSubdomainKey: subdomain,
	}

	model.Domains[domain.Key] = domain
	ctx.domainKeysBySubKey[toParseFile.Domain] = domain.Key
	ctx.subdomainKeysByPath[toParseFile.Domain+"/"+_DEFAULT_SUBDOMAIN_NAME] = defaultSubdomainKey
	return nil
}

// parseSubdomainFile handles parsing a .subdomain file.
func parseSubdomainFile(model *core.Model, ctx *parseContext, toParseFile fileToParse, contents string) error {
	domainKey, ok := ctx.domainKeysBySubKey[toParseFile.Domain]
	if !ok {
		return errors.Errorf("domain '%s' not found for subdomain '%s'", toParseFile.Domain, toParseFile.Subdomain)
	}
	domain := model.Domains[domainKey]

	subdomain, err := parseSubdomain(domainKey, toParseFile.Subdomain, toParseFile.PathRel, contents)
	if err != nil {
		return err
	}

	domain.Subdomains[subdomain.Key] = subdomain
	model.Domains[domainKey] = domain
	ctx.subdomainKeysByPath[toParseFile.Domain+"/"+toParseFile.Subdomain] = subdomain.Key
	return nil
}

// parseClassFile handles parsing a .class file.
func parseClassFile(model *core.Model, ctx *parseContext, toParseFile fileToParse, contents string) error {
	entityDesc := "class '" + toParseFile.Class + "'"
	domain, subdomain, domainKey, subdomainKey, err := lookupSubdomain(model, ctx, toParseFile, entityDesc)
	if err != nil {
		return err
	}

	classSubKey := toParseFile.Class
	if idx := len(toParseFile.Domain) + 1; idx < len(toParseFile.Class) {
		classSubKey = toParseFile.Class[idx:]
	}

	class, associations, err := parseClass(subdomainKey, classSubKey, toParseFile.PathRel, contents)
	if err != nil {
		return err
	}

	for _, assoc := range associations {
		ctx.allClassAssociations[assoc.Key] = assoc
	}

	if subdomain.Classes == nil {
		subdomain.Classes = make(map[identity.Key]model_class.Class)
	}
	subdomain.Classes[class.Key] = class
	updateSubdomain(model, domainKey, subdomainKey, domain, subdomain)
	return nil
}

// parseUseCaseFile handles parsing a .use_case file.
func parseUseCaseFile(model *core.Model, ctx *parseContext, toParseFile fileToParse, contents string) error {
	entityDesc := "use case '" + toParseFile.UseCase + "'"
	domain, subdomain, domainKey, subdomainKey, err := lookupSubdomain(model, ctx, toParseFile, entityDesc)
	if err != nil {
		return err
	}

	useCaseSubKey := toParseFile.UseCase
	if idx := len(toParseFile.Domain) + 1; idx < len(toParseFile.UseCase) {
		useCaseSubKey = toParseFile.UseCase[idx:]
	}

	useCase, err := parseUseCase(subdomainKey, useCaseSubKey, toParseFile.PathRel, contents)
	if err != nil {
		return err
	}

	if subdomain.UseCases == nil {
		subdomain.UseCases = make(map[identity.Key]model_use_case.UseCase)
	}
	subdomain.UseCases[useCase.Key] = useCase
	updateSubdomain(model, domainKey, subdomainKey, domain, subdomain)
	return nil
}

// removeEmptyDefaultSubdomains removes default subdomains that have no content.
func removeEmptyDefaultSubdomains(model *core.Model) {
	for domainKey, domain := range model.Domains {
		for subdomainKey, subdomain := range domain.Subdomains {
			if subdomainKey.SubKey == _DEFAULT_SUBDOMAIN_NAME && isEmptySubdomain(subdomain) {
				delete(domain.Subdomains, subdomainKey)
			}
		}
		if len(domain.Subdomains) == 0 {
			domain.Subdomains = nil
		}
		model.Domains[domainKey] = domain
	}
}

// isEmptySubdomain returns true if the subdomain has no classes, generalizations,
// use cases, use case generalizations, class associations, or use case shares.
func isEmptySubdomain(subdomain model_domain.Subdomain) bool {
	return len(subdomain.Classes) == 0 &&
		len(subdomain.Generalizations) == 0 &&
		len(subdomain.UseCases) == 0 &&
		len(subdomain.UseCaseGeneralizations) == 0 &&
		len(subdomain.ClassAssociations) == 0 &&
		len(subdomain.UseCaseShares) == 0
}
