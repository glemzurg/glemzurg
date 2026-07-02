package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestApplySelfAttributeTLADisplayNames(t *testing.T) {
	class := model_class.NewClass(
		helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("finance")), "wallet")), "account_balance_change")),
		model_class.ClassLinks{},
		model_class.ClassDetails{Name: "Account Balance Change"},
	)
	class.SetAttributes([]model_class.Attribute{
		helper.Must(model_class.NewAttribute(
			helper.Must(identity.NewAttributeKey(class.Key, "amount")),
			model_class.AttributeDetails{Name: "Amount", Details: ""},
			"",
			nil,
			false,
			model_class.AttributeAnnotations{},
		)),
	})

	got := applySelfAttributeTLADisplayNames(class, "self.amount /= 0")
	require.Equal(t, "self.Amount /= 0", got)

	got = applySelfAttributeTLADisplayNames(class, "self.amount' = self.amount + 1")
	require.Equal(t, "self.Amount' = self.Amount + 1", got)
}

func TestGenerateClassMdContentsRendersSelfAmountAsDisplayTLAName(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "account_balance_change"))
	attrKey := helper.Must(identity.NewAttributeKey(classKey, "amount"))
	invKey := helper.Must(identity.NewAttributeInvariantKey(attrKey, "0"))

	attr := helper.Must(model_class.NewAttribute(
		attrKey,
		model_class.AttributeDetails{Name: "Amount", Details: "How much the balance moved."},
		"[unconstrained..unconstrained] at 1 penny",
		nil,
		false,
		model_class.AttributeAnnotations{},
	))
	attr.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			invKey,
			model_logic.LogicTypeAssessment,
			"Cannot be zero.",
			"",
			logic_spec.ExpressionSpec{
				Notation:      model_logic.NotationTLAPlus,
				Specification: "self.amount /= 0",
			},
			nil,
		),
	})

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account Balance Change"})
	class.SetAttributes([]model_class.Attribute{attr})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	contents, err := generateClassMdContents(req_flat.NewRequirements(model), class, "", "")
	require.NoError(t, err)
	require.Contains(t, contents, "**self.Amount /= 0**")
	require.NotContains(t, contents, "self.amount")
}
