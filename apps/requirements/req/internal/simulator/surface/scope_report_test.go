package surface

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ScopeReportSuite struct {
	suite.Suite
}

func TestScopeReportSuite(t *testing.T) {
	suite.Run(t, new(ScopeReportSuite))
}

func (s *ScopeReportSuite) TestBuildScopeEntries_WholeSubdomainWhenAllClassesIncluded() {
	model, orderKey, itemKey := scopeTestModelTwoClasses()
	inScope := map[identity.Key]model_class.Class{
		orderKey: model.Domains[domainKey].Subdomains[subdomainKey].Classes[orderKey],
		itemKey:  model.Domains[domainKey].Subdomains[subdomainKey].Classes[itemKey],
	}

	entries := BuildScopeEntries(&model, inScope)
	s.Require().Len(entries, 1)
	s.Equal(ScopeSubdomain, entries[0].Kind)
	s.Equal("d/s", entries[0].Path)
}

func (s *ScopeReportSuite) TestBuildScopeEntries_ListsClassesWhenPartialSubdomain() {
	model, orderKey, itemKey := scopeTestModelTwoClasses()
	inScope := map[identity.Key]model_class.Class{
		orderKey: model.Domains[domainKey].Subdomains[subdomainKey].Classes[orderKey],
	}

	entries := BuildScopeEntries(&model, inScope)
	s.Require().Len(entries, 1)
	s.Equal(ScopeClass, entries[0].Kind)
	s.Equal("d/s/order", entries[0].Path)
	s.NotContains(entries[0].Path, itemKey.SubKey)
}

func (s *ScopeReportSuite) TestBuildScopeEntries_EmptyWhenNothingInScope() {
	model, _, _ := scopeTestModelTwoClasses()
	entries := BuildScopeEntries(&model, map[identity.Key]model_class.Class{})
	s.Empty(entries)
}

func TestAllNonRealizedClasses(t *testing.T) {
	model, orderKey, itemKey := scopeTestModelTwoClasses()
	all := AllNonRealizedClasses(&model)
	require.Len(t, all, 2)
	require.Contains(t, all, orderKey)
	require.Contains(t, all, itemKey)
}

func scopeTestModelTwoClasses() (core.Model, identity.Key, identity.Key) {
	orderKey := orderClassKey
	itemKey := itemClassKey

	order := model_class.NewClass(orderKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order"})
	orderState := mustKey("domain/d/subdomain/s/class/order/state/open")
	order.States = map[identity.Key]model_state.State{
		orderState: model_state.NewState(orderState, "Open", "", ""),
	}
	item := model_class.NewClass(itemKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Item"})
	itemState := mustKey("domain/d/subdomain/s/class/item/state/active")
	item.States = map[identity.Key]model_state.State{
		itemState: model_state.NewState(itemState, "Active", "", ""),
	}

	sub := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	sub.Classes = map[identity.Key]model_class.Class{orderKey: order, itemKey: item}
	dom := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	dom.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: sub}
	model := core.NewModel("m", core.ModelDetails{Name: "m", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: dom}
	return model, orderKey, itemKey
}
