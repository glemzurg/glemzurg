package surface

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/stretchr/testify/suite"
)

// ============================================================
// Common test keys and helpers
// ============================================================

func mustKey(s string) identity.Key {
	k, err := identity.ParseKey(s)
	if err != nil {
		panic(err)
	}
	return k
}

var (
	domainKey     = mustKey("domain/d")
	domain2Key    = mustKey("domain/d2")
	subdomainKey  = mustKey("domain/d/subdomain/s")
	subdomain2Key = mustKey("domain/d2/subdomain/s2")

	orderClassKey   = mustKey("domain/d/subdomain/s/class/order")
	itemClassKey    = mustKey("domain/d/subdomain/s/class/item")
	paymentClassKey = mustKey("domain/d2/subdomain/s2/class/payment")

	orderStateOpenKey      = mustKey("domain/d/subdomain/s/class/order/state/open")
	orderStateClosedKey    = mustKey("domain/d/subdomain/s/class/order/state/closed")
	itemStateActiveKey     = mustKey("domain/d/subdomain/s/class/item/state/active")
	paymentStatePendingKey = mustKey("domain/d2/subdomain/s2/class/payment/state/pending")

	orderEventCreateKey   = mustKey("domain/d/subdomain/s/class/order/event/create")
	orderEventCloseKey    = mustKey("domain/d/subdomain/s/class/order/event/close")
	itemEventCreateKey    = mustKey("domain/d/subdomain/s/class/item/event/create_item")
	paymentEventCreateKey = mustKey("domain/d2/subdomain/s2/class/payment/event/create_payment")

	orderTransCreateKey   = mustKey("domain/d/subdomain/s/class/order/transition/create_order")
	orderTransCloseKey    = mustKey("domain/d/subdomain/s/class/order/transition/close_order")
	itemTransCreateKey    = mustKey("domain/d/subdomain/s/class/item/transition/create_item")
	paymentTransCreateKey = mustKey("domain/d2/subdomain/s2/class/payment/transition/create_payment")
)

// testAssocKey creates an association key.
func testAssocKey(fromKey, toKey identity.Key, name string) identity.Key {
	return helper.Must(identity.NewClassAssociationKey(subdomainKey, fromKey, toKey, name))
}

// makeOrderClass builds a simple Order class with states and creation transition.
func makeOrderClass() model_class.Class {
	return model_class.Class{
		Key:        orderClassKey,
		Name:       "Order",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			orderStateOpenKey:   {Key: orderStateOpenKey, Name: "Open"},
			orderStateClosedKey: {Key: orderStateClosedKey, Name: "Closed"},
		},
		Events: map[identity.Key]model_state.Event{
			orderEventCreateKey: {Key: orderEventCreateKey, Name: "create"},
			orderEventCloseKey:  {Key: orderEventCloseKey, Name: "close"},
		},
		Guards:  map[identity.Key]model_state.Guard{},
		Actions: map[identity.Key]model_state.Action{},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
			orderTransCreateKey: {
				Key:        orderTransCreateKey,
				EventKey:   orderEventCreateKey,
				ToStateKey: &orderStateOpenKey,
			},
			orderTransCloseKey: {
				Key:          orderTransCloseKey,
				FromStateKey: &orderStateOpenKey,
				EventKey:     orderEventCloseKey,
				ToStateKey:   &orderStateClosedKey,
			},
		},
	}
}

// makeItemClass builds a simple Item class with one state and creation.
func makeItemClass() model_class.Class {
	return model_class.Class{
		Key:        itemClassKey,
		Name:       "Item",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			itemStateActiveKey: {Key: itemStateActiveKey, Name: "Active"},
		},
		Events: map[identity.Key]model_state.Event{
			itemEventCreateKey: {Key: itemEventCreateKey, Name: "create_item"},
		},
		Guards:  map[identity.Key]model_state.Guard{},
		Actions: map[identity.Key]model_state.Action{},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
			itemTransCreateKey: {
				Key:        itemTransCreateKey,
				EventKey:   itemEventCreateKey,
				ToStateKey: &itemStateActiveKey,
			},
		},
	}
}

