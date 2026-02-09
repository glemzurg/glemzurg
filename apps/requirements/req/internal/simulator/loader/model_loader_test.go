package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/identity"
	"github.com/glemzurg/go-tlaplus/internal/req_model"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_class"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_domain"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_state"
	"github.com/stretchr/testify/suite"
)

func TestModelLoaderSuite(t *testing.T) {
	suite.Run(t, new(ModelLoaderSuite))
}

type ModelLoaderSuite struct {
	suite.Suite
	tempDir string
}

func (s *ModelLoaderSuite) SetupTest() {
	s.tempDir = s.T().TempDir()
}

func mustKey(str string) identity.Key {
	k, err := identity.ParseKey(str)
	if err != nil {
		panic(err)
	}
	return k
}

func (s *ModelLoaderSuite) TestRoundTrip() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	eventKey := mustKey("domain/d/subdomain/s/class/order/event/create")
	transKey := mustKey("domain/d/subdomain/s/class/order/transition/create")
	subdomainKey := mustKey("domain/d/subdomain/s")
	domainKey := mustKey("domain/d")

	model := buildTestModel(domainKey, subdomainKey, classKey, stateKey, eventKey, transKey)

	// Save.
	path := filepath.Join(s.tempDir, "model.json")
	err := SaveModel(model, path)
	s.Require().NoError(err)

	// Load.
	loaded, err := LoadModel(path)
	s.Require().NoError(err)

	// Verify key fields.
	s.Equal("test", loaded.Key)
	s.Equal("Test Model", loaded.Name)
	s.Len(loaded.Domains, 1)

	domain := loaded.Domains[domainKey]
	s.Len(domain.Subdomains, 1)

	subdomain := domain.Subdomains[subdomainKey]
	s.Len(subdomain.Classes, 1)

	class := subdomain.Classes[classKey]
	s.Equal("Order", class.Name)
	s.Len(class.States, 1)
	s.Len(class.Events, 1)
	s.Len(class.Transitions, 1)
}

func (s *ModelLoaderSuite) TestRoundTripWithDataTypeRules() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	attrKey := mustKey("domain/d/subdomain/s/class/order/attribute/amount")
	stateKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	eventKey := mustKey("domain/d/subdomain/s/class/order/event/create")
	transKey := mustKey("domain/d/subdomain/s/class/order/transition/create")
	subdomainKey := mustKey("domain/d/subdomain/s")
	domainKey := mustKey("domain/d")

	model := buildTestModel(domainKey, subdomainKey, classKey, stateKey, eventKey, transKey)

	// Add an attribute with a parseable DataTypeRules (enum format).
	attr, err := model_class.NewAttribute(attrKey, "amount", "", "enum of small, medium, large", "", "", false, "", nil)
	s.Require().NoError(err)
	s.Require().NotNil(attr.DataType, "enum DataTypeRules should parse successfully")

	domain := model.Domains[domainKey]
	subdomain := domain.Subdomains[subdomainKey]
	class := subdomain.Classes[classKey]
	class.Attributes[attrKey] = attr
	subdomain.Classes[classKey] = class
	domain.Subdomains[subdomainKey] = subdomain
	model.Domains[domainKey] = domain

	// Save.
	path := filepath.Join(s.tempDir, "model_with_dt.json")
	err = SaveModel(model, path)
	s.Require().NoError(err)

	// Load.
	loaded, err := LoadModel(path)
	s.Require().NoError(err)

	// Verify the attribute round-tripped.
	loadedDomain := loaded.Domains[domainKey]
	loadedSubdomain := loadedDomain.Subdomains[subdomainKey]
	loadedClass := loadedSubdomain.Classes[classKey]
	loadedAttr := loadedClass.Attributes[attrKey]

	s.Equal("amount", loadedAttr.Name)
	s.Equal("enum of small, medium, large", loadedAttr.DataTypeRules)
	// DataType should survive round-trip (either preserved from JSON or re-parsed).
	s.NotNil(loadedAttr.DataType)
}

func (s *ModelLoaderSuite) TestNonExistentFile() {
	_, err := LoadModel("/nonexistent/path/model.json")
	s.Error(err)
	s.Contains(err.Error(), "reading model file")
}

func (s *ModelLoaderSuite) TestInvalidJSON() {
	path := filepath.Join(s.tempDir, "bad.json")
	err := os.WriteFile(path, []byte("not json"), 0644)
	s.Require().NoError(err)

	_, err = LoadModel(path)
	s.Error(err)
	s.Contains(err.Error(), "parsing model JSON")
}

func (s *ModelLoaderSuite) TestInvalidModel() {
	// A model with no Key should fail validation.
	path := filepath.Join(s.tempDir, "invalid.json")
	err := os.WriteFile(path, []byte(`{"Name": "Test"}`), 0644)
	s.Require().NoError(err)

	_, err = LoadModel(path)
	s.Error(err)
	s.Contains(err.Error(), "validating model")
}

// buildTestModel creates a minimal valid model for testing.
func buildTestModel(
	domainKey, subdomainKey, classKey, stateKey, eventKey, transKey identity.Key,
) *req_model.Model {
	toStateKey := stateKey
	return &req_model.Model{
		Key:  "test",
		Name: "Test Model",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "D",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:  subdomainKey,
						Name: "S",
						Classes: map[identity.Key]model_class.Class{
							classKey: {
								Key:        classKey,
								Name:       "Order",
								Attributes: map[identity.Key]model_class.Attribute{},
								States: map[identity.Key]model_state.State{
									stateKey: {Key: stateKey, Name: "Open"},
								},
								Events: map[identity.Key]model_state.Event{
									eventKey: {Key: eventKey, Name: "create"},
								},
								Guards:  map[identity.Key]model_state.Guard{},
								Actions: map[identity.Key]model_state.Action{},
								Queries: map[identity.Key]model_state.Query{},
								Transitions: map[identity.Key]model_state.Transition{
									transKey: {
										Key:          transKey,
										FromStateKey: nil,
										EventKey:     eventKey,
										ToStateKey:   &toStateKey,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
