package parser_human

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_named_set"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_spec"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseModel(key, filename, contents string) (model req_model.Model, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return req_model.Model{}, err
	}

	// Parse the YAML data section for invariants, global functions, and named sets.
	var invariants []model_logic.Logic
	var globalFunctions map[identity.Key]model_logic.GlobalFunction
	var namedSets map[identity.Key]model_named_set.NamedSet

	if parsedFile.Data != "" {
		yamlData := map[string]any{}
		if err := yaml.Unmarshal([]byte(parsedFile.Data), &yamlData); err != nil {
			return req_model.Model{}, errors.WithStack(err)
		}

		// Parse invariants using logicListFromYamlData with a wrapper for NewInvariantKey.
		invariantKeyFunc := func(_ identity.Key, subKey string) (identity.Key, error) {
			return identity.NewInvariantKey(subKey)
		}
		invariants, err = logicListFromYamlData(yamlData, "invariants",
			model_logic.LogicTypeAssessment, identity.Key{}, invariantKeyFunc)
		if err != nil {
			return req_model.Model{}, errors.Wrap(err, "model invariants")
		}

		// Parse global functions (written as a list in YAML, stored as a map in the model).
		if gfsAny, found := yamlData["global_functions"]; found {
			gfsList, ok := gfsAny.([]any)
			if !ok {
				return req_model.Model{}, errors.Errorf("global_functions must be a list")
			}
			globalFunctions = make(map[identity.Key]model_logic.GlobalFunction, len(gfsList))
			for _, gfAny := range gfsList {
				gfMap, ok := gfAny.(map[string]any)
				if !ok {
					return req_model.Model{}, errors.Errorf("each global_function must be a map")
				}

				name := ""
				if n, ok := gfMap["name"]; ok {
					name = n.(string)
				}

				// Build the key from the lowercase name.
				gfKey, err := identity.NewGlobalFunctionKey(strings.ToLower(name))
				if err != nil {
					return req_model.Model{}, errors.WithStack(err)
				}

				// Parse parameters.
				var parameters []string
				if p, ok := gfMap["parameters"]; ok {
					paramsList, ok := p.([]any)
					if !ok {
						return req_model.Model{}, errors.Errorf("global function parameters must be a list")
					}
					for _, param := range paramsList {
						parameters = append(parameters, param.(string))
					}
				}

				// Logic members are inline in the global function YAML entry.
				description := ""
				if d, ok := gfMap["description"]; ok {
					description = d.(string)
				}
				specification := ""
				if s, ok := gfMap["specification"]; ok {
					specification = s.(string)
				}

				spec, err := model_spec.NewExpressionSpec(model_logic.NotationTLAPlus, specification, nil)
				if err != nil {
					return req_model.Model{}, errors.Wrapf(err, "global function %q expression spec", name)
				}

				logic, err := model_logic.NewLogic(gfKey, model_logic.LogicTypeValue, description, "", spec, nil)
				if err != nil {
					return req_model.Model{}, errors.Wrapf(err, "global function %q logic", name)
				}

				gf, err := model_logic.NewGlobalFunction(gfKey, name, parameters, logic)
				if err != nil {
					return req_model.Model{}, errors.Wrapf(err, "global function %q", name)
				}
				globalFunctions[gfKey] = gf
			}
		}

		// Parse named sets (written as a list in YAML, stored as a map in the model).
		if nsAny, found := yamlData["named_sets"]; found {
			nsList, ok := nsAny.([]any)
			if !ok {
				return req_model.Model{}, errors.Errorf("named_sets must be a list")
			}
			namedSets = make(map[identity.Key]model_named_set.NamedSet, len(nsList))
			for _, nsItemAny := range nsList {
				nsMap, ok := nsItemAny.(map[string]any)
				if !ok {
					return req_model.Model{}, errors.Errorf("each named_set must be a map")
				}

				name := ""
				if n, ok := nsMap["name"]; ok {
					name = n.(string)
				}

				nsKey, err := identity.NewNamedSetKey(strings.ToLower(name))
				if err != nil {
					return req_model.Model{}, errors.WithStack(err)
				}

				description := ""
				if d, ok := nsMap["description"]; ok {
					description = d.(string)
				}

				specification := ""
				if s, ok := nsMap["specification"]; ok {
					specification = s.(string)
				}

				spec, err := model_spec.NewExpressionSpec(model_logic.NotationTLAPlus, specification, nil)
				if err != nil {
					return req_model.Model{}, errors.Wrapf(err, "named set %q expression spec", name)
				}

				var typeSpec *model_spec.TypeSpec
				if tsStr, ok := nsMap["type_spec"].(string); ok && tsStr != "" {
					ts, err := model_spec.NewTypeSpec(model_logic.NotationTLAPlus, tsStr, nil)
					if err != nil {
						return req_model.Model{}, errors.Wrapf(err, "named set %q type spec", name)
					}
					typeSpec = &ts
				}

				ns, err := model_named_set.NewNamedSet(nsKey, name, description, spec, typeSpec)
				if err != nil {
					return req_model.Model{}, errors.Wrapf(err, "named set %q", name)
				}
				namedSets[nsKey] = ns
			}
		}
	}

	// There is no uml comment for a "model" entity (it is not displayed).
	markdown := stripMarkdownTitle(parsedFile.Markdown)

	if parsedFile.UmlComment != "" {
		markdown += "\n\n" + parsedFile.UmlComment
	}

	model, err = req_model.NewModel(
		strings.TrimSpace(strings.ToLower(key)),
		parsedFile.Title,
		markdown,
		invariants,
		globalFunctions,
		namedSets,
	)
	if err != nil {
		return req_model.Model{}, errors.Wrap(err, "failed to create model")
	}

	return model, nil
}