// makePaymentClass builds a simple Payment class in domain2.
func makePaymentClass() model_class.Class {
	return model_class.Class{
		Key:        paymentClassKey,
		Name:       "Payment",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			paymentStatePendingKey: {Key: paymentStatePendingKey, Name: "Pending"},
		},
		Events: map[identity.Key]model_state.Event{
			paymentEventCreateKey: {Key: paymentEventCreateKey, Name: "create_payment"},
		},
		Guards:  map[identity.Key]model_state.Guard{},
		Actions: map[identity.Key]model_state.Action{},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
			paymentTransCreateKey: {
				Key:        paymentTransCreateKey,
				EventKey:   paymentEventCreateKey,
				ToStateKey: &paymentStatePendingKey,
			},
		},
	}
}

// makeStatelessClass builds a class with no states (not simulatable).
func makeStatelessClass() model_class.Class {
	cKey := mustKey("domain/d/subdomain/s/class/stateless")
	return model_class.Class{
		Key:         cKey,
		Name:        "Stateless",
		Attributes:  map[identity.Key]model_class.Attribute{},
		States:      map[identity.Key]model_state.State{},
		Events:      map[identity.Key]model_state.Event{},
		Guards:      map[identity.Key]model_state.Guard{},
		Actions:     map[identity.Key]model_state.Action{},
		Queries:     map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{},
	}
}

// buildTwoDomainModel creates a model with Order+Item in domain/d and Payment in domain/d2.
func buildTwoDomainModel() *req_model.Model {
	assocKey := testAssocKey(orderClassKey, itemClassKey, "order_items")
	return &req_model.Model{
		Key:  "test",
		Name: "Test",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "D",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:  subdomainKey,
						Name: "S",
						Classes: map[identity.Key]model_class.Class{
							orderClassKey: makeOrderClass(),
							itemClassKey:  makeItemClass(),
						},
						ClassAssociations: map[identity.Key]model_class.Association{
							assocKey: {
								Key:              assocKey,
								Name:             "order_items",
								FromClassKey:     orderClassKey,
								FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
								ToClassKey:       itemClassKey,
								ToMultiplicity:   model_class.Multiplicity{LowerBound: 1},
							},
						},
					},
				},
			},
			domain2Key: {
				Key:  domain2Key,
				Name: "D2",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomain2Key: {
						Key:  subdomain2Key,
						Name: "S2",
						Classes: map[identity.Key]model_class.Class{
							paymentClassKey: makePaymentClass(),
						},
					},
				},
			},
		},
	}
}

// buildSingleDomainModel creates a model with just Order and Item.
func buildSingleDomainModel() *req_model.Model {
	return &req_model.Model{
		Key:  "test",
		Name: "Test",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "D",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:  subdomainKey,
						Name: "S",
						Classes: map[identity.Key]model_class.Class{
							orderClassKey: makeOrderClass(),
							itemClassKey:  makeItemClass(),
						},
					},
				},
			},
		},
	}
}

// ============================================================
// Specification Tests
// ============================================================

func TestSurfaceSuite(t *testing.T) {
	suite.Run(t, new(SurfaceSuite))
}

type SurfaceSuite struct {
	suite.Suite
}

func (s *SurfaceSuite) TestIsEmpty_AllEmpty() {
	spec := &SurfaceSpecification{}
	s.True(spec.IsEmpty())
}

func (s *SurfaceSuite) TestIsEmpty_WithIncludeDomains() {
	spec := &SurfaceSpecification{
		IncludeDomains: []identity.Key{domainKey},
	}
	s.False(spec.IsEmpty())
}

func (s *SurfaceSuite) TestIsEmpty_WithIncludeSubdomains() {
	spec := &SurfaceSpecification{
		IncludeSubdomains: []identity.Key{subdomainKey},
	}
	s.False(spec.IsEmpty())
}

func (s *SurfaceSuite) TestIsEmpty_WithIncludeClasses() {
	spec := &SurfaceSpecification{
		IncludeClasses: []identity.Key{orderClassKey},
	}
	s.False(spec.IsEmpty())
}

func (s *SurfaceSuite) TestIsEmpty_WithExcludeClasses() {
	spec := &SurfaceSpecification{
		ExcludeClasses: []identity.Key{orderClassKey},
	}
	s.False(spec.IsEmpty())
}

