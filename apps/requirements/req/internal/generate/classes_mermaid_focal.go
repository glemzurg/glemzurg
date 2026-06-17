package generate

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// classesMermaidFocalClassStyle is the Mermaid style for the class whose page is being viewed.
const classesMermaidFocalClassStyle = "stroke:#9370DB,stroke-width:3px"

func hasMermaidFocalClass(focalClassKey *identity.Key) bool {
	return focalClassKey != nil
}

func mermaidFocalClassKey(focalClassKey *identity.Key) identity.Key {
	return *focalClassKey
}
