package generate

import (
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
)

func expressionSpecDisplayForClass(class model_class.Class, spec logic_spec.ExpressionSpec) string {
	display := expressionSpecDisplay(spec)
	if display == "" {
		return ""
	}
	return applySelfAttributeTLADisplayNames(class, display)
}

func applySelfAttributeTLADisplayNames(class model_class.Class, specification string) string {
	replacements := selfAttributeTLADisplayReplacements(class)
	if len(replacements) == 0 {
		return specification
	}
	result := specification
	for _, repl := range replacements {
		result = replaceSelfAttributeFieldReference(result, repl.key, repl.display)
	}
	return result
}

type selfAttributeTLAReplacement struct {
	key     string
	display string
}

func selfAttributeTLADisplayReplacements(class model_class.Class) []selfAttributeTLAReplacement {
	var replacements []selfAttributeTLAReplacement
	for _, attr := range class.Attributes {
		key := attr.Key.SubKey
		display := model_class.AttributeTLAFieldName(attr.Name)
		if key == "" || key == display {
			continue
		}
		replacements = append(replacements, selfAttributeTLAReplacement{key: key, display: display})
	}
	if len(replacements) == 0 {
		return nil
	}

	sort.Slice(replacements, func(i, j int) bool {
		return len(replacements[i].key) > len(replacements[j].key)
	})
	return replacements
}

func replaceSelfAttributeFieldReference(specification, key, display string) string {
	prefix := "self." + key
	replacement := "self." + display
	var b strings.Builder
	i := 0
	for i < len(specification) {
		idx := strings.Index(specification[i:], prefix)
		if idx < 0 {
			b.WriteString(specification[i:])
			break
		}
		abs := i + idx
		end := abs + len(prefix)
		if end < len(specification) && isTLAIdentifierContinue(specification[end]) {
			b.WriteString(specification[i : abs+1])
			i = abs + 1
			continue
		}
		b.WriteString(specification[i:abs])
		b.WriteString(replacement)
		i = end
	}
	return b.String()
}

func isTLAIdentifierContinue(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_'
}

func expressionSpecBoldDisplayForClass(class model_class.Class, spec logic_spec.ExpressionSpec) string {
	display := expressionSpecDisplayForClass(class, spec)
	if display == "" {
		return ""
	}
	return "**" + display + "**"
}
