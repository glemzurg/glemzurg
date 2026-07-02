package generate

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
)

const classesMermaidNoteLineBreak = "<br>"

// classesMermaidAssociationLinkNote formats a note on a dashed association link node.
func classesMermaidAssociationLinkNote(assoc model_class.Association) string {
	if !renderAssociationLinkNodeMermaid(assoc) {
		return ""
	}
	return classesMermaidNoteLine(mermaidNodeID("assoc", assoc.Key), assoc.UmlComment)
}

// classesMermaidNoteLine formats a single-line Mermaid class-diagram note.
func classesMermaidNoteLine(target, comment string) string {
	text := classesMermaidNoteText(comment)
	if text == "" || strings.TrimSpace(target) == "" {
		return ""
	}
	return "note for " + target + ` "` + text + "\"\n"
}

// classesMermaidNoteText converts uml_comment prose into Mermaid note label text.
func classesMermaidNoteText(comment string) string {
	trimmed := strings.TrimSpace(comment)
	if trimmed == "" {
		return ""
	}
	lines := strings.Split(trimmed, "\n")
	parts := make([]string, len(lines))
	for i, line := range lines {
		parts[i] = strings.TrimRight(line, " \t")
	}
	return classesMermaidEscapeNoteText(strings.Join(parts, classesMermaidNoteLineBreak))
}

func classesMermaidEscapeNoteText(text string) string {
	return strings.ReplaceAll(text, `"`, "#quot;")
}