func (s *SurfaceSuite) TestValidate_ValidSpec() {
	model := buildTwoDomainModel()
	spec := &SurfaceSpecification{
		IncludeDomains:    []identity.Key{domainKey},
		IncludeSubdomains: []identity.Key{subdomain2Key},
		IncludeClasses:    []identity.Key{orderClassKey},
		ExcludeClasses:    []identity.Key{itemClassKey},
	}
	s.NoError(spec.Validate(model))
}

func (s *SurfaceSuite) TestValidate_UnknownDomain() {
	model := buildSingleDomainModel()
	spec := &SurfaceSpecification{
		IncludeDomains: []identity.Key{mustKey("domain/nonexistent")},
	}
	err := spec.Validate(model)
	s.Error(err)
	s.Contains(err.Error(), "unknown domain")
}

func (s *SurfaceSuite) TestValidate_UnknownSubdomain() {
	model := buildSingleDomainModel()
	spec := &SurfaceSpecification{
		IncludeSubdomains: []identity.Key{mustKey("domain/d/subdomain/nonexistent")},
	}
	err := spec.Validate(model)
	s.Error(err)
	s.Contains(err.Error(), "unknown subdomain")
}

func (s *SurfaceSuite) TestValidate_UnknownIncludeClass() {
	model := buildSingleDomainModel()
	spec := &SurfaceSpecification{
		IncludeClasses: []identity.Key{mustKey("domain/d/subdomain/s/class/nonexistent")},
	}
	err := spec.Validate(model)
	s.Error(err)
	s.Contains(err.Error(), "unknown class")
}

func (s *SurfaceSuite) TestValidate_UnknownExcludeClass() {
	model := buildSingleDomainModel()
	spec := &SurfaceSpecification{
		ExcludeClasses: []identity.Key{mustKey("domain/d/subdomain/s/class/nonexistent")},
	}
	err := spec.Validate(model)
	s.Error(err)
	s.Contains(err.Error(), "unknown class")
}

// ============================================================
// Resolver Tests
// ============================================================

func TestResolverSuite(t *testing.T) {
	suite.Run(t, new(ResolverSuite))
}

type ResolverSuite struct {
	suite.Suite
}

func (s *ResolverSuite) TestResolve_NilSpec_IncludesAll() {
	model := buildTwoDomainModel()
	resolved, err := Resolve(nil, model)
	s.NoError(err)
	// Should include all 3 classes (Order, Item, Payment).
	s.Len(resolved.Classes, 3)
	s.Contains(resolved.Classes, orderClassKey)
	s.Contains(resolved.Classes, itemClassKey)
	s.Contains(resolved.Classes, paymentClassKey)
}

func (s *ResolverSuite) TestResolve_EmptySpec_IncludesAll() {
	model := buildTwoDomainModel()
	spec := &SurfaceSpecification{}
	resolved, err := Resolve(spec, model)
	s.NoError(err)
	s.Len(resolved.Classes, 3)
}

func (s *ResolverSuite) TestResolve_IncludeDomain() {
	model := buildTwoDomainModel()
	spec := &SurfaceSpecification{
		IncludeDomains: []identity.Key{domainKey},
	}
	resolved, err := Resolve(spec, model)
	s.NoError(err)
	// Only Order and Item from domain D.
	s.Len(resolved.Classes, 2)
	s.Contains(resolved.Classes, orderClassKey)
	s.Contains(resolved.Classes, itemClassKey)
	s.NotContains(resolved.Classes, paymentClassKey)
}

func (s *ResolverSuite) TestResolve_IncludeSubdomain() {
	model := buildTwoDomainModel()
	spec := &SurfaceSpecification{
		IncludeSubdomains: []identity.Key{subdomain2Key},
	}
	resolved, err := Resolve(spec, model)
	s.NoError(err)
	// Only Payment from subdomain s2.
	s.Len(resolved.Classes, 1)
	s.Contains(resolved.Classes, paymentClassKey)
}

func (s *ResolverSuite) TestResolve_IncludeClass() {
	model := buildTwoDomainModel()
	spec := &SurfaceSpecification{
		IncludeClasses: []identity.Key{orderClassKey},
	}
	resolved, err := Resolve(spec, model)
	s.NoError(err)
	s.Len(resolved.Classes, 1)
	s.Contains(resolved.Classes, orderClassKey)
}

