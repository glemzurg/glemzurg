package parser_human

import (
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_named_set"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseModel(key, filename, contents string) (model core.Model, err error) {
	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return core.Model{}, err
	}

	// Parse the YAML data section for invariants, global functions, and named sets.
	var invariants []model_logic.Logic
	var globalFunctions map[identity.Key]model_logic.GlobalFunction
	var namedSets map[identity.Key]model_named_set.NamedSet

	if parsedFile.Data != "" {
		yamlData := map[string]any{}
		if err := yaml.Unmarshal([]byte(parsedFile.Data), &yamlData); err != nil {
			return core.Model{}, errors.WithStack(err)
		}

		invariantKeyFunc := func(_ identity.Key, subKey string) (identity.Key, error) {
			return identity.NewInvariantKey(subKey)
		}
		invariants, err = logicListFromYamlData(yamlData, "invariants",
			model_logic.LogicTypeAssessment, identity.Key{}, invariantKeyFunc)
		if err != nil {
			return core.Model{}, errors.Wrap(err, "model invariants")
		}

		globalFunctions, err = parseGlobalFunctions(yamlData)
		if err != nil {
			return core.Model{}, err
		}

		namedSets, err = parseNamedSets(yamlData)
		if err != nil {
			return core.Model{}, err
		}
	}

	// There is no uml comment for a "model" entity (it is not displayed).
	markdown := stripMarkdownTitle(parsedFile.Markdown)

	if parsedFile.UmlComment != "" {
		markdown += "\n\n" + parsedFile.UmlComment
	}

	model, err = core.NewModel(
		strings.TrimSpace(strings.ToLower(key)),
		parsedFile.Title,
		markdown,
		invariants,
		globalFunctions,
		namedSets,
	)
	if err != nil {
		return core.Model{}, errors.Wrap(err, "failed to create model")
	}

	return model, nil
}

// parseGlobalFunctions parses the global_functions list from YAML data.
func parseGlobalFunctions(yamlData map[string]any) (map[identity.Key]model_logic.GlobalFunction, error) {
	gfsAny, found := yamlData["global_functions"]
	if !found {
		return nil, nil
	}
	gfsList, ok := gfsAny.([]any)
	if !ok {
		return nil, errors.Errorf("global_functions must be a list")
	}
	globalFunctions := make(map[identity.Key]model_logic.GlobalFunction, len(gfsList))
	for _, gfAny := range gfsList {
		gfMap, ok := gfAny.(map[string]any)
		if !ok {
			return nil, errors.Errorf("each global_function must be a map")
		}
		gf, err := parseOneGlobalFunction(gfMap)
		if err != nil {
			return nil, err
		}
		globalFunctions[gf.Key] = gf
	}
	return globalFunctions, nil
}

// parseOneGlobalFunction parses a single global function from a YAML map.
func parseOneGlobalFunction(gfMap map[string]any) (model_logic.GlobalFunction, error) {
	name := ""
	if n, ok := gfMap["name"]; ok {
		name = n.(string)
	}

	gfKey, err := identity.NewGlobalFunctionKey(strings.ToLower(name))
	if err != nil {
		return model_logic.GlobalFunction{}, errors.WithStack(err)
	}

	var parameters []string
	if p, ok := gfMap["parameters"]; ok {
		paramsList, ok := p.([]any)
		if !ok {
			return model_logic.GlobalFunction{}, errors.Errorf("global function parameters must be a list")
		}
		for _, param := range paramsList {
			parameters = append(parameters, param.(string))
		}
	}

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
		return model_logic.GlobalFunction{}, errors.Wrapf(err, "global function %q expression spec", name)
	}

	logic, err := model_logic.NewLogic(gfKey, model_logic.LogicTypeValue, description, "", spec, nil)
	if err != nil {
		return model_logic.GlobalFunction{}, errors.Wrapf(err, "global function %q logic", name)
	}

	gf, err := model_logic.NewGlobalFunction(gfKey, name, parameters, logic)
	if err != nil {
		return model_logic.GlobalFunction{}, errors.Wrapf(err, "global function %q", name)
	}
	return gf, nil
}

