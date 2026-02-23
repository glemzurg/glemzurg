package parser_ai

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ReadModelTree reads a complete model tree from the filesystem.
// The modelDir is the root directory where the model is stored.
func ReadModelTree(modelDir string) (*inputModel, error) {
	// Read model.json
	modelContent, err := os.ReadFile(filepath.Join(modelDir, "model.json"))
	if err != nil {
		return nil, err
	}
	model, err := parseModel(modelContent, filepath.Join(modelDir, "model.json"))
	if err != nil {
		return nil, err
	}

	// Initialize child maps and slices
	model.Actors = make(map[string]*inputActor)
	model.ActorGeneralizations = make(map[string]*inputActorGeneralization)
	model.GlobalFunctions = make(map[string]*inputGlobalFunction)
	model.Domains = make(map[string]*inputDomain)
	model.DomainAssociations = make(map[string]*inputDomainAssociation)
	model.ClassAssociations = make(map[string]*inputClassAssociation)

	// Read invariants
	invariantsDir := filepath.Join(modelDir, "invariants")
	if entries, err := os.ReadDir(invariantsDir); err == nil {
		// Sort entries to preserve order (001, 002, ...)
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
		for _, name := range names {
			filePath := filepath.Join(invariantsDir, name)
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			logic, err := parseLogic(content, filePath)
			if err != nil {
				return nil, err
			}
			model.Invariants = append(model.Invariants, *logic)
		}
	}

	// Read actors
	actorsDir := filepath.Join(modelDir, "actors")
	if entries, err := os.ReadDir(actorsDir); err == nil {
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
				return nil, err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			actor, err := parseActor(content, filePath)
			if err != nil {
				return nil, err
			}
			model.Actors[key] = actor
		}
	}

	// Read actor generalizations
	agDir := filepath.Join(modelDir, "actor_generalizations")
	if entries, err := os.ReadDir(agDir); err == nil {
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
				return nil, err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			gen, err := parseActorGeneralization(content, filePath)
			if err != nil {
				return nil, err
			}
			model.ActorGeneralizations[key] = gen
		}
	}

	// Read global functions
	gfDir := filepath.Join(modelDir, "global_functions")
	if entries, err := os.ReadDir(gfDir); err == nil {
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
				return nil, err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			gf, err := parseGlobalFunction(content, filePath)
			if err != nil {
				return nil, err
			}
			model.GlobalFunctions[key] = gf
		}
	}

	// Read model-level class associations
	assocDir := filepath.Join(modelDir, "class_associations")
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

			// Validate association filename format (model level: domain.subdomain.class--domain.subdomain.class--name)
			if err := ValidateAssociationFilename(key, AssocLevelModel, filePath); err != nil {
				return nil, err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			assoc, err := parseAssociation(content, filePath)
			if err != nil {
				return nil, err
			}
			model.ClassAssociations[key] = assoc
		}
	}

	// Read domain associations
	daDir := filepath.Join(modelDir, "domain_associations")
	if entries, err := os.ReadDir(daDir); err == nil {
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
				return nil, err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			da, err := parseDomainAssociation(content, filePath)
			if err != nil {
				return nil, err
			}
			model.DomainAssociations[key] = da
		}
	}

	// Read domains
	domainsDir := filepath.Join(modelDir, "domains")
	if entries, err := os.ReadDir(domainsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			domainKey := entry.Name()
			domainDir := filepath.Join(domainsDir, domainKey)

			// Validate key format
			if err := ValidateKey(domainKey, "domain_key", filepath.Join(domainDir, "domain.json")); err != nil {
				return nil, err
			}

			domain, err := readDomainTree(domainDir)
			if err != nil {
				return nil, err
			}
			model.Domains[domainKey] = domain
		}
	}

	// Validate model completeness
	if err := validateModelCompleteness(model); err != nil {
		return nil, err
	}

	// Validate cross-references in the tree
	if err := validateModelTree(model); err != nil {
		return nil, err
	}

	return model, nil
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
				return nil, err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			assoc, err := parseAssociation(content, filePath)
			if err != nil {
				return nil, err
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
				return nil, err
			}

			subdomain, err := readSubdomainTree(subdomainDir)
			if err != nil {
				return nil, err
			}
			domain.Subdomains[subdomainKey] = subdomain
		}
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

	// Read subdomain-level class associations
	assocDir := filepath.Join(subdomainDir, "class_associations")
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

			// Validate association filename format (subdomain level: class--class--name)
			if err := ValidateAssociationFilename(key, AssocLevelSubdomain, filePath); err != nil {
				return nil, err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			assoc, err := parseAssociation(content, filePath)
			if err != nil {
				return nil, err
			}
			subdomain.ClassAssociations[key] = assoc
		}
	}

	// Read generalizations
	genDir := filepath.Join(subdomainDir, "class_generalizations")
	if entries, err := os.ReadDir(genDir); err == nil {
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

			// Validate key format
			if err := ValidateKey(key, "generalization_key", filePath); err != nil {
				return nil, err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			gen, err := parseClassGeneralization(content, filePath)
			if err != nil {
				return nil, err
			}
			subdomain.ClassGeneralizations[key] = gen
		}
	}

	// Read use case generalizations
	ucgDir := filepath.Join(subdomainDir, "use_case_generalizations")
	if entries, err := os.ReadDir(ucgDir); err == nil {
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

			// Validate key format
			if err := ValidateKey(key, "use_case_generalization_key", filePath); err != nil {
				return nil, err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			gen, err := parseUseCaseGeneralization(content, filePath)
			if err != nil {
				return nil, err
			}
			subdomain.UseCaseGeneralizations[key] = gen
		}
	}

	// Read classes
	classesDir := filepath.Join(subdomainDir, "classes")
	if entries, err := os.ReadDir(classesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			classKey := entry.Name()
			classDir := filepath.Join(classesDir, classKey)

			// Validate key format
			if err := ValidateKey(classKey, "class_key", filepath.Join(classDir, "class.json")); err != nil {
				return nil, err
			}

			class, err := readClassTree(classDir)
			if err != nil {
				return nil, err
			}
			subdomain.Classes[classKey] = class
		}
	}

	// Read use cases
	useCasesDir := filepath.Join(subdomainDir, "use_cases")
	if entries, err := os.ReadDir(useCasesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			useCaseKey := entry.Name()
			useCaseDir := filepath.Join(useCasesDir, useCaseKey)

			// Validate key format
			if err := ValidateKey(useCaseKey, "use_case_key", filepath.Join(useCaseDir, "use_case.json")); err != nil {
				return nil, err
			}

			useCase, err := readUseCaseTree(useCaseDir)
			if err != nil {
				return nil, err
			}
			subdomain.UseCases[useCaseKey] = useCase
		}
	}

	return subdomain, nil
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

	// Read state_machine.json if present
	smPath := filepath.Join(classDir, "state_machine.json")
	if smContent, err := os.ReadFile(smPath); err == nil {
		sm, err := parseStateMachine(smContent, smPath)
		if err != nil {
			return nil, err
		}
		class.StateMachine = sm
	}

	// Read actions
	actionsDir := filepath.Join(classDir, "actions")
	if entries, err := os.ReadDir(actionsDir); err == nil {
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

			// Validate key format
			if err := ValidateKey(key, "action_key", filePath); err != nil {
				return nil, err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			action, err := parseAction(content, filePath)
			if err != nil {
				return nil, err
			}
			class.Actions[key] = action
		}
	}

	// Read queries
	queriesDir := filepath.Join(classDir, "queries")
	if entries, err := os.ReadDir(queriesDir); err == nil {
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

			// Validate key format
			if err := ValidateKey(key, "query_key", filePath); err != nil {
				return nil, err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			query, err := parseQuery(content, filePath)
			if err != nil {
				return nil, err
			}
			class.Queries[key] = query
		}
	}

	return class, nil
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
				return nil, err
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			scenario, err := parseScenario(content, filePath)
			if err != nil {
				return nil, err
			}
			useCase.Scenarios[key] = scenario
		}
	}

	return useCase, nil
}
