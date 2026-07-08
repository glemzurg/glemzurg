package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
			testName: "valid system creation event name",
			event: Event{
				Key:  validKey,
				Name: EventNameNew,
			},
		},
		{
			testName: "valid system final event name",
			event: Event{
				Key:  validKey,
				Name: EventNameDestroy,
			},
		},
		{
			testName: "valid event with all optional fields",
			event: Event{
				Key:            validKey,
				Name:           "Name",
				Details:        "Details",
				ParameterNames: []string{"ParamA", "ParamB"},
			},
		},
		{
			testName: "error empty key",
			event: Event{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "key type is required",
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
		{
			testName: "error name with invalid chars",
			event: Event{
				Key:  validKey,
				Name: "Fail On Name/DOB",
			},
			errstr: "EVENT_NAME_INVALID_CHARS",
		},
		{
			testName: "error name with spaces",
			event: Event{
				Key:  validKey,
				Name: "Payment Received",
			},
			errstr: "EVENT_NAME_INVALID_CHARS",
		},
		{
			testName: "error blank parameter name",
			event: Event{
				Key:            validKey,
				Name:           "Name",
				ParameterNames: []string{"ParamA", ""},
			},
			errstr: "EVENT_PARAMETER_NAME_REQUIRED",
		},
		{
			testName: "error duplicate parameter names after normalization",
			event: Event{
				Key:            validKey,
				Name:           "Name",
				ParameterNames: []string{"CountryCode", "countrycode"},
			},
			errstr: "EVENT_PARAMETER_NAME_DUPLICATE",
		},
		{
			testName: "error parameter name with spaces",
			event: Event{
				Key:            validKey,
				Name:           "Submit",
				ParameterNames: []string{"user id"},
			},
			errstr: "EVENT_PARAMETER_NAME_INVALID_CHARS",
		},
		{
			testName: "error name with hyphen",
			event: Event{
				Key:  validKey,
				Name: "Pay-Now",
			},
			errstr: "EVENT_NAME_INVALID_CHARS",
		},
		{
			testName: "error parameter name with hyphen",
			event: Event{
				Key:            validKey,
				Name:           "Submit",
				ParameterNames: []string{"user-id"},
			},
			errstr: "EVENT_PARAMETER_NAME_INVALID_CHARS",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			ctx := coreerr.NewContext("test", "")
			err := tt.event.Validate(ctx)
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewEvent maps parameter names correctly.
func (suite *EventSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	key := helper.Must(identity.NewEventKey(classKey, "event1"))

	event := NewEvent(key, "Name", "Details", []string{"ParamA", "ParamB"})
	suite.Equal(Event{
		Key:            key,
		Name:           "Name",
		Details:        "Details",
		ParameterNames: []string{"ParamA", "ParamB"},
	}, event)

	event = NewEvent(key, "Name", "Details", nil)
	suite.Equal(Event{
		Key:     key,
		Name:    "Name",
		Details: "Details",
	}, event)
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *EventSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewEventKey(classKey, "event1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))

	ctx := coreerr.NewContext("test", "")

	event := Event{
		Key:  validKey,
		Name: "",
	}
	err := event.ValidateWithParent(ctx, &classKey)
	suite.Require().ErrorContains(err, "Name", "ValidateWithParent should call Validate()")

	event = Event{
		Key:  validKey,
		Name: "Name",
	}
	err = event.ValidateWithParent(ctx, &otherClassKey)
	suite.Require().ErrorContains(err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	err = event.ValidateWithParent(ctx, &classKey)
	suite.Require().NoError(err)

	event = Event{
		Key:            validKey,
		Name:           "Name",
		ParameterNames: []string{"param1"},
	}
	err = event.ValidateWithParent(ctx, &classKey)
	suite.Require().NoError(err)
}

func (suite *EventSuite) TestIsValidEventName() {
	tests := []struct {
		testName string
		name     string
		valid    bool
	}{
		{testName: "submit", name: "Submit", valid: true},
		{testName: "system new", name: EventNameNew, valid: true},
		{testName: "system destroy", name: EventNameDestroy, valid: true},
		{testName: "with spaces", name: "Payment Received", valid: false},
		{testName: "with slash", name: "bad/name", valid: false},
	}
	for _, tc := range tests {
		suite.Run(tc.testName, func() {
			suite.Equal(tc.valid, isValidEventName(tc.name))
		})
	}
}

func (suite *EventSuite) TestSystemEventNames() {
	suite.True(IsSystemCreationEvent(EventNameNew))
	suite.False(IsSystemCreationEvent("Add"))
	suite.True(IsSystemFinalEvent(EventNameDestroy))
	suite.False(IsSystemFinalEvent("Delete"))
	suite.Equal("«new»", SystemEventDisplayName(EventNameNew))
	suite.Equal("«destroy»", SystemEventDisplayName(EventNameDestroy))
	suite.Equal("Submit", SystemEventDisplayName("Submit"))
	suite.Equal("«new»", SystemEventTLAName(EventNameNew))
	suite.Equal("«new»", SystemEventTLAName("«new»"))
	suite.Equal("«destroy»", SystemEventTLAName(EventNameDestroy))
	suite.True(IsSystemEventTLAName(EventNameNew))
	suite.True(IsSystemEventTLAName("«new»"))
	suite.False(IsSystemEventTLAName("Submit"))
}
