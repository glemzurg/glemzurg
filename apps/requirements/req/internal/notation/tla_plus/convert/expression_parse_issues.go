package convert

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// ExpressionParseIssue records one TLA+ specification that failed to parse or lower.
type ExpressionParseIssue struct {
	ClassKey identity.Key // zero value means a model-level site.
	Location string
	Message  string
	SpecText string
}

// CollectUnparsedExpressionIssues finds every non-empty ExpressionSpec that did not
// lower successfully and diagnoses each with the strict parser so the web display
// can show actionable error text.
func CollectUnparsedExpressionIssues(model *core.Model) []ExpressionParseIssue {
	globalFunctions := BuildGlobalFunctionMap(model)
	namedSets := BuildNamedSetMap(model)
	allActions := BuildAllActionsMap(model)

	var issues []ExpressionParseIssue

	modelCtx := &LowerContext{
		GlobalFunctions: globalFunctions,
		NamedSets:       namedSets,
		AllActions:      allActions,
	}
	modelPF := NewExpressionParseFuncStrict(modelCtx)

	for i := range model.Invariants {
		if issue := diagnoseUnparsedSpec(&model.Invariants[i].Spec, modelPF, identity.Key{}, fmt.Sprintf("model invariant %d", i)); issue != nil {
			issues = append(issues, *issue)
		}
	}

	for gfKey, gf := range model.GlobalFunctions {
		params := parameterNameSet(gf.Parameters)
		gfCtx := &LowerContext{
			GlobalFunctions: globalFunctions,
			NamedSets:       namedSets,
			AllActions:      allActions,
			Parameters:      params,
		}
		gfPF := NewExpressionParseFuncStrict(gfCtx)
		loc := fmt.Sprintf("global function %q", gfKey.String())
		if issue := diagnoseUnparsedSpec(&gf.Logic.Spec, gfPF, identity.Key{}, loc); issue != nil {
			issues = append(issues, *issue)
		}
	}

	for nsKey, ns := range model.NamedSets {
		loc := fmt.Sprintf("named set %q", nsKey.String())
		if issue := diagnoseUnparsedSpec(&ns.Spec, modelPF, identity.Key{}, loc); issue != nil {
			issues = append(issues, *issue)
		}
	}

	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for classKey, class := range subdomain.Classes {
				classIssues := collectClassExpressionIssues(&class, globalFunctions, namedSets, allActions)
				for i := range classIssues {
					classIssues[i].ClassKey = classKey
					issues = append(issues, classIssues[i])
				}
			}
		}
	}

	return issues
}

func collectClassExpressionIssues(
	class *model_class.Class,
	globalFunctions, namedSets, allActions map[string]identity.Key,
) []ExpressionParseIssue {
	attrNames := BuildAttributeNameMap(class)
	actionNames := BuildActionNameMap(class)
	queryNames := BuildQueryNameMap(class)

	classCtx := &LowerContext{
		ClassKey:        class.Key,
		AttributeNames:  attrNames,
		ActionNames:     actionNames,
		QueryNames:      queryNames,
		GlobalFunctions: globalFunctions,
		NamedSets:       namedSets,
		AllActions:      allActions,
	}
	classPF := NewExpressionParseFuncStrict(classCtx)

	var issues []ExpressionParseIssue

	for i := range class.Invariants {
		loc := fmt.Sprintf("class invariant %d", i)
		if issue := diagnoseUnparsedSpec(&class.Invariants[i].Spec, classPF, class.Key, loc); issue != nil {
			issues = append(issues, *issue)
		}
	}

	for _, attr := range class.Attributes {
		if attr.DerivationPolicy != nil {
			loc := fmt.Sprintf("attribute %q derivation", attr.Key.String())
			if issue := diagnoseUnparsedSpec(&attr.DerivationPolicy.Spec, classPF, class.Key, loc); issue != nil {
				issues = append(issues, *issue)
			}
		}
		for i := range attr.Invariants {
			loc := fmt.Sprintf("attribute %q invariant %d", attr.Key.String(), i)
			if issue := diagnoseUnparsedSpec(&attr.Invariants[i].Spec, classPF, class.Key, loc); issue != nil {
				issues = append(issues, *issue)
			}
		}
	}

	for gKey, guard := range class.Guards {
		loc := fmt.Sprintf("guard %q", gKey.String())
		if issue := diagnoseUnparsedSpec(&guard.Logic.Spec, classPF, class.Key, loc); issue != nil {
			issues = append(issues, *issue)
		}
	}

	issues = append(issues, collectActionExpressionIssues(class, classCtx)...)
	issues = append(issues, collectQueryExpressionIssues(class, classCtx)...)

	return issues
}

func collectActionExpressionIssues(class *model_class.Class, classCtx *LowerContext) []ExpressionParseIssue {
	var issues []ExpressionParseIssue
	for actKey, action := range class.Actions {
		actPF := NewExpressionParseFuncStrict(ContextWithParameters(classCtx, action.Parameters))
		for i := range action.Requires {
			loc := fmt.Sprintf("action %q require %d", actKey.String(), i)
			if issue := diagnoseUnparsedSpec(&action.Requires[i].Spec, actPF, class.Key, loc); issue != nil {
				issues = append(issues, *issue)
			}
		}
		for i := range action.Guarantees {
			loc := fmt.Sprintf("action %q guarantee %d", actKey.String(), i)
			if issue := diagnoseUnparsedSpec(&action.Guarantees[i].Spec, actPF, class.Key, loc); issue != nil {
				issues = append(issues, *issue)
			}
		}
		for i := range action.SafetyRules {
			loc := fmt.Sprintf("action %q safety rule %d", actKey.String(), i)
			if issue := diagnoseUnparsedSpec(&action.SafetyRules[i].Spec, actPF, class.Key, loc); issue != nil {
				issues = append(issues, *issue)
			}
		}
	}
	return issues
}

func collectQueryExpressionIssues(class *model_class.Class, classCtx *LowerContext) []ExpressionParseIssue {
	var issues []ExpressionParseIssue
	for qKey, query := range class.Queries {
		qPF := NewExpressionParseFuncStrict(ContextWithParameters(classCtx, query.Parameters))
		for i := range query.Requires {
			loc := fmt.Sprintf("query %q require %d", qKey.String(), i)
			if issue := diagnoseUnparsedSpec(&query.Requires[i].Spec, qPF, class.Key, loc); issue != nil {
				issues = append(issues, *issue)
			}
		}
		for i := range query.Guarantees {
			loc := fmt.Sprintf("query %q guarantee %d", qKey.String(), i)
			if issue := diagnoseUnparsedSpec(&query.Guarantees[i].Spec, qPF, class.Key, loc); issue != nil {
				issues = append(issues, *issue)
			}
		}
	}
	return issues
}

func diagnoseUnparsedSpec(
	spec *logic_spec.ExpressionSpec,
	pf StrictExpressionParseFunc,
	classKey identity.Key,
	location string,
) *ExpressionParseIssue {
	if spec == nil || spec.Specification == "" || spec.ParseOk() {
		return nil
	}
	_, _, err := pf(spec.Specification)
	if err == nil {
		return nil
	}
	return &ExpressionParseIssue{
		ClassKey: classKey,
		Location: location,
		Message:  err.Error(),
		SpecText: spec.Specification,
	}
}

func parameterNameSet(params []string) map[string]bool {
	if len(params) == 0 {
		return nil
	}
	names := make(map[string]bool, len(params))
	for _, p := range params {
		names[p] = true
	}
	return names
}
