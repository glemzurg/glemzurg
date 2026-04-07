package generate

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"text/template"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// UseCaseShare pairs a share type with the related use case for template rendering.
type UseCaseShare struct {
	ShareType string
	UseCase   model_use_case.UseCase
}

//go:embed templates/*
var _templateFS embed.FS

// _templateRegistry maps template filenames to pointers where the parsed templates should be stored.
var _templateRegistry = map[string]**template.Template{
	"model.md.template":            &_modelMdTemplate,
	"actor.md.template":            &_actorMdTemplate,
	"domain.md.template":           &_domainMdTemplate,
	"domains.mermaid.template":     &_domainsMermaidTemplate,
	"use_cases.mermaid.template":   &_useCasesMermaidTemplate,
	"classes.mermaid.template":     &_classesMermaidTemplate,
	"class.md.template":            &_classMdTemplate,
	"class-state.mermaid.template": &_classStateMermaidTemplate,
	"use_case.md.template":         &_useCaseMdTemplate,
	"subdomain.md.template":        &_subdomainMdTemplate,
	"subdomains.mermaid.template":  &_subdomainsMermaidTemplate,
}

func init() {
	// Walk through the embedded file system to find and parse all .template files.
	err := fs.WalkDir(_templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		// Report any error walking into this path.
		if err != nil {
			return errors.WithStack(err)
		}

		// Ignore directories as data.
		if d.IsDir() {
			return nil
		}

		// Skip non-template files.
		if filepath.Ext(path) != ".template" {
			return nil
		}

		return parseAndRegisterTemplate(path)
	})
	if err != nil {
		log.Fatalf("Failed to parse templates: %+v", err)
	}
}

// parseAndRegisterTemplate reads, parses, and registers a single template file.
func parseAndRegisterTemplate(path string) error {
	// Read from the embedded file system.
	content, err := _templateFS.ReadFile(path)
	if err != nil {
		return errors.WithStack(err)
	}

	// Parse the template and add it to the set.
	tmplName := filepath.Base(path)
	tmpl, err := template.New(tmplName).Funcs(_funcMap).Parse(string(content))
	if err != nil {
		return errors.WithStack(err)
	}

	log.Printf("Parsed template: %s", tmplName)

	// Register the template using the registry.
	target, found := _templateRegistry[tmplName]
	if !found {
		return errors.WithStack(errors.Errorf(`unknown template filename: '%s'`, tmplName))
	}
	*target = tmpl

	return nil
}

// The templates in the system.
var _modelMdTemplate *template.Template
var _actorMdTemplate *template.Template
var _domainMdTemplate *template.Template
var _domainsMermaidTemplate *template.Template
var _useCasesMermaidTemplate *template.Template
var _classesMermaidTemplate *template.Template
var _classMdTemplate *template.Template
var _classStateMermaidTemplate *template.Template
var _useCaseMdTemplate *template.Template
var _subdomainMdTemplate *template.Template
var _subdomainsMermaidTemplate *template.Template

