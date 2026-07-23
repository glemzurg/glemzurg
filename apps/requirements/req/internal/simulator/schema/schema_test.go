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

func (s *SchemaTestSuite) TestNewEmpty() {
	sch := NewEmpty()
	s.False(sch.IsClassInScope(identity.Key{}))
	s.Empty(sch.ClassKeys())
	s.Empty(sch.AssociationKeys())
}

func (s *SchemaTestSuite) TestNewFromModel_Nil() {
	sch := NewFromModel(nil)
	s.NotNil(sch)
	s.Empty(sch.ClassKeys())
}

func (s *SchemaTestSuite) TestNewFromModel_ClassesAttributesAssociations() {
	model, orderKey, lineKey, assocKey, attrKey := s.sampleModel()

	sch := NewFromModel(model)

	s.Same(model, sch.CoreModel())
	s.True(sch.IsClassInScope(orderKey))
	s.True(sch.IsClassInScope(lineKey))

	order, ok := sch.Class(orderKey)
	s.True(ok)
	s.Equal("Order", order.Name)
	s.Require().Len(order.Attributes, 1)
	s.Equal(attrKey, order.Attributes[0].Key)

	fullClass, ok := sch.ModelClass(orderKey)
	s.True(ok)
	s.Equal("Order", fullClass.Name)

	attrs := sch.Attributes(orderKey)
	s.Require().Len(attrs, 1)
	s.Equal("status", attrs[0].Name)

	assoc, ok := sch.Association(assocKey)
	s.True(ok)
	s.Equal("Lines", assoc.Name)
	s.Equal(orderKey, assoc.FromClassKey)
	s.Equal(lineKey, assoc.ToClassKey)
	s.Nil(assoc.AssociationClassKey)
	s.False(sch.IsAssociationClass(orderKey))

	s.Len(sch.ClassKeys(), 2)
	s.Len(sch.AssociationKeys(), 1)
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