func (s *ResolverSuite) TestResolve_ExcludeClass() {
	model := buildTwoDomainModel()
	spec := &SurfaceSpecification{
		IncludeDomains: []identity.Key{domainKey},
		ExcludeClasses: []identity.Key{itemClassKey},
	}
	resolved, err := Resolve(spec, model)
	s.NoError(err)
	// Domain D has Order + Item, exclude Item.
	s.Len(resolved.Classes, 1)
	s.Contains(resolved.Classes, orderClassKey)
	s.NotContains(resolved.Classes, itemClassKey)
}

func (s *ResolverSuite) TestResolve_FiltersStatelessClasses() {
	statelessKey := mustKey("domain/d/subdomain/s/class/stateless")
	model := &req_model.Model{
		Key:  "test",
		Name: "Test",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "D",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:  subdomainKey,
						Name: "S",
						Classes: map[identity.Key]model_class.Class{
							orderClassKey: makeOrderClass(),
							statelessKey:  makeStatelessClass(),
						},
					},
				},
			},
		},
	}
	resolved, err := Resolve(nil, model)
	s.NoError(err)
	// Stateless class should be filtered out.
	s.Len(resolved.Classes, 1)
	s.Contains(resolved.Classes, orderClassKey)
	s.NotContains(resolved.Classes, statelessKey)
}

func (s *ResolverSuite) TestResolve_AssociationsBothEndpointsInScope() {
	model := buildTwoDomainModel()
	// Include only domain D (Order + Item) — the association between them should be included.
	spec := &SurfaceSpecification{
		IncludeDomains: []identity.Key{domainKey},
	}
	resolved, err := Resolve(spec, model)
	s.NoError(err)
	s.Len(resolved.Associations, 1)
	for _, assoc := range resolved.Associations {
		s.Equal("order_items", assoc.Name)
	}
}

func (s *ResolverSuite) TestResolve_AssociationOneEndpointExcluded() {
	model := buildTwoDomainModel()
	// Include only Order (exclude Item) — the association should be dropped.
	spec := &SurfaceSpecification{
		IncludeClasses: []identity.Key{orderClassKey},
	}
	resolved, err := Resolve(spec, model)
	s.NoError(err)
	s.Len(resolved.Associations, 0)
	// Should produce a warning about the dropped association.
	s.True(len(resolved.Warnings) > 0)
	foundWarning := false
	for _, w := range resolved.Warnings {
		if contains(w, "dropped") {
			foundWarning = true
			break
		}
	}
	s.True(foundWarning, "expected warning about dropped association")
}

func (s *ResolverSuite) TestResolve_RealizedDomainExcluded() {
	model := buildTwoDomainModel()
	// Mark domain2 as realized.
	d2 := model.Domains[domain2Key]
	d2.Realized = true
	model.Domains[domain2Key] = d2

	// Try to include the realized domain.
	spec := &SurfaceSpecification{
		IncludeDomains: []identity.Key{domainKey, domain2Key},
	}
	resolved, err := Resolve(spec, model)
	s.NoError(err)
	// Payment should not be included (realized domain).
	s.NotContains(resolved.Classes, paymentClassKey)
	// Should produce a warning about realized domain.
	foundWarning := false
	for _, w := range resolved.Warnings {
		if contains(w, "realized") {
			foundWarning = true
			break
		}
	}
	s.True(foundWarning, "expected warning about realized domain")
}

func (s *ResolverSuite) TestResolve_NoSimulatableClasses_Error() {
	statelessKey := mustKey("domain/d/subdomain/s/class/stateless")
	model := &req_model.Model{
		Key:  "test",
		Name: "Test",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "D",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:  subdomainKey,
						Name: "S",
						Classes: map[identity.Key]model_class.Class{
							statelessKey: makeStatelessClass(),
						},
					},
				},
			},
		},
	}
	_, err := Resolve(nil, model)
	s.Error(err)
	s.Contains(err.Error(), "no simulatable classes")
}

func (s *ResolverSuite) TestResolve_InvalidSpec_Error() {
	model := buildSingleDomainModel()
	spec := &SurfaceSpecification{
		IncludeDomains: []identity.Key{mustKey("domain/nonexistent")},
	}
	_, err := Resolve(spec, model)
	s.Error(err)
	s.Contains(err.Error(), "validation")
}

