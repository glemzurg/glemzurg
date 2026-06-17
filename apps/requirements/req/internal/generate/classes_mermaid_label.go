package generate

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"

const classesMermaidLabelLineBreak = "<br/>"

// classesMermaidClassLabel renders stereotypes above the class name when present.
func classesMermaidClassLabel(class model_class.Class) string {
	if class.ActorKey != nil {
		return "«actor»" + classesMermaidLabelLineBreak + class.Name
	}
	return class.Name
}

// classesMermaidAssociationLinkLabel renders the association stereotype above the link name.
func classesMermaidAssociationLinkLabel(name string) string {
	return "«association»" + classesMermaidLabelLineBreak + name
}
