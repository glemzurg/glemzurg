// Package sqlgen turns a requirements model into PostgreSQL text.
// Statement shape lives in templates. PostgreSQL details the requirements model
// does not carry, such as column types and ON DELETE, come from annotation notes.
// A missing note is an error so the annotation that statement needs stays visible.
package sqlgen

import (
	"github.com/glemzurg/glemzurg/apps/code/postgresql/internal/annot"
	"github.com/glemzurg/glemzurg/apps/requirements/req/pkg/core"
)

// CreateTables renders CREATE TABLE statements and the foreign keys of to-one associations.
func CreateTables(model core.Model, notes annot.Set) (string, error) {
	tables, byClass, err := collectTables(model, notes)
	if err != nil {
		return "", err
	}
	keys, err := collectForeignKeys(model, notes, tables, byClass)
	if err != nil {
		return "", err
	}
	return render(tables, keys)
}
