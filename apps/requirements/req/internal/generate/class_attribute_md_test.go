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

func TestAttributeCommentsInvariantsRendersDetailsAndInvariantSpec(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	attrKey := helper.Must(identity.NewAttributeKey(classKey, "social_only"))
	invKey := helper.Must(identity.NewAttributeInvariantKey(attrKey, "0"))

	attr := helper.Must(model_class.NewAttribute(
		attrKey,
		model_class.AttributeDetails{
			Name:    "Is Social Only",
			Details: "A social-only jurisdiction only allows social currencies.",
		},
		"enum of TRUE, FALSE",
		nil,
		false,
		model_class.AttributeAnnotations{},
	))
	attr.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			invKey,
			model_logic.LogicTypeAssessment,
			"If no jurisdiction then we must be social.",
			"",
			logic_spec.ExpressionSpec{
				Notation:      model_logic.NotationTLAPlus,
				Specification: `IF self.jurisdiction_code = NULL THEN self.social_only = TRUE ELSE TRUE`,
			},
			nil,
		),
	})

	jurisdictionCodeAttr := helper.Must(model_class.NewAttribute(
		helper.Must(identity.NewAttributeKey(classKey, "jurisdiction_code")),
		model_class.AttributeDetails{Name: "Jurisdiction Code", Details: ""},
		"unconstrained",
		nil,
		true,
		model_class.AttributeAnnotations{},
	))

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})
	class.SetAttributes([]model_class.Attribute{jurisdictionCodeAttr, attr})

	got := attributeCommentsInvariantsForClass(class, attr)
	require.Contains(t, got, "A social-only jurisdiction only allows social currencies.")
	require.Contains(t, got, "If no jurisdiction then we must be social.<br>**IF self.JurisdictionCode = NULL THEN self.IsSocialOnly = TRUE ELSE TRUE**")
	require.NotContains(t, got, "- If no jurisdiction")
	require.Contains(t, got, "<br><br>")
}

func TestAttributeCommentsInvariantsSeparatesMultipleInvariants(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	attrKey := helper.Must(identity.NewAttributeKey(classKey, "jurisdiction_code"))
	inv0 := helper.Must(identity.NewAttributeInvariantKey(attrKey, "0"))
	inv1 := helper.Must(identity.NewAttributeInvariantKey(attrKey, "1"))

	attr := helper.Must(model_class.NewAttribute(
		attrKey,
		model_class.AttributeDetails{Name: "Jurisdiction Code", Details: ""},
		"unconstrained",
		nil,
		true,
		model_class.AttributeAnnotations{},
	))
	attr.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			inv0,
			model_logic.LogicTypeAssessment,
			"Allowed jurisdiction.",
			"",
			logic_spec.ExpressionSpec{
				Notation:      model_logic.NotationTLAPlus,
				Specification: `IF self.jurisdiction_code = NULL THEN TRUE ELSE self.jurisdiction_code \\in _JurisdictionCodes`,
			},
			nil,
		),
		model_logic.NewLogic(
			inv1,
			model_logic.LogicTypeAssessment,
			"Null code implies social only.",
			"",
			logic_spec.ExpressionSpec{
				Notation:      model_logic.NotationTLAPlus,
				Specification: `IF self.jurisdiction_code = NULL THEN self.social_only = TRUE ELSE TRUE`,
			},
			nil,
		),
	})

	socialOnlyAttr := helper.Must(model_class.NewAttribute(
		helper.Must(identity.NewAttributeKey(classKey, "social_only")),
		model_class.AttributeDetails{Name: "Is Social Only", Details: ""},
		"enum of TRUE, FALSE",
		nil,
		false,
		model_class.AttributeAnnotations{},
	))

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})
	class.SetAttributes([]model_class.Attribute{attr, socialOnlyAttr})

	got := attributeCommentsInvariantsForClass(class, attr)
	require.Contains(t, got, "Allowed jurisdiction.<br>**IF self.JurisdictionCode = NULL THEN TRUE ELSE self.JurisdictionCode")
	require.Contains(t, got, "_JurisdictionCodes**")
	require.Contains(t, got, "Null code implies social only.<br>**IF self.JurisdictionCode = NULL THEN self.IsSocialOnly = TRUE ELSE TRUE**")
	require.Contains(t, got, "_JurisdictionCodes**<br><br>Null code")
	require.NotContains(t, got, "- Allowed")
}

func TestClassMarkdownUsesCommentsInvariantsColumn(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	attrKey := helper.Must(identity.NewAttributeKey(classKey, "social_only"))
	invKey := helper.Must(identity.NewAttributeInvariantKey(attrKey, "0"))

	attr := helper.Must(model_class.NewAttribute(
		attrKey,
		model_class.AttributeDetails{Name: "Is Social Only", Details: "Social only flag."},
		"enum of TRUE, FALSE",
		nil,
		false,
		model_class.AttributeAnnotations{},
	))
	attr.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(
			invKey,
			model_logic.LogicTypeAssessment,
			"Must be social when code is null.",
			"",
			logic_spec.ExpressionSpec{
				Notation:      model_logic.NotationTLAPlus,
				Specification: `IF self.jurisdiction_code = NULL THEN self.social_only = TRUE ELSE TRUE`,
			},
			nil,
		),
	})

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})
	class.SetAttributes([]model_class.Attribute{attr})

	subdomain := model_domain.NewSubdomain(subdomainKey, "Wallet", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{classKey: class}
	domain := model_domain.NewDomain(domainKey, "Finance", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}

	reqs := req_flat.NewRequirements(model)
	contents, err := generateClassMdContents(reqs, class, nil, "", "")
	require.NoError(t, err)
	require.Contains(t, contents, "| Comments / Invariants |")
	require.Contains(t, contents, "Must be social when code is null.")
	require.NotContains(t, contents, "### Is Social Only Invariants")
}