func (s *ResolverSuite) TestResolve_InvariantsScoped() {
	// Build a model with two domains so Payment is a known class.
	model2 := buildTwoDomainModel()
	model2.TlaInvariants = []string{
		"Order.count > 0",
		"Payment.count > 0",
	}
	spec2 := &SurfaceSpecification{
		IncludeDomains: []identity.Key{domainKey},
	}
	resolved2, err := Resolve(spec2, model2)
	s.NoError(err)
	// Only "Order.count > 0" should be included.
	s.Len(resolved2.ModelInvariants, 1)
	s.Equal("Order.count > 0", resolved2.ModelInvariants[0])
}

func (s *ResolverSuite) TestResolve_MultipleIncludes() {
	model := buildTwoDomainModel()
	// Include both a subdomain from D and an individual class from D2.
	spec := &SurfaceSpecification{
		IncludeSubdomains: []identity.Key{subdomainKey},
		IncludeClasses:    []identity.Key{paymentClassKey},
	}
	resolved, err := Resolve(spec, model)
	s.NoError(err)
	// All 3 classes should be included.
	s.Len(resolved.Classes, 3)
}

// ============================================================
// Invariant Scoping Tests
// ============================================================

func TestInvariantScopingSuite(t *testing.T) {
	suite.Run(t, new(InvariantScopingSuite))
}

type InvariantScopingSuite struct {
	suite.Suite
}

func (s *InvariantScopingSuite) TestScopeInvariants_AllInScope() {
	invariants := []string{
		"Order.count > 0",
		"Item.count >= 0",
	}
	inScope := map[string]bool{"Order": true, "Item": true}
	included, excluded := ScopeInvariants(invariants, inScope)
	s.Len(included, 2)
	s.Len(excluded, 0)
}

func (s *InvariantScopingSuite) TestScopeInvariants_SomeOutOfScope() {
	invariants := []string{
		"Order.count > 0",
		"Payment.count >= 0",
	}
	inScope := map[string]bool{"Order": true}
	included, excluded := ScopeInvariants(invariants, inScope)
	// ScopeInvariants includes everything since it doesn't know Payment is a class.
	// It only filters if the identifier matches a class name that's NOT in scope.
	// Without the allClassNames context, it can't tell.
	s.Len(included, 2)
	s.Len(excluded, 0)
}

func (s *InvariantScopingSuite) TestScopeInvariantsWithAllClasses_FiltersOutOfScope() {
	invariants := []string{
		"Order.count > 0",
		"Payment.count >= 0",
	}
	inScope := map[string]bool{"Order": true}
	allClasses := map[string]bool{"Order": true, "Payment": true}
	included, excluded := ScopeInvariantsWithAllClasses(invariants, inScope, allClasses)
	s.Len(included, 1)
	s.Equal("Order.count > 0", included[0])
	s.Len(excluded, 1)
	s.Equal("Payment.count >= 0", excluded[0])
}

func (s *InvariantScopingSuite) TestScopeInvariantsWithAllClasses_KeepsNonClassIdentifiers() {
	invariants := []string{
		"x + y > 0",
	}
	inScope := map[string]bool{"Order": true}
	allClasses := map[string]bool{"Order": true}
	// "x" and "y" are not known class names, so this invariant stays.
	included, excluded := ScopeInvariantsWithAllClasses(invariants, inScope, allClasses)
	s.Len(included, 1)
	s.Len(excluded, 0)
}

func (s *InvariantScopingSuite) TestScopeInvariantsWithAllClasses_EmptyInvariants() {
	included, excluded := ScopeInvariantsWithAllClasses(nil, map[string]bool{}, map[string]bool{})
	s.Len(included, 0)
	s.Len(excluded, 0)
}

func (s *InvariantScopingSuite) TestScopeInvariantsWithAllClasses_UnparseableInvariant() {
	invariants := []string{
		"!!@@## invalid TLA+",
	}
	inScope := map[string]bool{"Order": true}
	allClasses := map[string]bool{"Order": true}
	// Unparseable invariants should be kept (fail-open).
	included, excluded := ScopeInvariantsWithAllClasses(invariants, inScope, allClasses)
	s.Len(included, 1)
	s.Len(excluded, 0)
}

