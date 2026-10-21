package parser_json

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func TestMultiplicityTokens(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{name: "one", token: "1"},
		{name: "optional", token: "0..1"},
		{name: "any", token: "any"},
		{name: "one to many", token: "1..many"},
		{name: "range", token: "2..4"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			multiplicity, err := model_class.NewMultiplicity(tc.token)
			require.NoError(t, err)
			got, err := multiplicityFromCore(multiplicity)
			require.NoError(t, err)
			assert.Equal(t, tc.token, got)
		})
	}
}

func TestWriteReadDomainsClassesAssociations(t *testing.T) {
	model := classStructureModel(t)
	require.NoError(t, model.Validate())

	doc := writeModel(t, model)
	text := string(doc)
	assert.Contains(t, text, `"to_multiplicity":"1..many"`)
	assert.Contains(t, text, `"to_multiplicity":"any"`)
	assert.Contains(t, text, `"from_multiplicity":"0..1"`)
	assert.Less(t, strings.Index(text, `"key":"domain/commerce"`), strings.Index(text, `"key":"domain/logistics"`))

	got := readModel(t, doc)
	assert.Equal(t, model.Domains, got.Domains)
	assert.Equal(t, model.DomainAssociations, got.DomainAssociations)
	assert.Equal(t, model.ClassAssociations, got.ClassAssociations)
	assert.Equal(t, model.Actors, got.Actors)
}

