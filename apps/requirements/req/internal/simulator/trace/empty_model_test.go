package trace

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

// emptyModel is a blank *core.Model for tests that need a schema without surface classes.
func emptyModel() *core.Model {
	m := core.NewModel("empty", core.ModelDetails{Name: "empty", Details: ""}, "", nil, nil, nil)
	return &m
}

// emptySchema builds schema.New(emptyModel(), schema.RunScopeAll()) for test simulation state.
func emptySchema() *schema.Schema {
	return schema.New(emptyModel(), schema.RunScopeAll())
}
