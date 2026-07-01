package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestParameterTypeSpecDisplayInMarkdown(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "widget"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "adjust"))

	amountParam, err := model_state.NewParameter(actionKey, "Amount", "unconstrained", false)
	require.NoError(t, err)
	require.NotNil(t, amountParam.DataType)

	typeSpec, err := logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "STRING", nil)
	require.NoError(t, err)
	amountParam.DataType.TypeSpec = &typeSpec

	labelParam, err := model_state.NewParameter(actionKey, "Label", "unconstrained", false)
	require.NoError(t, err)

	optionalParam, err := model_state.NewParameter(actionKey, "Note", "unconstrained", true)
	require.NoError(t, err)

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Widget"})
	class.SetActions(map[identity.Key]model_state.Action{
		actionKey: model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Adjust", Details: "Adjust the widget"}, nil, nil, nil, []model_state.Parameter{amountParam, labelParam, optionalParam}),
	})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	reqs := req_flat.NewRequirements(model)
	contents, err := generateClassMdContents(reqs, class, "", "")
	require.NoError(t, err)
	require.Contains(t, contents, "- *Amount.* __unconstrained__ (STRING)")
	require.Contains(t, contents, "- *Label.* __unconstrained__")
	require.NotContains(t, contents, "- *Label.* __unconstrained__ (")
	require.Contains(t, contents, "- *Note.* __unconstrained__ (nullable)")
}

func TestParameterSimulationDisplayInMarkdown(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "transaction"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "initialize"))
	paramKey := helper.Must(identity.NewParameterKey(actionKey, "amounts"))
	reqKey := helper.Must(identity.NewParameterSimulationRequireKey(paramKey, "0"))
	specKey := helper.Must(identity.NewParameterSimulationSpecKey(paramKey))

	amountsParam, err := model_state.NewParameter(actionKey, "Amounts", "unordered of unconstrained", false)
	require.NoError(t, err)
	require.NotNil(t, amountsParam.DataType)
	typeSpec, err := logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "[account: account, amount: Int]", nil)
	require.NoError(t, err)
	amountsParam.DataType.TypeSpec = &typeSpec

	reqLogic := model_logic.NewLogic(
		reqKey,
		model_logic.LogicTypeAssessment,
		"At least one account must exist before creating a transaction.",
		"",
		logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "Account /= {}"},
		nil,
	)
	specLogic := model_logic.NewLogic(
		specKey,
		model_logic.LogicTypeValue,
		"",
		"",
		logic_spec.ExpressionSpec{
			Notation:      model_logic.NotationTLAPlus,
			Specification: `LET ac == CHOOSE a \in Account : TRUE IN {[account |-> ac, amount |-> 100]}`,
		},
		nil,
	)
	amountsParam.SetSimulation(&model_state.ParameterSimulation{
		Details:       "Sample one balance change on an existing account using penny-grid amounts.",
		Requires:      []model_logic.Logic{reqLogic},
		Specification: &specLogic,
	})

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Transaction"})
	class.SetActions(map[identity.Key]model_state.Action{
		actionKey: model_state.NewAction(
			actionKey,
			model_state.ActionDetails{Name: "Initialize", Details: "Create a transaction."},
			nil,
			nil,
			nil,
			[]model_state.Parameter{amountsParam},
		),
	})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	reqs := req_flat.NewRequirements(model)
	contents, err := generateClassMdContents(reqs, class, "", "")
	require.NoError(t, err)
	require.Contains(t, contents, "- *Amounts.*")
	require.Contains(t, contents, "    - Simulation:")
	require.Contains(t, contents, "Sample one balance change on an existing account using penny-grid amounts.")
	require.Contains(t, contents, "        - Requires:")
	require.Contains(t, contents, "At least one account must exist before creating a transaction.")
	require.Contains(t, contents, "**Account /= {}**")
	require.Contains(t, contents, "        - Specification:\n            - **LET ac == CHOOSE a \\in Account : TRUE IN {[account |-> ac, amount |-> 100]}**")
}

func TestNullableQueryParameterDisplayInMarkdown(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "widget"))
	queryKey := helper.Must(identity.NewQueryKey(classKey, "lookup"))

	requiredParam, err := model_state.NewParameter(queryKey, "Id", "Nat", false)
	require.NoError(t, err)
	optionalParam, err := model_state.NewParameter(queryKey, "Format", "unconstrained", true)
	require.NoError(t, err)

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Widget"})
	class.SetQueries(map[identity.Key]model_state.Query{
		queryKey: model_state.NewQuery(
			queryKey,
			"Lookup",
			"Look up a widget",
			nil,
			nil,
			[]model_state.Parameter{requiredParam, optionalParam},
		),
	})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	reqs := req_flat.NewRequirements(model)
	contents, err := generateClassMdContents(reqs, class, "", "")
	require.NoError(t, err)
	require.Contains(t, contents, "- *Id.* _(unparsed)_ Nat")
	require.NotContains(t, contents, "- *Id.* _(unparsed)_ Nat (nullable)")
	require.Contains(t, contents, "- *Format.* __unconstrained__ (nullable)")
}

func TestUnconstrainedAttributeTypeSpecDisplayInMarkdown(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "widget"))
	attrKey := helper.Must(identity.NewAttributeKey(classKey, "note"))

	attr, err := model_class.NewAttribute(attrKey, model_class.AttributeDetails{Name: "Note", Details: "A note field."}, "unconstrained", nil, false, model_class.AttributeAnnotations{})
	require.NoError(t, err)
	require.NotNil(t, attr.DataType)

	typeSpec, err := logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "STRING", nil)
	require.NoError(t, err)
	attr.DataType.TypeSpec = &typeSpec

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Widget"})
	class.SetAttributes([]model_class.Attribute{attr})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	reqs := req_flat.NewRequirements(model)
	contents, err := generateClassMdContents(reqs, class, "", "")
	require.NoError(t, err)
	require.Contains(t, contents, "| Note | __unconstrained__ | false | STRING | A note field. |")
}
