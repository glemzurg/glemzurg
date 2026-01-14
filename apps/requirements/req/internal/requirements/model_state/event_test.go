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

func (suite *EventSuite) TestNew() {

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))

	tests := []struct {
		testName   string
		key        identity.Key
		name       string
		details    string
		parameters []EventParameter
		obj        Event
		errstr     string
	}{
		// OK.
		{
			testName:   "ok with all fields",
			key:        helper.Must(identity.NewEventKey(classKey, "event1")),
			name:       "Name",
			details:    "Details",
			parameters: []EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			obj: Event{
				Key:        helper.Must(identity.NewEventKey(classKey, "event1")),
				Name:       "Name",
				Details:    "Details",
				Parameters: []EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			},
		},
		{
			testName:   "ok with minimal fields",
			key:        helper.Must(identity.NewEventKey(classKey, "event2")),
			name:       "Name",
			details:    "",
			parameters: nil,
			obj: Event{
				Key:        helper.Must(identity.NewEventKey(classKey, "event2")),
				Name:       "Name",
				Details:    "",
				Parameters: nil,
			},
		},

		// Error states.
		{
			testName:   "error empty key",
			key:        identity.Key{},
			name:       "Name",
			details:    "Details",
			parameters: []EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			errstr:     "keyType: cannot be blank",
		},
		{
			testName:   "error wrong key type",
			key:        helper.Must(identity.NewDomainKey("domain1")),
			name:       "Name",
			details:    "Details",
			parameters: []EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			errstr:     "Key: invalid key type 'domain' for event",
		},
		{
			testName:   "error with blank name",
			key:        helper.Must(identity.NewEventKey(classKey, "event3")),
			name:       "",
			details:    "Details",
			parameters: []EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			errstr:     `Name: cannot be blank`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewEvent(tt.key, tt.name, tt.details, tt.parameters)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.obj, obj)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Empty(t, obj)
			}
		})
	}
}