func generateModelContent(model req_model.Model) string {
	builder := NewYamlBuilder()

	// Generate invariants YAML.
	generateLogicSequence(builder, "invariants", model.Invariants)

	// Generate global functions YAML (written as a list, sorted by key for deterministic output).
	if len(model.GlobalFunctions) > 0 {
		// Sort global functions by key for deterministic output.
		var gfKeys []identity.Key
		for k := range model.GlobalFunctions {
			gfKeys = append(gfKeys, k)
		}
		sort.Slice(gfKeys, func(i, j int) bool {
			return gfKeys[i].String() < gfKeys[j].String()
		})

		var gfBuilders []*YamlBuilder
		for _, gfKey := range gfKeys {
			gf := model.GlobalFunctions[gfKey]
			gfBuilder := NewYamlBuilder()
			gfBuilder.AddField("name", gf.Name)
			if len(gf.Parameters) > 0 {
				gfBuilder.AddSequenceField("parameters", gf.Parameters)
			}
			gfBuilder.AddField("description", gf.Logic.Description)
			gfBuilder.AddQuotedField("specification", gf.Logic.Spec.Specification)
			gfBuilders = append(gfBuilders, gfBuilder)
		}
		builder.AddSequenceOfMappings("global_functions", gfBuilders)
	}

	// Generate named sets YAML (written as a list, sorted by key for deterministic output).
	if len(model.NamedSets) > 0 {
		var nsKeys []identity.Key
		for k := range model.NamedSets {
			nsKeys = append(nsKeys, k)
		}
		sort.Slice(nsKeys, func(i, j int) bool {
			return nsKeys[i].String() < nsKeys[j].String()
		})

		var nsBuilders []*YamlBuilder
		for _, nsKey := range nsKeys {
			ns := model.NamedSets[nsKey]
			nsBuilder := NewYamlBuilder()
			nsBuilder.AddField("name", ns.Name)
			nsBuilder.AddField("description", ns.Description)
			nsBuilder.AddQuotedField("specification", ns.Spec.Specification)
			if ns.TypeSpec != nil && ns.TypeSpec.Specification != "" {
				nsBuilder.AddField("type_spec", ns.TypeSpec.Specification)
			}
			nsBuilders = append(nsBuilders, nsBuilder)
		}
		builder.AddSequenceOfMappings("named_sets", nsBuilders)
	}

	dataStr, _ := builder.Build()

	return generateFileContent(prependMarkdownTitle(model.Name, model.Details), "", dataStr)
}

// yamlQuote wraps a string in appropriate YAML quotes.
// Uses single quotes if the string contains backslashes (to preserve TLA+ notation like \sqsubseteq),
// double quotes if it contains special unicode, or no quotes if plain.
func yamlQuote(s string) string {
	if strings.ContainsRune(s, '\\') {
		return "'" + s + "'"
	}
	// Check if string contains non-ASCII characters that need quoting.
	for _, r := range s {
		if r > 127 {
			return fmt.Sprintf("%q", s)
		}
	}
	return s
}
