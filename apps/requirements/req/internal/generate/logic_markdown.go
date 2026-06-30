package generate

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
)

func expressionSpecBoldDisplay(spec logic_spec.ExpressionSpec) string {
	display := expressionSpecDisplay(spec)
	if display == "" {
		return ""
	}
	return "**" + display + "**"
}

func logicBoldSpecTextForClass(class model_class.Class, logic model_logic.Logic) string {
	if logic.Target != "" {
		spec := "???"
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
		spec := "???"
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

// logicMarkdownSpecLinesForClass renders logic spec lines with self.<key> mapped to display TLA names.
func logicMarkdownSpecLinesForClass(class model_class.Class, logic model_logic.Logic) string {
	var lines []string
	if bold := logicBoldSpecTextForClass(class, logic); bold != "" {
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

// logicMarkdownSpecLines renders indented continuation lines for a logic item in list markdown.
func logicMarkdownSpecLines(logic model_logic.Logic) string {
	var lines []string
	if bold := logicBoldSpecText(logic); bold != "" {
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
