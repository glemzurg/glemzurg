package generate

// classesMermaidStereotypeLine returns Mermaid class-annotation syntax on its own line.
// Mermaid renders <<name>> as «name» above the class title when followed by the node id.
func classesMermaidStereotypeLine(name, nodeID string) string {
	return "<<" + name + ">> " + nodeID + "\n"
}