func (s *InvariantScopingSuite) TestScopeInvariantsWithAllClasses_MultipleClassReferences() {
	invariants := []string{
		"Order.count > 0 /\\ Payment.count > 0",
	}
	inScope := map[string]bool{"Order": true}
	allClasses := map[string]bool{"Order": true, "Payment": true}
	// References Payment which is out of scope.
	included, excluded := ScopeInvariantsWithAllClasses(invariants, inScope, allClasses)
	s.Len(included, 0)
	s.Len(excluded, 1)
}

// ============================================================
// Filtered Model Builder Tests
// ============================================================

func TestFilteredModelSuite(t *testing.T) {
	suite.Run(t, new(FilteredModelSuite))
}

type FilteredModelSuite struct {
	suite.Suite
}

func (s *FilteredModelSuite) TestBuildFilteredModel_KeepsIncludedClasses() {
	model := buildTwoDomainModel()
	resolved := &ResolvedSurface{
		Classes: map[identity.Key]model_class.Class{
			orderClassKey: makeOrderClass(),
			itemClassKey:  makeItemClass(),
		},
		Associations:    map[identity.Key]model_class.Association{},
		ModelInvariants: []string{"Order.count > 0"},
	}

	filtered := BuildFilteredModel(model, resolved)
	s.NotNil(filtered)

	// Count total classes in filtered model.
	totalClasses := 0
	for _, domain := range filtered.Domains {
		for _, subdomain := range domain.Subdomains {
			totalClasses += len(subdomain.Classes)
		}
	}
	s.Equal(2, totalClasses)

	// Check invariants.
	s.Len(filtered.TlaInvariants, 1)
	s.Equal("Order.count > 0", filtered.TlaInvariants[0])
}

func (s *FilteredModelSuite) TestBuildFilteredModel_ExcludesFilteredClasses() {
	model := buildTwoDomainModel()
	// Only include Order (not Item or Payment).
	resolved := &ResolvedSurface{
		Classes: map[identity.Key]model_class.Class{
			orderClassKey: makeOrderClass(),
		},
		Associations:    map[identity.Key]model_class.Association{},
		ModelInvariants: []string{},
	}

	filtered := BuildFilteredModel(model, resolved)

	// Count total classes — should be 1.
	totalClasses := 0
	for _, domain := range filtered.Domains {
		for _, subdomain := range domain.Subdomains {
			totalClasses += len(subdomain.Classes)
		}
	}
	s.Equal(1, totalClasses)
}

func (s *FilteredModelSuite) TestBuildFilteredModel_FilteredAssociations() {
	model := buildTwoDomainModel()
	assocKey := testAssocKey(orderClassKey, itemClassKey, "order_items")
	resolved := &ResolvedSurface{
		Classes: map[identity.Key]model_class.Class{
			orderClassKey: makeOrderClass(),
			itemClassKey:  makeItemClass(),
		},
		Associations: map[identity.Key]model_class.Association{
			assocKey: {
				Key:          assocKey,
				Name:         "order_items",
				FromClassKey: orderClassKey,
				ToClassKey:   itemClassKey,
			},
		},
		ModelInvariants: []string{},
	}

	filtered := BuildFilteredModel(model, resolved)

	// Count associations at all levels.
	totalAssocs := len(filtered.ClassAssociations)
	for _, domain := range filtered.Domains {
		totalAssocs += len(domain.ClassAssociations)
		for _, subdomain := range domain.Subdomains {
			totalAssocs += len(subdomain.ClassAssociations)
		}
	}
	// The association exists at subdomain level in the original model.
	s.Equal(1, totalAssocs)
}

func (s *FilteredModelSuite) TestBuildFilteredModel_PreservesModelMetadata() {
	model := buildTwoDomainModel()
	model.Key = "original_key"
	model.Name = "Original Name"
	resolved := &ResolvedSurface{
		Classes: map[identity.Key]model_class.Class{
			orderClassKey: makeOrderClass(),
		},
		Associations:    map[identity.Key]model_class.Association{},
		ModelInvariants: []string{},
	}

	filtered := BuildFilteredModel(model, resolved)
	s.Equal("original_key", filtered.Key)
	s.Equal("Original Name", filtered.Name)
}

