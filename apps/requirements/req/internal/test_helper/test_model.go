package testhelper

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"

// Create a very elaborate model that can be used for testing in various packages around the system.
func GetTestModel() req_model.Model {
	return req_model.Model{}
}
