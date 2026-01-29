package parser_ai

import (
	"os"
	"path/filepath"
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

	// Initialize child maps
	model.Actors = make(map[string]*inputActor)
	model.Domains = make(map[string]*inputDomain)
	model.Associations = make(map[string]*inputAssociation)

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
			content, err := os.ReadFile(filepath.Join(actorsDir, name))
			if err != nil {
				return nil, err
			}
			actor, err := parseActor(content, filepath.Join(actorsDir, name))
			if err != nil {
				return nil, err
			}
			model.Actors[key] = actor
		}
	}

	// Read model-level associations
	assocDir := filepath.Join(modelDir, "associations")
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
			content, err := os.ReadFile(filepath.Join(assocDir, name))
			if err != nil {
				return nil, err
			}
			assoc, err := parseAssociation(content, filepath.Join(assocDir, name))
			if err != nil {
				return nil, err
			}
			model.Associations[key] = assoc
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
			domain, err := readDomainTree(filepath.Join(domainsDir, domainKey))
			if err != nil {
				return nil, err
			}
			model.Domains[domainKey] = domain
		}
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
	domain.Associations = make(map[string]*inputAssociation)

	// Read domain-level associations
	assocDir := filepath.Join(domainDir, "associations")
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
			content, err := os.ReadFile(filepath.Join(assocDir, name))
			if err != nil {
				return nil, err
			}
			assoc, err := parseAssociation(content, filepath.Join(assocDir, name))
			if err != nil {
				return nil, err
			}
			domain.Associations[key] = assoc
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
			subdomain, err := readSubdomainTree(filepath.Join(subdomainsDir, subdomainKey))
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
	subdomain.Generalizations = make(map[string]*inputGeneralization)
	subdomain.Associations = make(map[string]*inputAssociation)

	// Read subdomain-level associations
	assocDir := filepath.Join(subdomainDir, "associations")
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
			content, err := os.ReadFile(filepath.Join(assocDir, name))
			if err != nil {
				return nil, err
			}
			assoc, err := parseAssociation(content, filepath.Join(assocDir, name))
			if err != nil {
				return nil, err
			}
			subdomain.Associations[key] = assoc
		}
	}

	// Read generalizations
	genDir := filepath.Join(subdomainDir, "generalizations")
	if entries, err := os.ReadDir(genDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".gen.json") {
				continue
			}
			key := strings.TrimSuffix(name, ".gen.json")
			content, err := os.ReadFile(filepath.Join(genDir, name))
			if err != nil {
				return nil, err
			}
			gen, err := parseGeneralization(content, filepath.Join(genDir, name))
			if err != nil {
				return nil, err
			}
			subdomain.Generalizations[key] = gen
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
			class, err := readClassTree(filepath.Join(classesDir, classKey))
			if err != nil {
				return nil, err
			}
			subdomain.Classes[classKey] = class
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
			content, err := os.ReadFile(filepath.Join(actionsDir, name))
			if err != nil {
				return nil, err
			}
			action, err := parseAction(content, filepath.Join(actionsDir, name))
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
			content, err := os.ReadFile(filepath.Join(queriesDir, name))
			if err != nil {
				return nil, err
			}
			query, err := parseQuery(content, filepath.Join(queriesDir, name))
			if err != nil {
				return nil, err
			}
			class.Queries[key] = query
		}
	}

	return class, nil
}
