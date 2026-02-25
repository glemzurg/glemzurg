package generate

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/pkg/errors"
)

//go:embed templates/*
var _templateFS embed.FS

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

		// Put the template into specific vars based on their use.
		// So this must be exact.
		switch tmplName {
		case "model.md.template":
			_modelMdTemplate = tmpl
		case "actor.md.template":
			_actorMdTemplate = tmpl
		case "domain.md.template":
			_domainMdTemplate = tmpl
		case "domains.dot.template":
			_domainsDotTemplate = tmpl
		case "use_cases.dot.template":
			_useCasesDotTemplate = tmpl
		case "classes.dot.template":
			_classesDotTemplate = tmpl
		case "class.md.template":
			_classMdTemplate = tmpl
		case "class-state.dot.template":
			_classStateDotTemplate = tmpl
		case "use_case.md.template":
			_useCaseMdTemplate = tmpl
		case "subdomain.md.template":
			_subdomainMdTemplate = tmpl
		case "subdomains.dot.template":
			_subdomainsDotTemplate = tmpl
		default:
			return errors.WithStack(errors.Errorf(`unknown template filename: '%s'`, tmplName))
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Failed to parse templates: %+v", err)
	}
}

// The templates in the system.
var _modelMdTemplate *template.Template
var _actorMdTemplate *template.Template
var _domainMdTemplate *template.Template
var _domainsDotTemplate *template.Template  // DOT input to GraphViz for SVG UML diagram.
var _useCasesDotTemplate *template.Template // DOT input to GraphViz for SVG UML diagram.
var _classesDotTemplate *template.Template  // DOT input to GraphViz for SVG UML diagram.
var _classMdTemplate *template.Template
var _classStateDotTemplate *template.Template // DOT input to GraphViz for SVG UML diagram.
var _useCaseMdTemplate *template.Template
var _subdomainMdTemplate *template.Template
var _subdomainsDotTemplate *template.Template // DOT input to GraphViz for SVG UML diagram.

// Define some function for our templates.
var _funcMap = template.FuncMap{
	"nodeid": func(idtype string, key identity.Key) string {
		keyStr := key.String()
		// Replace / with _
		keyStr = strings.ReplaceAll(keyStr, "/", "_")
		// Replace - with _
		keyStr = strings.ReplaceAll(keyStr, "-", "_")
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
	"first_md_paragraph": func(md string) (paragraph string) {
		return firstMdParagraph(md)
	},
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
	"action_signatures": func(reqs *req_flat.Requirements, transitions []model_state.Transition, stateActions []model_state.StateAction) (signatures []string) {

		eventLookup := reqs.EventLookup()

		// Keep track of each signature we find.
		signatureLookup := map[string]bool{}

		// If there is any state action we have a signature with no parameters.
		if len(stateActions) > 0 {
			signatureLookup[""] = true
		}

		// Create a signature for each event used.
		for _, transition := range transitions {
			event := eventLookup[transition.EventKey.String()]

			var paramNames []string
			for _, param := range event.Parameters {
				paramNames = append(paramNames, param.Name)
			}
			signature := strings.Join(paramNames, ", ")
			signatureLookup[signature] = true
		}

		// Put all the signatures in a list.
		for signature := range signatureLookup {
			signatures = append(signatures, signature)
		}

		// Make the signatures ordered for consistent display.
		sort.Strings(signatures)

		return signatures
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