// Define some function for our templates.
var _funcMap = template.FuncMap{
	"nodeid": func(idtype string, key identity.Key) string {
		keyStr := key.String()
		// Replace characters that are invalid in Mermaid node IDs.
		keyStr = strings.ReplaceAll(keyStr, "/", "_")
		keyStr = strings.ReplaceAll(keyStr, "-", "_")
		keyStr = strings.ReplaceAll(keyStr, ".", "_")
		return idtype + "_" + keyStr
	},

	"lookup": func(lookup map[string]string, key identity.Key) (value string) {
		keyStr := key.String()
		value, found := lookup[keyStr]
		if !found {
			panic(fmt.Sprintf("Unknown lookup key: '%s'", keyStr))
		}
		return value
	},
	"filename": func(objType string, key identity.Key, suffix, ext string) (filename string) {
		return convertKeyToFilename(objType, key.String(), suffix, ext)
	},
	"data_type_rules": func(rules string, dataType *model_data_type.DataType) (value string) {
		if dataType == nil {
			return `_(unparsed)_ ` + rules
		}
		return "__" + dataType.String() + "__"
	},
	"first_md_paragraph": firstMdParagraph,
	"first_md_sentence": func(md string) (paragraph string) {
		return firstSentence(firstMdParagraph(md))
	},
	"multiplicity": func(multiplicity model_class.Multiplicity) (value string) {
		return multiplicity.String()
	},
	"generalization_label": func(reqs *req_flat.Requirements, generalizationKey identity.Key) (value string) {
		generalizationLookup := reqs.GeneralizationLookup()
		generalization := generalizationLookup[generalizationKey.String()]
		complete := "«complete»"
		if !generalization.IsComplete {
			complete = "«incomplete»"
		}
		static := "«static»"
		if !generalization.IsStatic {
			static = "«dynamic»"
		}
		return complete + "\n" + static
	},
	"event_guard_signature": func(reqs *req_flat.Requirements, transition model_state.Transition) (eventCall string) {

		eventLookup := reqs.EventLookup()
		guardLookup := reqs.GuardLookup()

		// The event.
		event := eventLookup[transition.EventKey.String()]

		// Create a signature for the event.
		var paramNames []string
		for _, param := range event.Parameters {
			paramNames = append(paramNames, param.Name)
		}
		signature := strings.Join(paramNames, ", ")

		// The main call.
		eventCall = event.Name + "(" + signature + ")"

		// Add a guard if there is one.
		if transition.GuardKey != nil {
			guard := guardLookup[transition.GuardKey.String()]
			eventCall += " [" + guard.Logic.Description + "]"
		}

		return eventCall
	},
	"action_signature": func(action model_state.Action) (signature string) {
		var paramNames []string
		for _, param := range action.Parameters {
			paramNames = append(paramNames, param.Name)
		}
		return strings.Join(paramNames, ", ")
	},
	"query_signature": func(query model_state.Query) (signature string) {
		var paramNames []string
		for _, param := range query.Parameters {
			paramNames = append(paramNames, param.Name)
		}
		return strings.Join(paramNames, ", ")
	},

	// Formatting of bulleted lists.
	"main_bullet": func(bulletText string) (mainBullet string) {
		mainBullet, _ = splitBulletTextIntoMainAndSubBullets(bulletText)
		return mainBullet
	},
	"sub_bullets": func(bulletText string) (subBullets []string) {
		_, subBullets = splitBulletTextIntoMainAndSubBullets(bulletText)
		return subBullets
	},

	// Lookup methods for objects.
	"domain_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_domain.Domain) {
		lookup, _ := reqs.DomainLookup()
		return lookup[key.String()]
	},
	"class_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_class.Class) {
		lookup, _ := reqs.ClassLookup()
		return lookup[key.String()]
	},
	"state_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_state.State) {
		lookup := reqs.StateLookup()
		return lookup[key.String()]
	},
	"event_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_state.Event) {
		lookup := reqs.EventLookup()
		return lookup[key.String()]
	},
	"guard_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_state.Guard) {
		lookup := reqs.GuardLookup()
		return lookup[key.String()]
	},
	"action_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_state.Action) {
		lookup := reqs.ActionLookup()
		return lookup[key.String()]
	},
	"use_case_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_use_case.UseCase) {
		lookup := reqs.UseCaseLookup()
		return lookup[key.String()]
	},
	"scenario_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_scenario.Scenario) {
		lookup := reqs.ScenarioLookup()
		return lookup[key.String()]
	},
	"actor_lookup": func(reqs *req_flat.Requirements, key identity.Key) (actor model_actor.Actor) {
		lookup := reqs.ActorLookup()
		return lookup[key.String()]
	},
	"actor_classes": func(reqs *req_flat.Requirements, key identity.Key) (classes []model_class.Class) {
		lookup := reqs.ActorClassesLookup()
		return lookup[key.String()]
	},
	"class_domain": func(reqs *req_flat.Requirements, key identity.Key) (domain model_domain.Domain) {
		lookup := reqs.ClassDomainLookup()
		return lookup[key.String()]
	},
	"use_case_domain": func(reqs *req_flat.Requirements, key identity.Key) (domain model_domain.Domain) {
		lookup := reqs.UseCaseDomainLookup()
		return lookup[key.String()]
	},
	"class_subdomain": func(reqs *req_flat.Requirements, key identity.Key) (subdomain model_domain.Subdomain) {
		lookup := reqs.ClassSubdomainLookup()
		return lookup[key.String()]
	},
	"use_case_subdomain": func(reqs *req_flat.Requirements, key identity.Key) (subdomain model_domain.Subdomain) {
		lookup := reqs.UseCaseSubdomainLookup()
		return lookup[key.String()]
	},
	"domain_has_multiple_subdomains": func(reqs *req_flat.Requirements, domainKey identity.Key) bool {
		return reqs.DomainHasMultipleSubdomains(domainKey)
	},
	"domain_use_cases": func(reqs *req_flat.Requirements, key identity.Key) (useCases []model_use_case.UseCase) {
		lookup := reqs.DomainUseCasesLookup()
		return lookup[key.String()]
	},
	"domain_classes": func(reqs *req_flat.Requirements, key identity.Key) (classes []model_class.Class) {
		lookup := reqs.DomainClassesLookup()
		return lookup[key.String()]
	},
	"generalization_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_class.Generalization) {
		lookup := reqs.GeneralizationLookup()
		return lookup[key.String()]
	},
	"generalization_superclass": func(reqs *req_flat.Requirements, key identity.Key) (class model_class.Class) {
		lookup := reqs.GeneralizationSuperclassLookup()
		return lookup[key.String()]
	},
	"generalization_subclasses": func(reqs *req_flat.Requirements, key identity.Key) (classes []model_class.Class) {
		lookup := reqs.GeneralizationSubclassesLookup()
		return lookup[key.String()]
	},
	"action_transitions": func(reqs *req_flat.Requirements, key identity.Key) (transitions []model_state.Transition) {
		lookup := reqs.ActionTransitionsLookup()
		return lookup[key.String()]
	},
	"action_state_actions": func(reqs *req_flat.Requirements, key identity.Key) (stateActions []model_state.StateAction) {
		lookup := reqs.ActionStateActionsLookup()
		return lookup[key.String()]
	},
	"state_action_state": func(reqs *req_flat.Requirements, stateActionKey identity.Key) (state model_state.State) {
		// StateAction's key's parent is the State key.
		stateKeyStr := stateActionKey.GetParentKey()
		lookup := reqs.StateLookup()
		return lookup[stateKeyStr]
	},
	// class_actor_key returns the ActorKey for a Class (used when mapping use case actors to diagram nodes).
	"class_actor_key": func(reqs *req_flat.Requirements, classKey identity.Key) (actorKey *identity.Key) {
		classLookup, _ := reqs.ClassLookup()
		class := classLookup[classKey.String()]
		return class.ActorKey
	},

	// Lookup methods for types not yet exposed to templates.
	"query_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_state.Query) {
		lookup := reqs.QueryLookup()
		return lookup[key.String()]
	},
	"global_function_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_logic.GlobalFunction) {
		lookup := reqs.GlobalFunctionLookup()
		return lookup[key.String()]
	},
	"invariant_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_logic.Logic) {
		lookup := reqs.InvariantLookup()
		return lookup[key.String()]
	},
	"class_invariant_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_logic.Logic) {
		lookup := reqs.ClassInvariantLookup()
		return lookup[key.String()]
	},
	"object_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_scenario.Object) {
		lookup := reqs.ObjectLookup()
		return lookup[key.String()]
	},
	"actor_generalization_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_actor.Generalization) {
		lookup := reqs.ActorGeneralizationLookup()
		return lookup[key.String()]
	},
	"actor_generalization_superclass": func(reqs *req_flat.Requirements, key identity.Key) (value model_actor.Actor) {
		lookup := reqs.ActorGeneralizationSuperclassLookup()
		return lookup[key.String()]
	},
	"actor_generalization_subclasses": func(reqs *req_flat.Requirements, key identity.Key) (values []model_actor.Actor) {
		lookup := reqs.ActorGeneralizationSubclassesLookup()
		return lookup[key.String()]
	},
	"use_case_generalization_lookup": func(reqs *req_flat.Requirements, key identity.Key) (value model_use_case.Generalization) {
		lookup := reqs.UseCaseGeneralizationLookup()
		return lookup[key.String()]
	},
	"use_case_generalization_superclass": func(reqs *req_flat.Requirements, key identity.Key) (value model_use_case.UseCase) {
		lookup := reqs.UseCaseGeneralizationSuperclassLookup()
		return lookup[key.String()]
	},
	"use_case_generalization_subclasses": func(reqs *req_flat.Requirements, key identity.Key) (values []model_use_case.UseCase) {
		lookup := reqs.UseCaseGeneralizationSubclassesLookup()
		return lookup[key.String()]
	},
	// use_case_includes returns the mud-level use cases that a sea-level use case includes or extends.
	// Each result has ShareType ("include" or "extend") and the UseCase.
	"use_case_includes": func(reqs *req_flat.Requirements, key identity.Key) (results []UseCaseShare) {
		subdomainLookup := reqs.UseCaseSubdomainLookup()
		subdomain, found := subdomainLookup[key.String()]
		if !found {
			return nil
		}
		mudShares, found := subdomain.UseCaseShares[key]
		if !found {
			return nil
		}
		useCaseLookup := reqs.UseCaseLookup()
		for mudKey, share := range mudShares {
			if uc, ok := useCaseLookup[mudKey.String()]; ok {
				results = append(results, UseCaseShare{ShareType: share.ShareType, UseCase: uc})
			}
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].UseCase.Key.String() < results[j].UseCase.Key.String()
		})
		return results
	},
	// use_case_extended_by returns the sea-level use cases that include or extend a mud-level use case.
	"use_case_extended_by": func(reqs *req_flat.Requirements, key identity.Key) (results []UseCaseShare) {
		subdomainLookup := reqs.UseCaseSubdomainLookup()
		subdomain, found := subdomainLookup[key.String()]
		if !found {
			return nil
		}
		useCaseLookup := reqs.UseCaseLookup()
		for seaKey, mudShares := range subdomain.UseCaseShares {
			for mudKey, share := range mudShares {
				if mudKey == key {
					if uc, ok := useCaseLookup[seaKey.String()]; ok {
						results = append(results, UseCaseShare{ShareType: share.ShareType, UseCase: uc})
					}
				}
			}
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].UseCase.Key.String() < results[j].UseCase.Key.String()
		})
		return results
	},
	"class_indexes": func(attributes map[identity.Key]model_class.Attribute) (indexes [][]string) {
		// Group attribute names by index number.
		indexMap := map[uint][]string{}
		for _, attr := range attributes {
			for _, idx := range attr.IndexNums {
				indexMap[idx] = append(indexMap[idx], attr.Name)
			}
		}
		if len(indexMap) == 0 {
			return nil
		}
		// Collect and sort index numbers.
		var nums []uint
		for num := range indexMap {
			nums = append(nums, num)
		}
		slices.Sort(nums)
		// Build sorted result.
		for _, num := range nums {
			names := indexMap[num]
			sort.Strings(names)
			indexes = append(indexes, names)
		}
		return indexes
	},
}

