package schema

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SchemaTestSuite struct {
	suite.Suite
}

func TestSchemaSuite(t *testing.T) {
	suite.Run(t, new(SchemaTestSuite))
}

func (s *SchemaTestSuite) TestNew_RequiresModel() {
	s.Panics(func() { New(nil, RunScopeAll()) })
}

func (s *SchemaTestSuite) TestNew_EmptyModel() {
	sch := New(emptyModel(), RunScopeAll())
	s.NotNil(sch)
	s.Empty(sch.classKeys())
	s.Empty(sch.associationKeys())
	s.False(sch.IsClassInScope(identity.Key{}))
}

func (s *SchemaTestSuite) TestNew_ClassesAttributesAssociations() {
	model, orderKey, lineKey, assocKey, attrKey := s.sampleModel()

	sch := New(model, RunScopeAll())

	s.True(sch.IsClassInScope(orderKey))
	s.True(sch.IsClassInScope(lineKey))

	order, ok := sch.class(orderKey)
	s.True(ok)
	s.Equal("Order", order.Name)
	s.Require().Len(order.Attributes, 1)
	s.Equal(attrKey, order.Attributes[0].Key)

	attrs := sch.attributes(orderKey)
	s.Require().Len(attrs, 1)
	s.Equal("status", attrs[0].Name)

	assoc, ok := sch.association(assocKey)
	s.True(ok)
	s.Equal("Lines", assoc.Name)
	s.Equal(orderKey, assoc.FromClassKey)
	s.Equal(lineKey, assoc.ToClassKey)
	s.Nil(assoc.AssociationClassKey)
	s.False(sch.isAssociationClass(orderKey))

	s.Len(sch.classKeys(), 2)
	s.Len(sch.associationKeys(), 1)

	// Public triples: in-scope hits.
	orderPtr, inScope, err := sch.Class(orderKey)
	s.Require().NoError(err)
	s.True(inScope)
	s.Require().NotNil(orderPtr)
	s.Equal("Order", orderPtr.Name)

	assocPtr, inScope, err := sch.Association(assocKey)
	s.Require().NoError(err)
	s.True(inScope)
	s.Require().NotNil(assocPtr)
	s.Equal("Lines", assocPtr.Name)
}

func (s *SchemaTestSuite) TestClass_OutOfScopeAndUnknown() {
	model, orderKey, lineKey, assocKey, _ := s.sampleModel()
	// Only Order in scope — Line is out of scope; association is boundary.
	sch := New(model, NewRunScope([]identity.Key{orderKey}))

	order, inScope, err := sch.Class(orderKey)
	s.Require().NoError(err)
	s.True(inScope)
	s.NotNil(order)

	line, inScope, err := sch.Class(lineKey)
	s.Require().NoError(err)
	s.False(inScope)
	s.Nil(line)

	_, _, err = sch.Class(mustParse(s.T(), "domain/d/subdomain/s/class/missing"))
	s.Require().Error(err)

	// Both ends must be in scope for Association triple.
	a, inScope, err := sch.Association(assocKey)
	s.Require().NoError(err)
	s.False(inScope)
	s.Nil(a)

	boundary := sch.BoundaryAssociations()
	s.Require().Len(boundary, 1)
	s.Equal(assocKey, boundary[0].Key)

	name, inScope, err := sch.ExtentName(lineKey)
	s.Require().NoError(err)
	s.False(inScope)
	s.NotEmpty(name)

	// EachInScopeClassSim only includes in-scope classes.
	var names []string
	sch.EachInScopeClassSim(func(sim *ClassSimInfo) {
		names = append(names, sim.Class.Name)
	})
	s.Equal([]string{"Order"}, names)
}

func (s *SchemaTestSuite) sampleModel() (
	*core.Model,
	identity.Key,
	identity.Key,
	identity.Key,
	identity.Key,
) {
	t := s.T()
	domainKey := mustParse(t, "domain/d")
	subKey := mustParse(t, "domain/d/subdomain/s")
	orderKey := mustParse(t, "domain/d/subdomain/s/class/order")
	lineKey := mustParse(t, "domain/d/subdomain/s/class/line")
	assocKey, err := identity.NewClassAssociationKey(subKey, orderKey, lineKey, "lines")
	require.NoError(t, err)
	attrKey, err := identity.NewAttributeKey(orderKey, "status")
	require.NoError(t, err)

	attr, err := model_class.NewAttribute(
		attrKey,
		model_class.AttributeDetails{Name: "status", Details: ""},
		"string",
		nil,
		false,
		model_class.AttributeAnnotations{},
	)
	require.NoError(t, err)

	order := model_class.NewClass(
		orderKey,
		model_class.ClassLinks{},
		model_class.ClassDetails{Name: "Order"},
	)
	order.SetAttributes([]model_class.Attribute{attr})

	line := model_class.NewClass(
		lineKey,
		model_class.ClassLinks{},
		model_class.ClassDetails{Name: "Line"},
	)

	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Lines", Details: ""},
		model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult},
		model_class.AssociationEnd{ClassKey: lineKey, Multiplicity: toMult},
		model_class.AssociationOptions{},
	)

	subdomain := model_domain.NewSubdomain(subKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		orderKey: order,
		lineKey:  line,
	}
	subdomain.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}

	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subKey: subdomain,
	}

	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}

	return &model, orderKey, lineKey, assocKey, attrKey
}

func mustParse(t *testing.T, s string) identity.Key {
	t.Helper()
	k, err := identity.ParseKey(s)
	require.NoError(t, err)
	return k
}
