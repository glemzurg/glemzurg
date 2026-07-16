package surface

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type MemberAvailabilitySuite struct {
	suite.Suite
}

func TestMemberAvailabilitySuite(t *testing.T) {
	suite.Run(t, new(MemberAvailabilitySuite))
}

func (s *MemberAvailabilitySuite) TestCollectUnavailableMembers_DerivedOverAssociationClass() {
	// Order has Balance derived via Adjusts.AccountBalanceChange; ABC class out of surface.
	txKey := mustKey("domain/d/subdomain/s/class/transaction")
	acctKey := mustKey("domain/d/subdomain/s/class/account")
	abcKey := mustKey("domain/d/subdomain/s/class/account_balance_change")
	assocKey := testAssocKey(txKey, acctKey, "adjusts")

	txClass := model_class.NewClass(txKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Transaction", Details: "", UnfinishedNotes: "", UmlComment: ""})
	txState := mustKey("domain/d/subdomain/s/class/transaction/state/recorded")
	txClass.States = map[identity.Key]model_state.State{
		txState: model_state.NewState(txState, "Recorded", "", ""),
	}
	// Account: Balance derivation navigates reverse Adjusts + AccountBalanceChange.
	acctClass := model_class.NewClass(acctKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account", Details: "", UnfinishedNotes: "", UmlComment: ""})
	acctState := mustKey("domain/d/subdomain/s/class/account/state/exists")
	acctClass.States = map[identity.Key]model_state.State{
		acctState: model_state.NewState(acctState, "Exists", "", ""),
	}
	balanceKey := helper.Must(identity.NewAttributeKey(acctKey, "balance"))
	deriv := model_logic.NewLogic(
		mustKey("invariant/balance_deriv"),
		model_logic.LogicTypeValue,
		"Net of ABC",
		"",
		parsedSpecWithAssoc(`self._Adjusts.AccountBalanceChange`, map[string]identity.Key{
			"_Adjusts": assocKey,
		}),
		nil,
	)
	balanceAttr := helper.Must(model_class.NewAttribute(
		balanceKey,
		model_class.AttributeDetails{Name: "Balance", Details: ""},
		"",
		&deriv,
		false,
		model_class.AttributeAnnotations{},
	))
	acctClass.Attributes = []model_class.Attribute{balanceAttr}

	abcClass := model_class.NewClass(abcKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account Balance Change", Details: "", UnfinishedNotes: "", UmlComment: ""})
	abcState := mustKey("domain/d/subdomain/s/class/account_balance_change/state/recorded")
	abcClass.States = map[identity.Key]model_state.State{
		abcState: model_state.NewState(abcState, "Recorded", "", ""),
	}

	hostAssoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Adjusts", Details: ""},
		model_class.AssociationEnd{ClassKey: txKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationEnd{ClassKey: acctKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1..many"))},
		model_class.AssociationOptions{AssociationClassKey: &abcKey},
	)

	sub := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	sub.Classes = map[identity.Key]model_class.Class{
		txKey: txClass, acctKey: acctClass, abcKey: abcClass,
	}
	sub.ClassAssociations = map[identity.Key]model_class.Association{assocKey: hostAssoc}
	dom := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	dom.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: sub}
	model := core.NewModel("m", core.ModelDetails{Name: "m", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: dom}

	// Surface includes Transaction + Account only — not ABC.
	spec := &SurfaceSpecification{IncludeClasses: []identity.Key{txKey, acctKey}}
	resolved, err := Resolve(spec, &model)
	s.Require().NoError(err)
	_, err = BuildFilteredModel(&model, resolved)
	s.Require().NoError(err)

	s.Require().NotEmpty(resolved.UnavailableMembers)
	var found bool
	for _, m := range resolved.UnavailableMembers {
		if m.Kind == MemberDerived && m.MemberName == "Balance" {
			found = true
			s.Contains(m.MissingClasses, "Account Balance Change")
			s.Contains(m.Reason(), "Account Balance Change")
		}
	}
	s.True(found, "Balance should be surface-unavailable when ABC is out of scope")
}

func (s *MemberAvailabilitySuite) TestCollectUnavailableMembers_InScopePeerKeepsDerived() {
	defKey := mustKey("domain/d/subdomain/s/class/account_definition")
	acctKey := mustKey("domain/d/subdomain/s/class/account")
	assocKey := testAssocKey(defKey, acctKey, "defines")

	defClass := model_class.NewClass(defKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account Definition", Details: "", UnfinishedNotes: "", UmlComment: ""})
	defState := mustKey("domain/d/subdomain/s/class/account_definition/state/active")
	defClass.States = map[identity.Key]model_state.State{
		defState: model_state.NewState(defState, "Active", "", ""),
	}

	acctClass := model_class.NewClass(acctKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account", Details: "", UnfinishedNotes: "", UmlComment: ""})
	acctState := mustKey("domain/d/subdomain/s/class/account/state/exists")
	acctClass.States = map[identity.Key]model_state.State{
		acctState: model_state.NewState(acctState, "Exists", "", ""),
	}
	nameKey := helper.Must(identity.NewAttributeKey(acctKey, "name"))
	deriv := model_logic.NewLogic(
		mustKey("invariant/name_deriv"),
		model_logic.LogicTypeValue,
		"From definition",
		"",
		parsedSpecWithAssoc(`(CHOOSE d \in self._Defines : TRUE).name`, map[string]identity.Key{
			"_Defines": assocKey,
		}),
		nil,
	)
	nameAttr := helper.Must(model_class.NewAttribute(
		nameKey,
		model_class.AttributeDetails{Name: "Name", Details: ""},
		"",
		&deriv,
		false,
		model_class.AttributeAnnotations{},
	))
	acctClass.Attributes = []model_class.Attribute{nameAttr}

	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Defines", Details: ""},
		model_class.AssociationEnd{ClassKey: defKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))},
		model_class.AssociationEnd{ClassKey: acctKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationOptions{},
	)

	sub := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	sub.Classes = map[identity.Key]model_class.Class{defKey: defClass, acctKey: acctClass}
	sub.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}
	dom := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	dom.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: sub}
	model := core.NewModel("m", core.ModelDetails{Name: "m", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: dom}

	spec := &SurfaceSpecification{IncludeClasses: []identity.Key{defKey, acctKey}}
	resolved, err := Resolve(spec, &model)
	s.Require().NoError(err)
	_, err = BuildFilteredModel(&model, resolved)
	s.Require().NoError(err)

	for _, m := range resolved.UnavailableMembers {
		s.NotEqual("Name", m.MemberName, "Name should stay available when Account Definition is in scope")
	}
}

func parsedSpecWithAssoc(tla string, assocNames map[string]identity.Key) logic_spec.ExpressionSpec {
	ctx := &convert.LowerContext{AssociationNames: assocNames}
	pf := convert.NewExpressionParseFunc(ctx)
	return helper.Must(logic_spec.NewExpressionSpec("tla_plus", tla, pf))
}

func TestUnavailableMemberReason(t *testing.T) {
	m := UnavailableMember{
		Kind:           MemberDerived,
		MemberName:     "Balance",
		MissingClasses: []string{"Account Balance Change"},
	}
	require.Contains(t, m.Reason(), "Balance")
	require.Contains(t, m.Reason(), "Account Balance Change")
}

func (s *MemberAvailabilitySuite) TestClassInvariant_ExcludedWhenPeerClassOutOfSurface() {
	defKey := mustKey("domain/d/subdomain/s/class/account_definition")
	acctKey := mustKey("domain/d/subdomain/s/class/account")
	assocKey := testAssocKey(defKey, acctKey, "defines")

	defClass := model_class.NewClass(defKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account Definition", Details: "", UnfinishedNotes: "", UmlComment: ""})
	defState := mustKey("domain/d/subdomain/s/class/account_definition/state/active")
	defClass.States = map[identity.Key]model_state.State{
		defState: model_state.NewState(defState, "Active", "", ""),
	}

	acctClass := model_class.NewClass(acctKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account", Details: "", UnfinishedNotes: "", UmlComment: ""})
	acctState := mustKey("domain/d/subdomain/s/class/account/state/active")
	acctClass.States = map[identity.Key]model_state.State{
		acctState: model_state.NewState(acctState, "Active", "", ""),
	}
	inv := model_logic.NewLogic(
		mustKey("invariant/state_match"),
		model_logic.LogicTypeAssessment,
		"Account lifecycle state matches the definition",
		"",
		parsedSpecWithAssoc(`(CHOOSE d \in self._Defines : TRUE)._state = self._state`, map[string]identity.Key{
			"_Defines": assocKey,
		}),
		nil,
	)
	acctClass.Invariants = []model_logic.Logic{inv}

	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Defines", Details: ""},
		model_class.AssociationEnd{ClassKey: defKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1"))},
		model_class.AssociationEnd{ClassKey: acctKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationOptions{},
	)

	sub := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	sub.Classes = map[identity.Key]model_class.Class{defKey: defClass, acctKey: acctClass}
	sub.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}
	dom := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	dom.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: sub}
	model := core.NewModel("m", core.ModelDetails{Name: "m", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: dom}

	// Surface lists Account only — definition out of scope drops the invariant.
	spec := &SurfaceSpecification{IncludeClasses: []identity.Key{acctKey}}
	resolved, err := Resolve(spec, &model)
	s.Require().NoError(err)
	filtered, err := BuildFilteredModel(&model, resolved)
	s.Require().NoError(err)

	var filteredAcct model_class.Class
	for _, d := range filtered.Domains {
		for _, sd := range d.Subdomains {
			if c, ok := sd.Classes[acctKey]; ok {
				filteredAcct = c
			}
		}
	}
	s.Empty(filteredAcct.Invariants, "state-match invariant requires Account Definition on surface")
	foundWarn := false
	for _, w := range resolved.Warnings {
		if contains(w, "invariant excluded") {
			foundWarn = true
			break
		}
	}
	s.True(foundWarn)

	// Both classes in scope keeps the invariant.
	specBoth := &SurfaceSpecification{IncludeClasses: []identity.Key{acctKey, defKey}}
	resolvedBoth, err := Resolve(specBoth, &model)
	s.Require().NoError(err)
	filteredBoth, err := BuildFilteredModel(&model, resolvedBoth)
	s.Require().NoError(err)
	var kept model_class.Class
	for _, d := range filteredBoth.Domains {
		for _, sd := range d.Subdomains {
			if c, ok := sd.Classes[acctKey]; ok {
				kept = c
			}
		}
	}
	s.Len(kept.Invariants, 1)
}
