package generate

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

const missingSpecDisplay = "???"

func expressionSpecBoldDisplay(spec logic_spec.ExpressionSpec) string {
	display := expressionSpecDisplay(spec)
	if display == "" {
		return ""
	}
	return "**" + display + "**"
}

func logicBoldSpecTextForClass(class model_class.Class, logic model_logic.Logic) string {
	if logic.Target != "" {
		spec := missingSpecDisplay
		if logic.Spec.Specification != "" {
			spec = expressionSpecDisplayForClass(class, logic.Spec)
		}
		switch logic.Type {
		case model_logic.LogicTypeStateChange, model_logic.LogicTypeQuery, model_logic.LogicTypeDestroy:
			return "**" + logic.Target + "' = " + spec + "**"
		case model_logic.LogicTypeLet:
			return "**LET " + logic.Target + " = " + spec + "**"
		default:
			return "**LET " + logic.Target + " = " + spec + "**"
		}
	}
	if logic.Spec.Specification != "" {
		return expressionSpecBoldDisplayForClass(class, logic.Spec)
	}
	return ""
}

func logicBoldSpecText(logic model_logic.Logic) string {
	if logic.Target != "" {
		spec := missingSpecDisplay
		if logic.Spec.Specification != "" {
			spec = expressionSpecDisplay(logic.Spec)
		}
		switch logic.Type {
		case model_logic.LogicTypeStateChange, model_logic.LogicTypeQuery, model_logic.LogicTypeDestroy:
			return "**" + logic.Target + "' = " + spec + "**"
		case model_logic.LogicTypeLet:
			return "**LET " + logic.Target + " = " + spec + "**"
		default:
			return "**LET " + logic.Target + " = " + spec + "**"
		}
	}
	if logic.Spec.Specification != "" {
		return expressionSpecBoldDisplay(logic.Spec)
	}
	return ""
}

// classLogicMarkdownSpecLinesFromTemplate is the template entry point (class, logic, reqs).
func classLogicMarkdownSpecLinesFromTemplate(
	class model_class.Class,
	logic model_logic.Logic,
	reqs *req_flat.Requirements,
) string {
	if reqs == nil {
		return logicMarkdownSpecLinesForClass(class, logic, nil, nil)
	}
	return logicMarkdownSpecLinesForClass(class, logic, reqs.ClassAssociations, reqs.Classes)
}

// logicMarkdownSpecLinesForClass renders logic spec lines with self.<key> mapped to display TLA names.
// associations/classes resolve host association names for association-class reify selectors.
func logicMarkdownSpecLinesForClass(
	class model_class.Class,
	logic model_logic.Logic,
	associations map[identity.Key]model_class.Association,
	classes map[identity.Key]model_class.Class,
) string {
	var lines []string
	if model_logic.IsAssociationClassReify(logic) {
		lines = append(lines, associationClassReifyMarkdownLines(class, logic, associations, classes)...)
	} else if bold := logicBoldSpecTextForClass(class, logic); bold != "" {
		lines = append(lines, "    - "+bold)
	}
	if logic.Type == model_logic.LogicTypeDestroy {
		if event := strings.TrimSpace(logic.DestroyEventSpec.Specification); event != "" {
			lines = append(lines, "    - Each removed element sent: "+destroyEventSpecBoldDisplay(logic.DestroyEventSpec))
		}
	}
	if logic.Target != "" && logic.TargetTypeSpec != nil {
		if typeSpec := strings.TrimSpace(logic.TargetTypeSpec.Specification); typeSpec != "" {
			lines = append(lines, "    - Type: "+typeSpec)
		}
	}
	return strings.Join(lines, "\n")
}

// associationClassReifyMarkdownLines renders:
//
//   - **Adjusts selector: { r.account : r ∈ Amounts }**
//   - **AccountBalanceChange' = «new»(r.amount)**
func associationClassReifyMarkdownLines(
	class model_class.Class,
	logic model_logic.Logic,
	associations map[identity.Key]model_class.Association,
	classes map[identity.Key]model_class.Class,
) []string {
	var lines []string
	assocLabel := associationTLAFieldForACClassTarget(logic.Target, associations, classes)
	if assocLabel == "" {
		assocLabel = missingSpecDisplay
	}
	selectorSpec := missingSpecDisplay
	if strings.TrimSpace(logic.EndpointSelectorSpec.Specification) != "" {
		selectorSpec = expressionSpecDisplayForClass(class, logic.EndpointSelectorSpec)
	}
	lines = append(lines, "    - **"+assocLabel+" selector: "+selectorSpec+"**")

	createSpec := missingSpecDisplay
	if strings.TrimSpace(logic.Spec.Specification) != "" {
		createSpec = expressionSpecDisplayForClass(class, logic.Spec)
		createSpec = systemEventCallSpecDisplay(createSpec)
	}
	lines = append(lines, "    - **"+logic.Target+"' = "+createSpec+"**")
	return lines
}

// associationTLAFieldForACClassTarget returns the host association TLA field name for an AC class target.
func associationTLAFieldForACClassTarget(
	acClassTLAName string,
	associations map[identity.Key]model_class.Association,
	classes map[identity.Key]model_class.Class,
) string {
	for _, assoc := range associations {
		if assoc.AssociationClassKey == nil {
			continue
		}
		acClass, ok := classes[*assoc.AssociationClassKey]
		if !ok {
			continue
		}
		if model_class.ClassTLAName(acClass.Name) == acClassTLAName {
			return model_class.AssociationTLAFieldName(assoc.Name)
		}
	}
	return ""
}

