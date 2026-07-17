package generate

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// classesMermaidFocalClassStyle is the Mermaid stroke style for the class whose page is being viewed.
const classesMermaidFocalClassStyle = "stroke:#9370DB,stroke-width:3px"

// classesMermaidMarkedClassStyle is the Mermaid fill for classes with Marked=true (default boxes render blue).
const classesMermaidMarkedClassStyle = "fill:#FFEB3B"

func hasMermaidFocalClass(focalClassKey *identity.Key) bool {
	return focalClassKey != nil
}

func mermaidFocalClassKey(focalClassKey *identity.Key) identity.Key {
	return *focalClassKey
}

// classesMermaidClassBoxStyle returns combined Mermaid style properties for a class box, or empty for default.
// Marked classes get a yellow fill; the focal class (class page under view) keeps a stronger border.
func classesMermaidClassBoxStyle(class model_class.Class, focalClassKey *identity.Key) string {
	var parts []string
	if class.Marked {
		parts = append(parts, classesMermaidMarkedClassStyle)
	}
	if focalClassKey != nil && class.Key == *focalClassKey {
		parts = append(parts, classesMermaidFocalClassStyle)
	}
	return strings.Join(parts, ",")
}
