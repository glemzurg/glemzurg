package parser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func parseModel(key, filename, contents string) (model req_model.Model, err error) {

	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return req_model.Model{}, err
	}

	// Parse the YAML data section for invariants and global functions.
	var invariants []model_logic.Logic
	var globalFunctions map[identity.Key]model_logic.GlobalFunction

	if parsedFile.Data != "" {
		yamlData := map[string]any{}
		if err := yaml.Unmarshal([]byte(parsedFile.Data), &yamlData); err != nil {
			return req_model.Model{}, errors.WithStack(err)
		}

		// Parse invariants.
		if invariantsAny, found := yamlData["invariants"]; found {
			invariantsList, ok := invariantsAny.([]any)
			if !ok {
				return req_model.Model{}, errors.Errorf("invariants must be a list")
			}
			for _, invAny := range invariantsList {
				invMap, ok := invAny.(map[string]any)
				if !ok {
					return req_model.Model{}, errors.Errorf("each invariant must be a map")
				}

				name := ""
				if n, ok := invMap["name"]; ok {
					name = n.(string)
				}

				// Build the key from the name.
				invKey, err := identity.NewInvariantKey(strings.ToLower(name))
				if err != nil {
					return req_model.Model{}, errors.WithStack(err)
				}

				description := ""
				if d, ok := invMap["description"]; ok {
					description = d.(string)
				}
				specification := ""
				if s, ok := invMap["specification"]; ok {
					specification = s.(string)
				}

				inv := model_logic.Logic{
					Key:           invKey,
					Description:   description,
					Notation:      model_logic.NotationTLAPlus,
					Specification: specification,
				}
				invariants = append(invariants, inv)
			}
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

				gf := model_logic.GlobalFunction{
					Key:        gfKey,
					Name:       name,
					Parameters: parameters,
					Logic: model_logic.Logic{
						Key:           gfKey,
						Description:   description,
						Notation:      model_logic.NotationTLAPlus,
						Specification: specification,
					},
				}
				globalFunctions[gfKey] = gf
			}
		}
	}

	// There is no uml comment for a "model" entity (it is not displayed).
	markdown := parsedFile.Markdown

	if parsedFile.UmlComment != "" {
		markdown += "\n\n" + parsedFile.UmlComment
	}

	// Construct the model directly without calling NewModel, since invariants
	// and global functions parsed from YAML may not yet have full keys assigned.
	// Validation happens later in the top-level Parse flow.
	model = req_model.Model{
		Key:             strings.TrimSpace(strings.ToLower(key)),
		Name:            parsedFile.Title,
		Details:         markdown,
		Invariants:      invariants,
		GlobalFunctions: globalFunctions,
	}

	return model, nil
}

func generateModelContent(model req_model.Model) string {
	dataStr := ""

	// Generate invariants YAML.
	if len(model.Invariants) > 0 {
		dataStr += "invariants:\n"
		for _, inv := range model.Invariants {
			dataStr += "    - name: " + inv.Key.SubKey + "\n"
			dataStr += "      description: " + inv.Description + "\n"
			if inv.Specification != "" {
				dataStr += "      specification: " + yamlQuote(inv.Specification) + "\n"
			}
		}
	}

	// Generate global functions YAML (written as a list, sorted by key for deterministic output).
	if len(model.GlobalFunctions) > 0 {
		if dataStr != "" {
			dataStr += "\n"
		}

		// Sort global functions by key for deterministic output.
		var gfKeys []identity.Key
		for k := range model.GlobalFunctions {
			gfKeys = append(gfKeys, k)
		}
		sort.Slice(gfKeys, func(i, j int) bool {
			return gfKeys[i].String() < gfKeys[j].String()
		})

		dataStr += "global_functions:\n"
		for _, gfKey := range gfKeys {
			gf := model.GlobalFunctions[gfKey]
			dataStr += "    - name: " + gf.Name + "\n"
			if len(gf.Parameters) > 0 {
				paramStrs := make([]string, len(gf.Parameters))
				for i, p := range gf.Parameters {
					paramStrs[i] = fmt.Sprintf("%q", p)
				}
				dataStr += "      parameters: [" + strings.Join(paramStrs, ", ") + "]\n"
			}
			if gf.Logic.Description != "" {
				dataStr += "      description: " + gf.Logic.Description + "\n"
			}
			if gf.Logic.Specification != "" {
				dataStr += "      specification: " + yamlQuote(gf.Logic.Specification) + "\n"
			}
		}
	}

	return generateFileContent(model.Details, "", dataStr)
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