// logicMarkdownSpecLines renders indented continuation lines for a logic item in list markdown.
func logicMarkdownSpecLines(logic model_logic.Logic) string {
	var lines []string
	if model_logic.IsAssociationClassReify(logic) {
		// Without association catalog, still show selector before AC assignment.
		selectorSpec := missingSpecDisplay
		if strings.TrimSpace(logic.EndpointSelectorSpec.Specification) != "" {
			selectorSpec = expressionSpecDisplay(logic.EndpointSelectorSpec)
		}
		lines = append(lines, "    - **selector: "+selectorSpec+"**")
		createSpec := missingSpecDisplay
		if strings.TrimSpace(logic.Spec.Specification) != "" {
			createSpec = systemEventCallSpecDisplay(expressionSpecDisplay(logic.Spec))
		}
		lines = append(lines, "    - **"+logic.Target+"' = "+createSpec+"**")
	} else if bold := logicBoldSpecText(logic); bold != "" {
		lines = append(lines, "    - "+bold)
	}
	if logic.Type == model_logic.LogicTypeDestroy {
		if event := strings.TrimSpace(logic.DestroyEventSpec.Specification); event != "" {
			lines = append(lines, "    - Each removed element sent: "+destroyEventSpecBoldDisplay(logic.DestroyEventSpec))
		}
	}
	if logic.Target != "" && logic.TargetTypeSpec != nil {
		if typeSpec := strings.TrimSpace(logic.TargetTypeSpec.Specification); typeSpec != "" {
			lines = append(lines, "    - Type: "+typeSpec)
		}
	}
	return strings.Join(lines, "\n")
}

// destroyEventSpecBoldDisplay renders destroy_event using canonical TLA+ system event spellings.
func destroyEventSpecBoldDisplay(spec logic_spec.ExpressionSpec) string {
	display := systemEventCallSpecDisplay(spec.Specification)
	if display == "" {
		return ""
	}
	displaySpec := logic_spec.ExpressionSpec{Notation: spec.Notation, Specification: display}
	return expressionSpecBoldDisplay(displaySpec)
}

func systemEventCallSpecDisplay(specification string) string {
	display := specification
	display = strings.ReplaceAll(display, model_state.EventNameNew+"(", model_state.EventTLANameNew+"(")
	display = strings.ReplaceAll(display, model_state.EventNameDestroy+"(", model_state.EventTLANameDestroy+"(")
	return display
}

func expressionSpecBoldIndentedLine(spec logic_spec.ExpressionSpec) string {
	if spec.Specification == "" {
		return ""
	}
	return "    - " + expressionSpecBoldDisplay(spec)
}

func derivationPolicyMarkdownHTMLForClass(class model_class.Class, policy *model_logic.Logic) string {
	if policy == nil {
		return ""
	}
	var parts []string
	if desc := strings.TrimSpace(policy.Description); desc != "" {
		parts = append(parts, desc)
	}
	if policy.Spec.Specification != "" {
		parts = append(parts, expressionSpecBoldDisplayForClass(class, policy.Spec))
	}
	return strings.Join(parts, "<br>")
}

func indentMarkdownBlock(block, indent string) string {
	if block == "" {
		return ""
	}
	lines := strings.Split(block, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		out = append(out, indent+line)
	}
	return strings.Join(out, "\n")
}

func appendClassLogicMarkdownChild(lines *[]string, class model_class.Class, logic model_logic.Logic, parentIndent string) {
	if specLines := logicMarkdownSpecLinesForClass(class, logic, nil, nil); specLines != "" {
		*lines = append(*lines, indentMarkdownBlock(specLines, parentIndent))
	}
}

// parameterSimulationMarkdownLines renders simulator-only sampling metadata under an action parameter.
func parameterSimulationMarkdownLines(class model_class.Class, param model_state.Parameter) string {
	if param.Simulation == nil || !param.Simulation.HasSimulation() {
		return ""
	}
	sim := param.Simulation
	var lines []string
	lines = append(lines, "    - Simulation:")
	if details := strings.TrimSpace(sim.Details); details != "" {
		lines = append(lines, "        - "+details)
	}
	for _, rule := range sim.Rules {
		ruleHeader := "        - Rule:"
		if desc := strings.TrimSpace(rule.Details); desc != "" {
			ruleHeader = "        - " + desc
		}
		lines = append(lines, ruleHeader)
		if len(rule.Requires) > 0 {
			lines = append(lines, "            - Requires:")
			for _, req := range rule.Requires {
				reqIndent := "                "
				if desc := strings.TrimSpace(req.Description); desc != "" {
					lines = append(lines, reqIndent+"- "+desc)
					appendClassLogicMarkdownChild(&lines, class, req, reqIndent)
					continue
				}
				appendClassLogicMarkdownChild(&lines, class, req, "            ")
			}
		}
		if rule.Specification != nil {
			specHeaderIndent := "            "
			if desc := strings.TrimSpace(rule.Specification.Description); desc != "" {
				lines = append(lines, specHeaderIndent+"- "+desc)
				appendClassLogicMarkdownChild(&lines, class, *rule.Specification, specHeaderIndent)
			} else {
				lines = append(lines, specHeaderIndent+"- Specification:")
				appendClassLogicMarkdownChild(&lines, class, *rule.Specification, specHeaderIndent)
			}
		}
	}
	return strings.Join(lines, "\n")
}

func derivationPolicyMarkdownHTML(policy *model_logic.Logic) string {
	if policy == nil {
		return ""
	}
	var parts []string
	if desc := strings.TrimSpace(policy.Description); desc != "" {
		parts = append(parts, desc)
	}
	if policy.Spec.Specification != "" {
		parts = append(parts, expressionSpecBoldDisplay(policy.Spec))
	}
	return strings.Join(parts, "<br>")
}