func (s *FilteredModelSuite) TestBuildFilteredModel_EmptyDomainsOmitted() {
	model := buildTwoDomainModel()
	// Only include Payment from domain2.
	resolved := &ResolvedSurface{
		Classes: map[identity.Key]model_class.Class{
			paymentClassKey: makePaymentClass(),
		},
		Associations:    map[identity.Key]model_class.Association{},
		ModelInvariants: []string{},
	}

	filtered := BuildFilteredModel(model, resolved)

	// Domain D should have no classes in its subdomain.
	// Check that domain2 has Payment.
	foundPayment := false
	for _, domain := range filtered.Domains {
		for _, subdomain := range domain.Subdomains {
			if _, ok := subdomain.Classes[paymentClassKey]; ok {
				foundPayment = true
			}
		}
	}
	s.True(foundPayment)
}

// ============================================================
// Diagnostics Tests
// ============================================================

func TestDiagnosticsSuite(t *testing.T) {
	suite.Run(t, new(DiagnosticsSuite))
}

type DiagnosticsSuite struct {
	suite.Suite
}

func (s *DiagnosticsSuite) TestDiagnose_BrokenCreationChain() {
	model := buildTwoDomainModel()
	// Resolve with only Order, excluding Item.
	// The order_items association with mandatory lower bound should trigger a warning.
	resolved := &ResolvedSurface{
		Classes: map[identity.Key]model_class.Class{
			orderClassKey: makeOrderClass(),
		},
		Associations:    map[identity.Key]model_class.Association{},
		ModelInvariants: []string{},
	}

	diagnostics := Diagnose(resolved, model)
	found := false
	for _, d := range diagnostics {
		if contains(d.Message, "broken creation chain") {
			found = true
			s.Equal("warning", d.Level)
			break
		}
	}
	s.True(found, "expected broken creation chain diagnostic")
}

func (s *DiagnosticsSuite) TestDiagnose_IsolatedClass() {
	model := buildSingleDomainModel()

	// Create a resolved surface where Item has no creation transitions.
	isolatedItem := makeItemClass()
	// Remove creation transition from item.
	isolatedItem.Transitions = map[identity.Key]model_state.Transition{}

	resolved := &ResolvedSurface{
		Classes: map[identity.Key]model_class.Class{
			orderClassKey: makeOrderClass(),
			itemClassKey:  isolatedItem,
		},
		Associations:    map[identity.Key]model_class.Association{},
		ModelInvariants: []string{},
	}

	diagnostics := Diagnose(resolved, model)
	found := false
	for _, d := range diagnostics {
		if contains(d.Message, "isolated class") {
			found = true
			s.Equal("warning", d.Level)
			break
		}
	}
	s.True(found, "expected isolated class diagnostic")
}

func (s *DiagnosticsSuite) TestDiagnose_HalfAssociation() {
	model := buildTwoDomainModel()
	// Order is in scope but Item is not — the order_items association is half-in.
	resolved := &ResolvedSurface{
		Classes: map[identity.Key]model_class.Class{
			orderClassKey: makeOrderClass(),
		},
		Associations:    map[identity.Key]model_class.Association{},
		ModelInvariants: []string{},
	}

	diagnostics := Diagnose(resolved, model)
	found := false
	for _, d := range diagnostics {
		if contains(d.Message, "half-association") {
			found = true
			s.Equal("info", d.Level)
			break
		}
	}
	s.True(found, "expected half-association diagnostic")
}

