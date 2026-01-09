package requirements

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func TestModelSuite(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

type ModelSuite struct {
	suite.Suite
}

func (suite *ModelSuite) TestNew() {
	tests := []struct {
		key     identity.Key
		name    string
		details string
		obj     Model
		errstr  string
	}{
		// OK.
		{
			key:     helper.Must(identity.NewRootKey("model1")),
			name:    "Name",
			details: "Details",
			obj: Model{
				Key:     helper.Must(identity.NewRootKey("model1")),
				Name:    "Name",
				Details: "Details",
			},
		},
		{
			key:     helper.Must(identity.NewRootKey("model1")),
			name:    "Name",
			details: "",
			obj: Model{
				Key:     helper.Must(identity.NewRootKey("model1")),
				Name:    "Name",
				Details: "",
			},
		},

		// Error states.
		{
			key:     identity.Key{},
			name:    "Name",
			details: "Details",
			errstr:  "Key: (subKey: cannot be blank.).",
		},
		{
			key:     helper.Must(identity.NewRootKey("model1")),
			name:    "",
			details: "Details",
			errstr:  "Name: cannot be blank.",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)

		obj, err := NewModel(test.key, test.name, test.details)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
