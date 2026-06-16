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

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Widget"})
	class.SetActions(map[identity.Key]model_state.Action{
		actionKey: model_state.NewAction(
			actionKey,
			"Adjust",
			"Adjust the widget",
			nil,
			nil,
			nil,
			[]model_state.Parameter{amountParam, labelParam},
		),
	})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", "Test", "", "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	reqs := req_flat.NewRequirements(model)
	contents, err := generateClassMdContents(reqs, class, "", "")
	require.NoError(t, err)
	require.Contains(t, contents, "- *Amount.* __unconstrained__ (STRING)")
	require.Contains(t, contents, "- *Label.* __unconstrained__")
	require.NotContains(t, contents, "- *Label.* __unconstrained__ (")
}

func TestUnconstrainedAttributeTypeSpecDisplayInMarkdown(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "widget"))
	attrKey := helper.Must(identity.NewAttributeKey(classKey, "note"))

	attr, err := model_class.NewAttribute(attrKey, "Note", "A note field.", "unconstrained", nil, false, model_class.AttributeAnnotations{})
	require.NoError(t, err)
	require.NotNil(t, attr.DataType)

	typeSpec, err := logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "STRING", nil)
	require.NoError(t, err)
	attr.DataType.TypeSpec = &typeSpec

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Widget"})
	class.SetAttributes(map[identity.Key]model_class.Attribute{attrKey: attr})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", "Test", "", "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	reqs := req_flat.NewRequirements(model)
	contents, err := generateClassMdContents(reqs, class, "", "")
	require.NoError(t, err)
	require.Contains(t, contents, "| Note | __unconstrained__ | false | STRING | A note field. |")
}
