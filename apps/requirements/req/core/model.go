// Package core is the requirements model for other modules in this repository.
// The model implementation stays inside the requirements module. This package is
// the import path those modules use.
package core

import icore "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"

// Model is the documentation summary of a set of requirements.
type Model = icore.Model
