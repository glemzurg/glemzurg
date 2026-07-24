package schema

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
)

// emptyModel is a blank *core.Model for tests that need a schema without surface classes.
func emptyModel() *core.Model {
	m := core.NewModel("empty", core.ModelDetails{Name: "empty", Details: ""}, "", nil, nil, nil)
	return &m
}
