package surface

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveClassKeysByName_UnqualifiedMatchesAllSubdomains(t *testing.T) {
	model := buildTwoDomainModel()

	keys, err := ResolveClassKeysByName(model, []string{"payment"})
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, paymentClassKey, keys[0])
}

func TestResolveClassKeysByName_SubdomainClass(t *testing.T) {
	model := buildTwoDomainModel()

	keys, err := ResolveClassKeysByName(model, []string{"s/order"})
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, orderClassKey, keys[0])
}

func TestResolveClassKeysByName_DomainSubdomainClass(t *testing.T) {
	model := buildTwoDomainModel()

	keys, err := ResolveClassKeysByName(model, []string{"d2/s2/payment"})
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, paymentClassKey, keys[0])
}

func TestResolveClassKeysByName_SubdomainClassNotFound(t *testing.T) {
	model := buildTwoDomainModel()

	_, err := ResolveClassKeysByName(model, []string{"s2/order"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `no class "order" in subdomain d2/s2`)
}

func TestResolveClassKeysByName_AmbiguousSubdomain(t *testing.T) {
	model := buildAmbiguousSubdomainModel()

	_, err := ResolveClassKeysByName(model, []string{"shared/partner"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous across domains")
}

func TestResolveSubdomainKeysByPath_SubdomainSubkey(t *testing.T) {
	model := buildTwoDomainModel()

	keys, err := ResolveSubdomainKeysByPath(model, []string{"s"})
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, subdomainKey, keys[0])
}

func TestResolveSubdomainKeysByPath_DomainSubdomain(t *testing.T) {
	model := buildTwoDomainModel()

	keys, err := ResolveSubdomainKeysByPath(model, []string{"d2/s2"})
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, subdomain2Key, keys[0])
}

func TestResolveSubdomainKeysByPath_NotFound(t *testing.T) {
	model := buildTwoDomainModel()

	_, err := ResolveSubdomainKeysByPath(model, []string{"missing"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `no subdomain "missing"`)
}

func TestResolveSubdomainKeysByPath_AmbiguousSubdomain(t *testing.T) {
	model := buildAmbiguousSubdomainModel()

	_, err := ResolveSubdomainKeysByPath(model, []string{"shared"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous across domains")
}

func TestResolveSubdomainKeysByPath_DomainDisambiguatesSubdomain(t *testing.T) {
	model := buildAmbiguousSubdomainModel()

	keys, err := ResolveSubdomainKeysByPath(model, []string{"d1/shared"})
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, ambiguousSubdomain1Key, keys[0])
}

func TestResolveClassKeysByName_DomainDisambiguatesSubdomain(t *testing.T) {
	model := buildAmbiguousSubdomainModel()

	keys, err := ResolveClassKeysByName(model, []string{"d1/shared/partner"})
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, partnerD1ClassKey, keys[0])
}

var (
	ambiguousDomain1Key    = mustKey("domain/d1")
	ambiguousDomain2Key    = mustKey("domain/d2")
	ambiguousSubdomain1Key = mustKey("domain/d1/subdomain/shared")
	ambiguousSubdomain2Key = mustKey("domain/d2/subdomain/shared")
	partnerD1ClassKey      = mustKey("domain/d1/subdomain/shared/class/partner")
	partnerD2ClassKey      = mustKey("domain/d2/subdomain/shared/class/partner")
)

func buildAmbiguousSubdomainModel() *core.Model {
	partnerD1 := makePartnerClass(partnerD1ClassKey)
	partnerD2 := makePartnerClass(partnerD2ClassKey)

	domain1 := model_domain.NewDomain(ambiguousDomain1Key, "D1", "", "", false, "")
	subdomain1 := model_domain.NewSubdomain(ambiguousSubdomain1Key, "Shared", "", "", "")
	subdomain1.Classes = map[identity.Key]model_class.Class{
		partnerD1ClassKey: partnerD1,
	}
	domain1.Subdomains = map[identity.Key]model_domain.Subdomain{
		ambiguousSubdomain1Key: subdomain1,
	}

	domain2 := model_domain.NewDomain(ambiguousDomain2Key, "D2", "", "", false, "")
	subdomain2 := model_domain.NewSubdomain(ambiguousSubdomain2Key, "Shared", "", "", "")
	subdomain2.Classes = map[identity.Key]model_class.Class{
		partnerD2ClassKey: partnerD2,
	}
	domain2.Subdomains = map[identity.Key]model_domain.Subdomain{
		ambiguousSubdomain2Key: subdomain2,
	}

	model := core.NewModel("ambiguous", "Ambiguous", "", "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{
		ambiguousDomain1Key: domain1,
		ambiguousDomain2Key: domain2,
	}
	return &model
}

func makePartnerClass(classKey identity.Key) model_class.Class {
	stateKey := helper.Must(identity.NewStateKey(classKey, "active"))
	eventKey := helper.Must(identity.NewEventKey(classKey, "add"))
	transKey := helper.Must(identity.NewTransitionKey(classKey, "", "add", "", "", "active"))

	class := model_class.NewClass(
		classKey,
		model_class.ClassLinks{},
		model_class.ClassDetails{Name: "Partner", Details: "", UnfinishedNotes: "", UmlComment: ""},
	)
	class.States = map[identity.Key]model_state.State{
		stateKey: model_state.NewState(stateKey, "Active", "", ""),
	}
	class.Events = map[identity.Key]model_state.Event{
		eventKey: model_state.NewEvent(eventKey, "Add", "", nil),
	}
	class.Transitions = map[identity.Key]model_state.Transition{
		transKey: model_state.NewTransition(transKey, nil, eventKey, nil, nil, &stateKey, ""),
	}
	return class
}