// Split multi-line bullets into sub bullets.
func splitBulletTextIntoMainAndSubBullets(bulletText string) (mainBullet string, subBullets []string) {
	// If the text we want to put in a bullet is multiple lines,
	// every line after the first is a subbullet.

	// First clean up the edge whitespace.
	trimmedText := strings.TrimSpace(bulletText)
	if trimmedText == "" {
		return "", nil
	}

	// Main bullet.
	parts := strings.Split(trimmedText, "\n")
	mainBullet = strings.TrimSpace(parts[0])

	// Rest of bullets.
	for i := 1; i < len(parts); i++ {
		subBullets = append(subBullets, strings.TrimSpace(parts[i]))
	}

	return mainBullet, subBullets
}

// Convert an object key to a filename.
func convertKeyToFilename(objType, key, suffix, ext string) (filename string) {
	baseFilename := objType + "-" + strings.ReplaceAll(key, "/", ".")
	fullSuffix := ""
	if suffix != "" {
		fullSuffix = "-" + suffix
	}
	return baseFilename + fullSuffix + ext
}

func generateFromTemplate(template *template.Template, data any) (contents string, err error) {
	// Create a buffer to hold the output.
	var buf bytes.Buffer

	// Execute the template into the buffer.
	err = template.Execute(&buf, data)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// Convert the buffer to a string variable.
	contents = buf.String()

	return contents, nil
}