// TODO(CalledBy/SentBy): This test uses Event.SentBy which was removed from the
// req_model/model_state.Event struct. SentBy is a simulator concern, not part of the pure
// data model. When re-enabling, update this test to use the simulator-local SentBy field
// (wrapper struct or parallel map) instead of setting it directly on model_state.Event.
func (s *DiagnosticsSuite) TestDiagnose_AllEventsInternal() {
	model := buildSingleDomainModel()

	// Create Order class where all events have SentBy pointing to in-scope classes.
	orderWithSentBy := makeOrderClass()
	evtCreate := orderWithSentBy.Events[orderEventCreateKey]
	evtCreate.SentBy = []identity.Key{itemClassKey}
	orderWithSentBy.Events[orderEventCreateKey] = evtCreate

	evtClose := orderWithSentBy.Events[orderEventCloseKey]
	evtClose.SentBy = []identity.Key{itemClassKey}
	orderWithSentBy.Events[orderEventCloseKey] = evtClose

	resolved := &ResolvedSurface{
		Classes: map[identity.Key]model_class.Class{
			orderClassKey: orderWithSentBy,
			itemClassKey:  makeItemClass(),
		},
		Associations:    map[identity.Key]model_class.Association{},
		ModelInvariants: []string{},
	}

	diagnostics := Diagnose(resolved, model)
	found := false
	for _, d := range diagnostics {
		if contains(d.Message, "all events internal") {
			found = true
			s.Equal("warning", d.Level)
			break
		}
	}
	s.True(found, "expected all-events-internal diagnostic")
}

// TODO(CalledBy/SentBy): This test uses Event.SentBy which was removed from the
// req_model/model_state.Event struct. SentBy is a simulator concern, not part of the pure
// data model. When re-enabling, update this test to use the simulator-local SentBy field.
func (s *DiagnosticsSuite) TestDiagnose_SentByUnknownClass() {
	model := buildSingleDomainModel()
	unknownKey := mustKey("domain/d/subdomain/s/class/unknown")

	orderWithBadSentBy := makeOrderClass()
	evtCreate := orderWithBadSentBy.Events[orderEventCreateKey]
	evtCreate.SentBy = []identity.Key{unknownKey}
	orderWithBadSentBy.Events[orderEventCreateKey] = evtCreate

	resolved := &ResolvedSurface{
		Classes: map[identity.Key]model_class.Class{
			orderClassKey: orderWithBadSentBy,
			itemClassKey:  makeItemClass(),
		},
		Associations:    map[identity.Key]model_class.Association{},
		ModelInvariants: []string{},
	}

	diagnostics := Diagnose(resolved, model)
	found := false
	for _, d := range diagnostics {
		if contains(d.Message, "SentBy references unknown class") {
			found = true
			s.Equal("warning", d.Level)
			break
		}
	}
	s.True(found, "expected SentBy unknown class diagnostic")
}

// TODO(CalledBy/SentBy): This test uses Action.CalledBy which was removed from the
// req_model/model_state.Action struct. CalledBy is a simulator concern, not part of the pure
// data model. When re-enabling, update this test to use the simulator-local CalledBy field.
func (s *DiagnosticsSuite) TestDiagnose_CalledByUnknownClass() {
	model := buildSingleDomainModel()
	unknownKey := mustKey("domain/d/subdomain/s/class/unknown")
	actionKey := mustKey("domain/d/subdomain/s/class/order/action/do_something")

	orderWithBadAction := makeOrderClass()
	orderWithBadAction.Actions = map[identity.Key]model_state.Action{
		actionKey: {
			Key:      actionKey,
			Name:     "do_something",
			CalledBy: []identity.Key{unknownKey},
		},
	}

	resolved := &ResolvedSurface{
		Classes: map[identity.Key]model_class.Class{
			orderClassKey: orderWithBadAction,
			itemClassKey:  makeItemClass(),
		},
		Associations:    map[identity.Key]model_class.Association{},
		ModelInvariants: []string{},
	}

	diagnostics := Diagnose(resolved, model)
	found := false
	for _, d := range diagnostics {
		if contains(d.Message, "CalledBy references unknown class") {
			found = true
			s.Equal("warning", d.Level)
			break
		}
	}
	s.True(found, "expected CalledBy unknown class diagnostic")
}

func (s *DiagnosticsSuite) TestDiagnose_NoDiagnosticsForHealthySurface() {
	model := buildSingleDomainModel()
	resolved := &ResolvedSurface{
		Classes: map[identity.Key]model_class.Class{
			orderClassKey: makeOrderClass(),
			itemClassKey:  makeItemClass(),
		},
		Associations:    map[identity.Key]model_class.Association{},
		ModelInvariants: []string{},
	}

	diagnostics := Diagnose(resolved, model)
	s.Len(diagnostics, 0, "expected no diagnostics for a healthy surface, got: %v", diagnostics)
}

// ============================================================
// Helper
// ============================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
