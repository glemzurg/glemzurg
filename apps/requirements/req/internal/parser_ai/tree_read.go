package parser_ai

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
)

func ReadModel(inputModelPath string) (core.Model, error) {
	model, err := readModel(inputModelPath)
	if err != nil {
		// Check if any error in the chain is a ParseError (user-fixable).
		// For joined errors, errors.As finds the first match.
		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			return core.Model{}, fmt.Errorf("STOP AND REPORT THIS ERROR to the user. This is an unexpected internal error that cannot be fixed by changing input files: %w", err)
		}
		return core.Model{}, err
	}
	return model, nil
}

func readModel(inputModelPath string) (core.Model, error) {
	modelKey := filepath.Base(inputModelPath)

	inputModel, err := readModelTree(inputModelPath)
	if err != nil {
		// Tree reading errors are accumulated — return them all.
		return core.Model{}, err
	}

	modelPtr, err := ConvertToModel(inputModel, modelKey)
	if err != nil {
		return core.Model{}, err
	}

	// Lower all expressions with full context.
	if err := convert.LowerAllExpressions(modelPtr); err != nil {
		return core.Model{}, err
	}

	return *modelPtr, nil
}

// readModelTree reads a complete model tree from the filesystem.
// The modelDir is the root directory where the model is stored.
func readModelTree(modelDir string) (*inputModel, error) {
	// Read model.json
	modelContent, err := os.ReadFile(filepath.Join(modelDir, "model.json"))
	if err != nil {
		return nil, err
	}
	model, err := parseModel(modelContent, filepath.Join(modelDir, "model.json"))
	if err != nil {
		return nil, err
	}

	initModelMaps(model)

	var errs []error
	if err := readModelChildren(modelDir, model); err != nil {
		errs = append(errs, err)
	}

	// Only run validation if reading succeeded — validation depends on populated model.
	if len(errs) == 0 {
		if err := validateModelCompleteness(model); err != nil {
			errs = append(errs, err)
		}
		if err := validateModelTree(model); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return model, nil
}

// initModelMaps initializes all child maps on a freshly parsed model.
func initModelMaps(model *inputModel) {
	model.Actors = make(map[string]*inputActor)
	model.ActorGeneralizations = make(map[string]*inputActorGeneralization)
	model.GlobalFunctions = make(map[string]*inputGlobalFunction)
	model.NamedSets = make(map[string]*inputNamedSet)
	model.Domains = make(map[string]*inputDomain)
	model.DomainAssociations = make(map[string]*inputDomainAssociation)
	model.ClassAssociations = make(map[string]*inputClassAssociation)
}

// readModelChildren reads all child entities from the filesystem.
// Collects errors from all child types so the caller sees all problems at once.
func readModelChildren(modelDir string, model *inputModel) error {
	var errs []error
	if err := readModelInvariants(modelDir, model); err != nil {
		errs = append(errs, err)
	}
	if err := readModelActors(modelDir, model); err != nil {
		errs = append(errs, err)
	}
	if err := readModelActorGeneralizations(modelDir, model); err != nil {
		errs = append(errs, err)
	}
	if err := readModelGlobalFunctions(modelDir, model); err != nil {
		errs = append(errs, err)
	}
	if err := readModelNamedSets(modelDir, model); err != nil {
		errs = append(errs, err)
	}
	if err := readModelClassAssociations(modelDir, model); err != nil {
		errs = append(errs, err)
	}
	if err := readModelDomainAssociations(modelDir, model); err != nil {
		errs = append(errs, err)
	}
	if err := readModelDomains(modelDir, model); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// readModelInvariants reads model-level invariants from the filesystem.
func readModelInvariants(modelDir string, model *inputModel) error {
	return readInvariantsDir(filepath.Join(modelDir, "invariants"), &model.Invariants)
}

// readModelActors reads actor files from the filesystem.
// Collects errors from all actor files so the caller sees all problems at once.
func readModelActors(modelDir string, model *inputModel) error {
	actorsDir := filepath.Join(modelDir, "actors")
	entries, err := os.ReadDir(actorsDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".actor.json") {
			continue
		}
		key := strings.TrimSuffix(name, ".actor.json")
		filePath := filepath.Join(actorsDir, name)
		if err := ValidateKey(key, "actor_key", filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		actor, err := parseActor(content, filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		model.Actors[key] = actor
	}
	return errors.Join(errs...)
}

// readModelActorGeneralizations reads actor generalization files from the filesystem.
func readModelActorGeneralizations(modelDir string, model *inputModel) error {
	agDir := filepath.Join(modelDir, "actor_generalizations")
	entries, err := os.ReadDir(agDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".agen.json") {
			continue
		}
		key := strings.TrimSuffix(name, ".agen.json")
		filePath := filepath.Join(agDir, name)
		if err := ValidateKey(key, "actor_generalization_key", filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		gen, err := parseActorGeneralization(content, filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		model.ActorGeneralizations[key] = gen
	}
	return errors.Join(errs...)
}

// readModelGlobalFunctions reads global function files from the filesystem.
func readModelGlobalFunctions(modelDir string, model *inputModel) error {
	gfDir := filepath.Join(modelDir, "global_functions")
	entries, err := os.ReadDir(gfDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		key := strings.TrimSuffix(name, ".json")
		filePath := filepath.Join(gfDir, name)
		if err := ValidateKey(key, "global_function_key", filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		gf, err := parseGlobalFunction(content, filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		model.GlobalFunctions[key] = gf
	}
	return errors.Join(errs...)
}

// readModelNamedSets reads named set files from the filesystem.
func readModelNamedSets(modelDir string, model *inputModel) error {
	nsDir := filepath.Join(modelDir, "named_sets")
	entries, err := os.ReadDir(nsDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".nset.json") {
			continue
		}
		key := strings.TrimSuffix(name, ".nset.json")
		filePath := filepath.Join(nsDir, name)
		if err := ValidateKey(key, "named_set_key", filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ns, err := parseNamedSet(content, filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		model.NamedSets[key] = ns
	}
	return errors.Join(errs...)
}

// readModelClassAssociations reads model-level class association files from the filesystem.
func readModelClassAssociations(modelDir string, model *inputModel) error {
	assocDir := filepath.Join(modelDir, "class_associations")
	entries, err := os.ReadDir(assocDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".assoc.json") {
			continue
		}
		key := strings.TrimSuffix(name, ".assoc.json")
		filePath := filepath.Join(assocDir, name)
		if err := ValidateAssociationFilename(key, AssocLevelModel, filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		assoc, err := parseAssociation(content, filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		model.ClassAssociations[key] = assoc
	}
	return errors.Join(errs...)
}

// readModelDomainAssociations reads domain association files from the filesystem.
func readModelDomainAssociations(modelDir string, model *inputModel) error {
	daDir := filepath.Join(modelDir, "domain_associations")
	entries, err := os.ReadDir(daDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".domain_assoc.json") {
			continue
		}
		key := strings.TrimSuffix(name, ".domain_assoc.json")
		filePath := filepath.Join(daDir, name)
		if err := ValidateKey(key, "domain_association_key", filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		da, err := parseDomainAssociation(content, filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		model.DomainAssociations[key] = da
	}
	return errors.Join(errs...)
}

// readModelDomains reads domain directories from the filesystem.
func readModelDomains(modelDir string, model *inputModel) error {
	domainsDir := filepath.Join(modelDir, "domains")
	entries, err := os.ReadDir(domainsDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		domainKey := entry.Name()
		domainDir := filepath.Join(domainsDir, domainKey)
		if err := ValidateKey(domainKey, "domain_key", filepath.Join(domainDir, "domain.json")); err != nil {
			errs = append(errs, err)
			continue
		}
		domain, err := readDomainTree(domainDir)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		model.Domains[domainKey] = domain
	}
	return errors.Join(errs...)
}

// readDomainTree reads a domain and its children from the filesystem.
func readDomainTree(domainDir string) (*inputDomain, error) {
	// Read domain.json
	domainContent, err := os.ReadFile(filepath.Join(domainDir, "domain.json"))
	if err != nil {
		return nil, err
	}
	domain, err := parseDomain(domainContent, filepath.Join(domainDir, "domain.json"))
	if err != nil {
		return nil, err
	}

	// Initialize child maps
	domain.Subdomains = make(map[string]*inputSubdomain)
	domain.ClassAssociations = make(map[string]*inputClassAssociation)

	// Read domain-level class associations
	var errs []error
	assocDir := filepath.Join(domainDir, "class_associations")
	if entries, err := os.ReadDir(assocDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".assoc.json") {
				continue
			}
			key := strings.TrimSuffix(name, ".assoc.json")
			filePath := filepath.Join(assocDir, name)

			// Validate association filename format (domain level: subdomain.class--subdomain.class--name)
			if err := ValidateAssociationFilename(key, AssocLevelDomain, filePath); err != nil {
				errs = append(errs, err)
				continue
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			assoc, err := parseAssociation(content, filePath)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			domain.ClassAssociations[key] = assoc
		}
	}

	// Read subdomains
	subdomainsDir := filepath.Join(domainDir, "subdomains")
	if entries, err := os.ReadDir(subdomainsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			subdomainKey := entry.Name()
			subdomainDir := filepath.Join(subdomainsDir, subdomainKey)

			// Validate key format
			if err := ValidateKey(subdomainKey, "subdomain_key", filepath.Join(subdomainDir, "subdomain.json")); err != nil {
				errs = append(errs, err)
				continue
			}

			subdomain, err := readSubdomainTree(subdomainDir)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			domain.Subdomains[subdomainKey] = subdomain
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return domain, nil
}

// readSubdomainTree reads a subdomain and its children from the filesystem.
func readSubdomainTree(subdomainDir string) (*inputSubdomain, error) {
	// Read subdomain.json
	subdomainContent, err := os.ReadFile(filepath.Join(subdomainDir, "subdomain.json"))
	if err != nil {
		return nil, err
	}
	subdomain, err := parseSubdomain(subdomainContent, filepath.Join(subdomainDir, "subdomain.json"))
	if err != nil {
		return nil, err
	}

	// Initialize child maps
	subdomain.Classes = make(map[string]*inputClass)
	subdomain.ClassGeneralizations = make(map[string]*inputClassGeneralization)
	subdomain.ClassAssociations = make(map[string]*inputClassAssociation)
	subdomain.UseCases = make(map[string]*inputUseCase)
	subdomain.UseCaseGeneralizations = make(map[string]*inputUseCaseGeneralization)
	subdomain.UseCaseShares = make(map[string]map[string]*inputUseCaseShared)

	var errs []error
	if err := readSubdomainAssociations(subdomainDir, subdomain); err != nil {
		errs = append(errs, err)
	}
	if err := readSubdomainGeneralizations(subdomainDir, subdomain); err != nil {
		errs = append(errs, err)
	}
	if err := readSubdomainUseCaseGeneralizations(subdomainDir, subdomain); err != nil {
		errs = append(errs, err)
	}
	if err := readSubdomainClasses(subdomainDir, subdomain); err != nil {
		errs = append(errs, err)
	}
	if err := readSubdomainUseCases(subdomainDir, subdomain); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return subdomain, nil
}

// readSubdomainAssociations reads subdomain-level class associations.
func readSubdomainAssociations(subdomainDir string, subdomain *inputSubdomain) error {
	assocDir := filepath.Join(subdomainDir, "class_associations")
	entries, err := os.ReadDir(assocDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".assoc.json") {
			continue
		}
		key := strings.TrimSuffix(name, ".assoc.json")
		filePath := filepath.Join(assocDir, name)
		if err := ValidateAssociationFilename(key, AssocLevelSubdomain, filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		assoc, err := parseAssociation(content, filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		subdomain.ClassAssociations[key] = assoc
	}
	return errors.Join(errs...)
}

// readSubdomainGeneralizations reads class generalization files.
func readSubdomainGeneralizations(subdomainDir string, subdomain *inputSubdomain) error {
	genDir := filepath.Join(subdomainDir, "class_generalizations")
	entries, err := os.ReadDir(genDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".cgen.json") {
			continue
		}
		key := strings.TrimSuffix(name, ".cgen.json")
		filePath := filepath.Join(genDir, name)
		if err := ValidateKey(key, "generalization_key", filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		gen, err := parseClassGeneralization(content, filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		subdomain.ClassGeneralizations[key] = gen
	}
	return errors.Join(errs...)
}

// readSubdomainUseCaseGeneralizations reads use case generalization files.
func readSubdomainUseCaseGeneralizations(subdomainDir string, subdomain *inputSubdomain) error {
	ucgDir := filepath.Join(subdomainDir, "use_case_generalizations")
	entries, err := os.ReadDir(ucgDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".ucgen.json") {
			continue
		}
		key := strings.TrimSuffix(name, ".ucgen.json")
		filePath := filepath.Join(ucgDir, name)
		if err := ValidateKey(key, "use_case_generalization_key", filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		gen, err := parseUseCaseGeneralization(content, filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		subdomain.UseCaseGeneralizations[key] = gen
	}
	return errors.Join(errs...)
}

// readSubdomainClasses reads class directories from the filesystem.
func readSubdomainClasses(subdomainDir string, subdomain *inputSubdomain) error {
	classesDir := filepath.Join(subdomainDir, "classes")
	entries, err := os.ReadDir(classesDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		classKey := entry.Name()
		classDir := filepath.Join(classesDir, classKey)
		if err := ValidateKey(classKey, "class_key", filepath.Join(classDir, "class.json")); err != nil {
			errs = append(errs, err)
			continue
		}
		class, err := readClassTree(classDir)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		subdomain.Classes[classKey] = class
	}
	return errors.Join(errs...)
}

// readSubdomainUseCases reads use case directories from the filesystem.
func readSubdomainUseCases(subdomainDir string, subdomain *inputSubdomain) error {
	useCasesDir := filepath.Join(subdomainDir, "use_cases")
	entries, err := os.ReadDir(useCasesDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		useCaseKey := entry.Name()
		useCaseDir := filepath.Join(useCasesDir, useCaseKey)
		if err := ValidateKey(useCaseKey, "use_case_key", filepath.Join(useCaseDir, "use_case.json")); err != nil {
			errs = append(errs, err)
			continue
		}
		useCase, err := readUseCaseTree(useCaseDir)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		subdomain.UseCases[useCaseKey] = useCase
	}
	return errors.Join(errs...)
}

// readClassTree reads a class and its children from the filesystem.
func readClassTree(classDir string) (*inputClass, error) {
	// Read class.json
	classContent, err := os.ReadFile(filepath.Join(classDir, "class.json"))
	if err != nil {
		return nil, err
	}
	class, err := parseClass(classContent, filepath.Join(classDir, "class.json"))
	if err != nil {
		return nil, err
	}

	// Initialize child maps
	class.Actions = make(map[string]*inputAction)
	class.Queries = make(map[string]*inputQuery)

	var errs []error
	if err := readClassInvariants(classDir, class); err != nil {
		errs = append(errs, err)
	}
	if err := readClassAttributeInvariants(classDir, class); err != nil {
		errs = append(errs, err)
	}

	// Read state_machine.json if present
	smPath := filepath.Join(classDir, "state_machine.json")
	if smContent, err := os.ReadFile(smPath); err == nil {
		sm, err := parseStateMachine(smContent, smPath)
		if err != nil {
			errs = append(errs, err)
		} else {
			class.StateMachine = sm
		}
	}

	if err := readClassActions(classDir, class); err != nil {
		errs = append(errs, err)
	}
	if err := readClassQueries(classDir, class); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return class, nil
}

// readClassInvariants reads class-level invariant files.
func readClassInvariants(classDir string, class *inputClass) error {
	return readInvariantsDir(filepath.Join(classDir, "invariants"), &class.Invariants)
}

// readInvariantsDir reads invariant JSON files from a directory into the target slice.
func readInvariantsDir(dir string, target *[]inputLogic) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".invariant.json") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	var errs []error
	for _, name := range names {
		filePath := filepath.Join(dir, name)
		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		logic, err := parseLogic(content, filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		*target = append(*target, *logic)
	}
	return errors.Join(errs...)
}

// readClassAttributeInvariants reads per-attribute invariant files.
func readClassAttributeInvariants(classDir string, class *inputClass) error {
	for attrKey, attr := range class.Attributes {
		attrInvariantsDir := filepath.Join(classDir, "attributes", attrKey, "invariants")
		if err := readInvariantsDir(attrInvariantsDir, &attr.Invariants); err != nil {
			return err
		}
		class.Attributes[attrKey] = attr
	}
	return nil
}

// readClassActions reads action JSON files from the class directory.
func readClassActions(classDir string, class *inputClass) error {
	actionsDir := filepath.Join(classDir, "actions")
	entries, err := os.ReadDir(actionsDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		key := strings.TrimSuffix(name, ".json")
		filePath := filepath.Join(actionsDir, name)
		if err := ValidateKey(key, "action_key", filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		action, err := parseAction(content, filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if err := validateFilenameMatchesName(key, action.Name, "action", ErrActionFilenameInvalid, filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		class.Actions[key] = action
	}
	return errors.Join(errs...)
}

// readClassQueries reads query JSON files from the class directory.
func readClassQueries(classDir string, class *inputClass) error {
	queriesDir := filepath.Join(classDir, "queries")
	entries, err := os.ReadDir(queriesDir)
	if err != nil {
		return nil
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		key := strings.TrimSuffix(name, ".json")
		filePath := filepath.Join(queriesDir, name)
		if err := ValidateKey(key, "query_key", filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		query, err := parseQuery(content, filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if err := validateFilenameMatchesName(key, query.Name, "query", ErrQueryFilenameInvalid, filePath); err != nil {
			errs = append(errs, err)
			continue
		}
		class.Queries[key] = query
	}
	return errors.Join(errs...)
}

// readUseCaseTree reads a use case and its children from the filesystem.
func readUseCaseTree(useCaseDir string) (*inputUseCase, error) {
	// Read use_case.json
	useCaseContent, err := os.ReadFile(filepath.Join(useCaseDir, "use_case.json"))
	if err != nil {
		return nil, err
	}
	useCase, err := parseUseCase(useCaseContent, filepath.Join(useCaseDir, "use_case.json"))
	if err != nil {
		return nil, err
	}

	// Initialize child maps
	useCase.Scenarios = make(map[string]*inputScenario)

	// Read scenarios
	var errs []error
	scenariosDir := filepath.Join(useCaseDir, "scenarios")
	if entries, err := os.ReadDir(scenariosDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".scenario.json") {
				continue
			}
			key := strings.TrimSuffix(name, ".scenario.json")
			filePath := filepath.Join(scenariosDir, name)

			// Validate key format
			if err := ValidateKey(key, "scenario_key", filePath); err != nil {
				errs = append(errs, err)
				continue
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			scenario, err := parseScenario(content, filePath)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			useCase.Scenarios[key] = scenario
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return useCase, nil
}