func classStructureModel(t *testing.T) core.Model {
	t.Helper()

	commerceKey := mustKey(t, func() (identity.Key, error) { return identity.NewDomainKey("commerce") })
	logisticsKey := mustKey(t, func() (identity.Key, error) { return identity.NewDomainKey("logistics") })
	ordersKey := mustKey(t, func() (identity.Key, error) { return identity.NewSubdomainKey(commerceKey, "orders") })
	warehouseKey := mustKey(t, func() (identity.Key, error) { return identity.NewSubdomainKey(commerceKey, "warehouse") })
	freightKey := mustKey(t, func() (identity.Key, error) { return identity.NewSubdomainKey(logisticsKey, "freight") })
	yardsKey := mustKey(t, func() (identity.Key, error) { return identity.NewSubdomainKey(logisticsKey, "yards") })
	genKey := mustKey(t, func() (identity.Key, error) { return identity.NewGeneralizationKey(ordersKey, "priority") })
	orderKey := mustKey(t, func() (identity.Key, error) { return identity.NewClassKey(ordersKey, "order") })
	rushKey := mustKey(t, func() (identity.Key, error) { return identity.NewClassKey(ordersKey, "rush_order") })
	lineKey := mustKey(t, func() (identity.Key, error) { return identity.NewClassKey(ordersKey, "line") })
	receiptKey := mustKey(t, func() (identity.Key, error) { return identity.NewClassKey(ordersKey, "receipt") })
	binKey := mustKey(t, func() (identity.Key, error) { return identity.NewClassKey(warehouseKey, "bin") })
	shipmentKey := mustKey(t, func() (identity.Key, error) { return identity.NewClassKey(freightKey, "shipment") })
	customerKey := mustKey(t, func() (identity.Key, error) { return identity.NewActorKey("customer") })

	nameKey := mustKey(t, func() (identity.Key, error) { return identity.NewAttributeKey(orderKey, "name") })
	totalKey := mustKey(t, func() (identity.Key, error) { return identity.NewAttributeKey(orderKey, "total") })
	numKey := mustKey(t, func() (identity.Key, error) { return identity.NewAttributeKey(lineKey, "num") })
	totalDeriveKey := mustKey(t, func() (identity.Key, error) { return identity.NewAttributeDerivationKey(totalKey, "1") })
	nameInvKey := mustKey(t, func() (identity.Key, error) { return identity.NewAttributeInvariantKey(nameKey, "1") })
	classInvKey := mustKey(t, func() (identity.Key, error) { return identity.NewClassInvariantKey(orderKey, "1") })

	nameAttr, err := model_class.NewAttribute(nameKey, model_class.AttributeDetails{Name: "Name", Details: "Order name."}, "unconstrained", nil, false, model_class.AttributeAnnotations{IndexNums: []uint{1}})
	require.NoError(t, err)
	nameAttr.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(nameInvKey, model_logic.LogicTypeAssessment, "present", "", mustSpec(t, "self.name # \"\""), nil),
	})
	totalDerive := model_logic.NewLogic(totalDeriveKey, model_logic.LogicTypeValue, "sum", "", mustSpec(t, "0"), nil)
	totalAttr, err := model_class.NewAttribute(totalKey, model_class.AttributeDetails{Name: "Total"}, "unconstrained", &totalDerive, true, model_class.AttributeAnnotations{})
	require.NoError(t, err)
	numAttr, err := model_class.NewAttribute(numKey, model_class.AttributeDetails{Name: "Num"}, "unconstrained", nil, false, model_class.AttributeAnnotations{})
	require.NoError(t, err)

	order := model_class.NewClass(orderKey, model_class.ClassLinks{
		ActorKey:        &customerKey,
		SuperclassOfKey: &genKey,
	}, model_class.ClassDetails{Name: "Order", Details: "A customer order."})
	order.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(classInvKey, model_logic.LogicTypeAssessment, "named", "", mustSpec(t, "self.name # \"\""), nil),
	})
	order.SetAttributes([]model_class.Attribute{nameAttr, totalAttr})

	rush := model_class.NewClass(rushKey, model_class.ClassLinks{SubclassOfKey: &genKey}, model_class.ClassDetails{Name: "Rush Order"})
	rush.SetMarked(true)
	line := model_class.NewClass(lineKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Line"})
	line.SetAttributes([]model_class.Attribute{numAttr})
	receipt := model_class.NewClass(receiptKey, model_class.ClassLinks{FacetOf: &orderKey}, model_class.ClassDetails{Name: "Receipt"})
	bin := model_class.NewClass(binKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Bin"})
	shipment := model_class.NewClass(shipmentKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Shipment"})

	linesKey := mustKey(t, func() (identity.Key, error) {
		return identity.NewClassAssociationKey(ordersKey, orderKey, lineKey, "Has Lines")
	})
	linesInvKey := mustKey(t, func() (identity.Key, error) {
		return identity.NewClassAssociationInvariantKey(linesKey, "1")
	})
	lines := model_class.NewAssociation(linesKey, model_class.AssociationDetails{Name: "Has Lines", Details: "Order lines."},
		model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: mustMult(t, "1")},
		model_class.AssociationEnd{ClassKey: lineKey, Multiplicity: mustMult(t, "1..many")},
		model_class.AssociationOptions{Uniqueness: []model_class.AssociationUniqueness{
			model_class.NewAssociationUniqueness(nil, []identity.Key{numKey}),
		}})
	lines.SetInvariants([]model_logic.Logic{
		model_logic.NewLogic(linesInvKey, model_logic.LogicTypeAssessment, "numbered", "", mustSpec(t, "TRUE"), nil),
	})

	storesKey := mustKey(t, func() (identity.Key, error) {
		return identity.NewClassAssociationKey(commerceKey, orderKey, binKey, "Stores")
	})
	stores := model_class.NewAssociation(storesKey, model_class.AssociationDetails{Name: "Stores"},
		model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: mustMult(t, "1")},
		model_class.AssociationEnd{ClassKey: binKey, Multiplicity: mustMult(t, "any")},
		model_class.AssociationOptions{})

	shipsKey := mustKey(t, func() (identity.Key, error) {
		return identity.NewClassAssociationKey(identity.Key{}, orderKey, shipmentKey, "Ships")
	})
	ships := model_class.NewAssociation(shipsKey, model_class.AssociationDetails{Name: "Ships"},
		model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: mustMult(t, "0..1")},
		model_class.AssociationEnd{ClassKey: shipmentKey, Multiplicity: mustMult(t, "1")},
		model_class.AssociationOptions{UmlComment: "cross domain"})

	orders := model_domain.NewSubdomain(ordersKey, "Orders", "Customer orders.", "", "")
	orders.Generalizations = map[identity.Key]model_class.Generalization{
		genKey: model_class.NewGeneralization(genKey, model_class.GeneralizationDetails{Name: "Priority"}, "", model_class.GeneralizationTraits{IsComplete: true}, ""),
	}
	orders.Classes = map[identity.Key]model_class.Class{
		orderKey: order, rushKey: rush, lineKey: line, receiptKey: receipt,
	}
	orders.ClassAssociations = map[identity.Key]model_class.Association{linesKey: lines}

	warehouse := model_domain.NewSubdomain(warehouseKey, "Warehouse", "", "", "")
	warehouse.Classes = map[identity.Key]model_class.Class{binKey: bin}
	freight := model_domain.NewSubdomain(freightKey, "Freight", "", "", "")
	freight.Classes = map[identity.Key]model_class.Class{shipmentKey: shipment}
	yards := model_domain.NewSubdomain(yardsKey, "Yards", "", "", "")

	subAssocKey := mustKey(t, func() (identity.Key, error) {
		return identity.NewSubdomainAssociationKey(commerceKey, ordersKey, warehouseKey)
	})
	commerce := model_domain.NewDomain(commerceKey, "Commerce", "Selling.", "", false, "")
	commerce.Subdomains = map[identity.Key]model_domain.Subdomain{
		ordersKey: orders, warehouseKey: warehouse,
	}
	commerce.SubdomainAssociations = map[identity.Key]model_domain.SubdomainAssociation{
		subAssocKey: model_domain.NewSubdomainAssociation(subAssocKey, ordersKey, warehouseKey, "fulfillment"),
	}
	commerce.ClassAssociations = map[identity.Key]model_class.Association{storesKey: stores}

	logistics := model_domain.NewDomain(logisticsKey, "Logistics", "", "", true, "outside")
	logistics.Subdomains = map[identity.Key]model_domain.Subdomain{freightKey: freight, yardsKey: yards}

	domainAssocKey := mustKey(t, func() (identity.Key, error) {
		return identity.NewDomainAssociationKey(commerceKey, logisticsKey)
	})

	model := core.NewModel("sample", core.ModelDetails{Name: "Sample"}, "", nil, nil, nil)
	model.Actors = map[identity.Key]model_actor.Actor{
		customerKey: model_actor.NewActor(customerKey, "person", model_actor.GeneralizationRefs{}, model_actor.ActorDetails{Name: "Customer"}),
	}
	model.Domains = map[identity.Key]model_domain.Domain{commerceKey: commerce, logisticsKey: logistics}
	model.DomainAssociations = map[identity.Key]model_domain.Association{
		domainAssocKey: model_domain.NewAssociation(domainAssocKey, commerceKey, logisticsKey, ""),
	}
	model.ClassAssociations = map[identity.Key]model_class.Association{shipsKey: ships}
	return model
}

func mustMult(t *testing.T, token string) model_class.Multiplicity {
	t.Helper()
	multiplicity, err := model_class.NewMultiplicity(token)
	require.NoError(t, err)
	return multiplicity
}