// parseNamedSets parses the named_sets list from YAML data.
func parseNamedSets(yamlData map[string]any) (map[identity.Key]model_named_set.NamedSet, error) {
	nsAny, found := yamlData["named_sets"]
	if !found {
		return nil, nil
	}
	nsList, ok := nsAny.([]any)
	if !ok {
		return nil, errors.Errorf("named_sets must be a list")
	}
	namedSets := make(map[identity.Key]model_named_set.NamedSet, len(nsList))
	for _, nsItemAny := range nsList {
		nsMap, ok := nsItemAny.(map[string]any)
		if !ok {
			return nil, errors.Errorf("each named_set must be a map")
		}
		ns, err := parseOneNamedSet(nsMap)
		if err != nil {
			return nil, err
		}
		namedSets[ns.Key] = ns
	}
	return namedSets, nil
}

// parseOneNamedSet parses a single named set from a YAML map.
func parseOneNamedSet(nsMap map[string]any) (model_named_set.NamedSet, error) {
	name := ""
	if n, ok := nsMap["name"]; ok {
		name = n.(string)
	}

	nsKey, err := identity.NewNamedSetKey(strings.ToLower(strings.TrimPrefix(name, "_")))
	if err != nil {
		return model_named_set.NamedSet{}, errors.WithStack(err)
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
		return model_named_set.NamedSet{}, errors.Wrapf(err, "named set %q expression spec", name)
	}

	var typeSpec *model_spec.TypeSpec
	if tsStr, ok := nsMap["type_spec"].(string); ok && tsStr != "" {
		ts, err := model_spec.NewTypeSpec(model_logic.NotationTLAPlus, tsStr, nil)
		if err != nil {
			return model_named_set.NamedSet{}, errors.Wrapf(err, "named set %q type spec", name)
		}
		typeSpec = &ts
	}

	ns, err := model_named_set.NewNamedSet(nsKey, name, description, spec, typeSpec)
	if err != nil {
		return model_named_set.NamedSet{}, errors.Wrapf(err, "named set %q", name)
	}
	return ns, nil
}

func generateModelContent(model core.Model) string {
	builder := NewYamlBuilder()

	// Generate invariants YAML.
	generateLogicSequence(builder, "invariants", model.Invariants)

	// Generate global functions YAML.
	generateGlobalFunctionsYaml(builder, model.GlobalFunctions)

	// Generate named sets YAML.
	generateNamedSetsYaml(builder, model.NamedSets)

	dataStr, _ := builder.Build()

	return generateFileContent(prependMarkdownTitle(model.Name, model.Details), "", dataStr)
}

// generateGlobalFunctionsYaml generates the global_functions YAML section.
func generateGlobalFunctionsYaml(builder *YamlBuilder, globalFunctions map[identity.Key]model_logic.GlobalFunction) {
	if len(globalFunctions) == 0 {
		return
	}
	var gfKeys []identity.Key
	for k := range globalFunctions {
		gfKeys = append(gfKeys, k)
	}
	sort.Slice(gfKeys, func(i, j int) bool {
		return gfKeys[i].String() < gfKeys[j].String()
	})

	var gfBuilders []*YamlBuilder
	for _, gfKey := range gfKeys {
		gf := globalFunctions[gfKey]
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

// generateNamedSetsYaml generates the named_sets YAML section.
func generateNamedSetsYaml(builder *YamlBuilder, namedSets map[identity.Key]model_named_set.NamedSet) {
	if len(namedSets) == 0 {
		return
	}
	var nsKeys []identity.Key
	for k := range namedSets {
		nsKeys = append(nsKeys, k)
	}
	sort.Slice(nsKeys, func(i, j int) bool {
		return nsKeys[i].String() < nsKeys[j].String()
	})

	var nsBuilders []*YamlBuilder
	for _, nsKey := range nsKeys {
		ns := namedSets[nsKey]
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
