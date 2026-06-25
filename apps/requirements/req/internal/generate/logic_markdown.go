package generate

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
)

func expressionSpecBoldDisplay(spec logic_spec.ExpressionSpec) string {
	display := expressionSpecDisplay(spec)
	if display == "" {
		return ""
	}
	return "**" + display + "**"
}

func logicBoldSpecText(logic model_logic.Logic) string {
	if logic.Target != "" {
		spec := "???"
		if logic.Spec.Specification != "" {
			spec = expressionSpecDisplay(logic.Spec)
		}
		switch logic.Type {
		case model_logic.LogicTypeStateChange, model_logic.LogicTypeQuery:
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

// logicMarkdownSpecLines renders indented continuation lines for a logic item in list markdown.
func logicMarkdownSpecLines(logic model_logic.Logic) string {
	var lines []string
	if bold := logicBoldSpecText(logic); bold != "" {
		lines = append(lines, "    - "+bold)
	}
	if logic.Target != "" && logic.TargetTypeSpec != nil {
		if typeSpec := strings.TrimSpace(logic.TargetTypeSpec.Specification); typeSpec != "" {
			lines = append(lines, "    - Type: "+typeSpec)
		}
	}
	return strings.Join(lines, "\n")
}

func expressionSpecBoldIndentedLine(spec logic_spec.ExpressionSpec) string {
	if spec.Specification == "" {
		return ""
	}
	return "    - " + expressionSpecBoldDisplay(spec)
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
