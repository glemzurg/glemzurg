package generate

// classesMermaidStereotypeAnnotation returns Mermaid class-body stereotype syntax.
// Mermaid parses <<name>> as an annotation and renders it as «name» above the class title.
func classesMermaidStereotypeAnnotation(name string) string {
	return "<<" + name + ">>"
}
