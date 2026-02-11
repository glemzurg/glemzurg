package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestEventSuite(t *testing.T) {
	suite.Run(t, new(EventSuite))
}

type EventSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Event.
func (suite *EventSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewEventKey(classKey, "event1"))

	tests := []struct {
		testName string
		event    Event
		errstr   string
	}{
		{
			testName: "valid event minimal",
			event: Event{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "valid event with all optional fields",
			event: Event{
				Key:     validKey,
				Name:    "Name",
				Details: "Details",
			},
		},
		{
			testName: "error empty key",
			event: Event{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			event: Event{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "Key: invalid key type 'domain' for event",
		},
		{
			testName: "error blank name",
			event: Event{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewEvent maps parameters correctly and calls Validate.
func (suite *EventSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	key := helper.Must(identity.NewEventKey(classKey, "event1"))

	// Test parameters are mapped correctly.
	event, err := NewEvent(key, "Name", "Details",
		[]Parameter{{Name: "ParamA", DataTypeRules: "Nat"}})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Event{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		Parameters: []Parameter{{Name: "ParamA", DataTypeRules: "Nat"}},
	}, event)

	// Test with nil optional Parameters.
	event, err = NewEvent(key, "Name", "Details", nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Event{
		Key:     key,
		Name:    "Name",
		Details: "Details",
	}, event)

	// Test that Validate is called (invalid data should fail).
	_, err = NewEvent(key, "", "Details", nil)
	assert.ErrorContains(suite.T(), err, "Name")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *EventSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewEventKey(classKey, "event1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))

	// Test that Validate is called.
	event := Event{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := event.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - event key has class1 as parent, but we pass other_class.
	event = Event{
		Key:  validKey,
		Name: "Name",
	}
	err = event.ValidateWithParent(&otherClassKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = event.ValidateWithParent(&classKey)
	assert.NoError(suite.T(), err)
}
