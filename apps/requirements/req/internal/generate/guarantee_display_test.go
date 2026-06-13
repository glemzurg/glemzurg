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

func TestActionGuaranteeDisplayDescriptionInMarkdown(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "add"))
	nameAttrKey := helper.Must(identity.NewAttributeKey(classKey, "name"))
	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionKey, "0"))

	nameAttr, err := model_class.NewAttribute(nameAttrKey, "Display Name", "", "unconstrained", nil, false, model_class.AttributeAnnotations{})
	require.NoError(t, err)

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})
	class.SetAttributes(map[identity.Key]model_class.Attribute{nameAttrKey: nameAttr})
	class.SetActions(map[identity.Key]model_state.Action{
		actionKey: model_state.NewAction(
			actionKey,
			"Add",
			"Add a jurisdiction",
			nil,
			[]model_logic.Logic{{
				Key:    guaranteeKey,
				Type:   model_logic.LogicTypeStateChange,
				Target: "name",
				Spec:   logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "Name"},
			}},
			nil,
			nil,
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
	require.Contains(t, contents, "- Set Display Name")
	require.Contains(t, contents, "name' = Name")
}
