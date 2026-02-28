package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
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
	classKey := mustKey("domain/d/subdomain/default/class/order")
	stateKey := mustKey("domain/d/subdomain/default/class/order/state/open")
	eventKey := mustKey("domain/d/subdomain/default/class/order/event/create")
	transKey := mustKey("domain/d/subdomain/default/class/order/transition/create")
	subdomainKey := mustKey("domain/d/subdomain/default")
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
	classKey := mustKey("domain/d/subdomain/default/class/order")
	attrKey := mustKey("domain/d/subdomain/default/class/order/attribute/amount")
	stateKey := mustKey("domain/d/subdomain/default/class/order/state/open")
	eventKey := mustKey("domain/d/subdomain/default/class/order/event/create")
	transKey := mustKey("domain/d/subdomain/default/class/order/transition/create")
	subdomainKey := mustKey("domain/d/subdomain/default")
	domainKey := mustKey("domain/d")

	model := buildTestModel(domainKey, subdomainKey, classKey, stateKey, eventKey, transKey)

	// Add an attribute with a parseable DataTypeRules (enum format).
	attr, err := model_class.NewAttribute(attrKey, "amount", "", "enum of small, medium, large", nil, false, "", nil)
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

	transition := helper.Must(model_state.NewTransition(transKey, nil, eventKey, nil, nil, &toStateKey, ""))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateKey: helper.Must(model_state.NewState(stateKey, "Open", "", "")),
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventKey: helper.Must(model_state.NewEvent(eventKey, "create", "", nil)),
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transKey: transition,
	})

	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "S", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: class,
	}

	domain := helper.Must(model_domain.NewDomain(domainKey, "D", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	model := helper.Must(req_model.NewModel("test", "Test Model", "", nil, nil, nil))
	model.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}

	return &model
}
