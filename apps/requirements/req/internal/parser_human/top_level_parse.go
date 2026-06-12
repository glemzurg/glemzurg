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

// Parse reads a model from its source directory.
//
// A parse failure in a single .class file does not abort the whole model: that
// class becomes an empty placeholder and its error is returned in the
// []ParseFailure slice, so every other entity still renders. A failure in a
// .model / .domain / .subdomain file (entities that others depend on) or a
// filesystem walk error is catastrophic and returned as the error.
func Parse(modelPath string) (model core.Model, failures []ParseFailure, err error) {
	log.Printf("Parse files in '%s'", modelPath)

	// First, gather every thing we expect we need to parse, and what kind of entity it is.

	toParseFiles, err := findFilesToParse(modelPath)
	if err != nil {
		return core.Model{}, nil, errors.WithStack(err)
	}

	for _, toParseFile := range toParseFiles {
		log.Println("   walk:", toParseFile.String())
	}

	model, failures, err = parseForDatabase(modelPath, toParseFiles)
	if err != nil {
		return core.Model{}, nil, errors.WithStack(err)
	}

	// Verify the model is well-formed after the parse.
	if err = model.Validate(); err != nil {
		// With parse failures the model is known-partial — a placeholder class
		// can, for example, drop a generalization's superclass linkage. Treat
		// that validation error as a symptom, not catastrophic: hand back the
		// partial model + failures so the rest still renders.
		if len(failures) > 0 {
			log.Printf("   model validation reported issues (expected with %d parse failure(s)): %v", len(failures), err)
			return model, failures, nil
		}
		return core.Model{}, nil, errors.WithStack(err)
	}

	return model, failures, nil
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

func parseForDatabase(modelKey string, filesToParse []fileToParse) (model core.Model, failures []ParseFailure, err error) {
	// Ensure are sorted in the order by which they may need information from prior objects.
	sortFilesToParse(filesToParse)

	ctx := &parseContext{
		domainKeysBySubKey:   map[string]identity.Key{},
		subdomainKeysByPath:  map[string]identity.Key{},
		allClassAssociations: map[identity.Key]model_class.Association{},
	}

	// Parse each file according to its type.
	model, failures, err = parseAllFiles(modelKey, filesToParse, ctx)
	if err != nil {
		return core.Model{}, nil, err
	}

	// Post-processing: finalize the model.
	if err := finalizeModel(&model, ctx); err != nil {
		return core.Model{}, nil, err
	}

	return model, failures, nil
}

// parseAllFiles reads and parses each file in order, dispatching by file type.
// A .class file that fails to parse is isolated (placeholder class + recorded
// failure) so it does not abort the rest of the model.
func parseAllFiles(modelKey string, filesToParse []fileToParse, ctx *parseContext) (core.Model, []ParseFailure, error) {
	var model core.Model
	var failures []ParseFailure
	for _, toParseFile := range filesToParse {
		contentBytes, err := os.ReadFile(toParseFile.PathAbs)
		if err != nil {
			return core.Model{}, nil, errors.WithStack(err)
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
			var failure *ParseFailure
			failure, err = parseClassFileResilient(&model, ctx, toParseFile, contents)
			if failure != nil {
				log.Printf("   parse failure: %s: %s", failure.Path, failure.Err)
				failures = append(failures, *failure)
			}
		case _EXT_USE_CASE:
			err = parseUseCaseFile(&model, ctx, toParseFile, contents)
		default:
			err = errors.WithStack(errors.Errorf(`unknown filetype: '%s'`, toParseFile.FileType))
		}
		if err != nil {
			return core.Model{}, nil, err
		}
	}
	return model, failures, nil
}

// finalizeModel performs post-processing on the parsed model.
func finalizeModel(model *core.Model, ctx *parseContext) error {
	removeEmptyDefaultSubdomains(model)

	if len(ctx.allClassAssociations) > 0 {
		if err := model.SetClassAssociations(ctx.allClassAssociations); err != nil {
			return errors.Wrap(err, "failed to set class associations")
		}
	}

	if err := convert.LowerAllExpressions(model); err != nil {
		return errors.Wrap(err, "failed to lower expressions")
	}

	return nil
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
	subdomain := model_domain.NewSubdomain(defaultSubdomainKey, "Default", "", "", "")
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

// parseClassFileResilient parses a .class file. If the class body fails to
// parse, it does not abort the model: a placeholder empty class is added (so
// the class still lists, unpopulated, on its domain page) and a ParseFailure
// is returned. A failure that prevents even placing the class (e.g. its
// subdomain is missing) is returned as a hard error.
func parseClassFileResilient(model *core.Model, ctx *parseContext, toParseFile fileToParse, contents string) (*ParseFailure, error) {
	entityDesc := "class '" + toParseFile.Class + "'"
	domain, subdomain, domainKey, subdomainKey, err := lookupSubdomain(model, ctx, toParseFile, entityDesc)
	if err != nil {
		return nil, err
	}

	classSubKey := toParseFile.Class
	if idx := len(toParseFile.Domain) + 1; idx < len(toParseFile.Class) {
		classSubKey = toParseFile.Class[idx:]
	}

	class, associations, parseErr := safeParseClass(subdomainKey, classSubKey, toParseFile.PathRel, contents)
	if parseErr != nil {
		// Isolate the failure: place a valid empty class so the rest of the
		// model still parses, validates, and renders.
		placeholder, phErr := placeholderClass(subdomainKey, classSubKey)
		if phErr != nil {
			return nil, parseErr // Can't place the class — surface the parse error.
		}
		addClassToSubdomain(model, domainKey, subdomainKey, domain, subdomain, placeholder)
		return &ParseFailure{
			ClassKey: placeholder.Key,
			Name:     placeholder.Name,
			Path:     toParseFile.PathRel,
			Err:      parseErr.Error(),
		}, nil
	}

	for _, assoc := range associations {
		ctx.allClassAssociations[assoc.Key] = assoc
	}
	addClassToSubdomain(model, domainKey, subdomainKey, domain, subdomain, class)
	return nil, nil //nolint:nilnil // success: no resilient-parse failure to report
}

// addClassToSubdomain inserts a class into its subdomain and writes it back to the model.
func addClassToSubdomain(model *core.Model, domainKey, subdomainKey identity.Key, domain model_domain.Domain, subdomain model_domain.Subdomain, class model_class.Class) {
	if subdomain.Classes == nil {
		subdomain.Classes = make(map[identity.Key]model_class.Class)
	}
	subdomain.Classes[class.Key] = class
	updateSubdomain(model, domainKey, subdomainKey, domain, subdomain)
}

// placeholderClass builds a minimal valid empty class for a .class file that
// failed to parse. The key matches what a successful parse would have produced.
func placeholderClass(subdomainKey identity.Key, classSubKey string) (model_class.Class, error) {
	classKey, err := identity.NewClassKey(subdomainKey, classSubKey)
	if err != nil {
		return model_class.Class{}, errors.WithStack(err)
	}
	return model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: classSubKey, Details: "", UnfinishedNotes: "", UmlComment: ""}), nil
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
