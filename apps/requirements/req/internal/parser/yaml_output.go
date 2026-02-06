package parser

import (
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// YamlBuilder helps construct YAML output with proper formatting for multiline strings
// while maintaining deterministic field ordering for testability.
type YamlBuilder struct {
	node *yaml.Node
}

// NewYamlBuilder creates a new YamlBuilder with a mapping node.
func NewYamlBuilder() *YamlBuilder {
	return &YamlBuilder{
		node: &yaml.Node{
			Kind: yaml.MappingNode,
		},
	}
}

// AddField adds a key-value pair to the YAML output.
// Handles multiline strings using literal block scalar (|) notation.
func (b *YamlBuilder) AddField(key string, value string) {
	if value == "" {
		return
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	valueNode := createStringNode(value)

	b.node.Content = append(b.node.Content, keyNode, valueNode)
}

// AddFieldAlways adds a key-value pair even if empty.
func (b *YamlBuilder) AddFieldAlways(key string, value string) {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	valueNode := createStringNode(value)

	b.node.Content = append(b.node.Content, keyNode, valueNode)
}

// AddQuotedField adds a key-value pair with quotes forced.
func (b *YamlBuilder) AddQuotedField(key string, value string) {
	if value == "" {
		return
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
		Style: yaml.DoubleQuotedStyle,
	}

	b.node.Content = append(b.node.Content, keyNode, valueNode)
}

// AddBoolField adds a boolean field.
func (b *YamlBuilder) AddBoolField(key string, value bool) {
	if !value {
		return
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "true",
		Tag:   "!!bool",
	}

	b.node.Content = append(b.node.Content, keyNode, valueNode)
}

// AddIntSliceField adds a field with an integer slice value.
func (b *YamlBuilder) AddIntSliceField(key string, values []int) {
	if len(values) == 0 {
		return
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	seqNode := &yaml.Node{
		Kind:  yaml.SequenceNode,
		Style: yaml.FlowStyle,
	}
	for _, v := range values {
		seqNode.Content = append(seqNode.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: intToString(v),
			Tag:   "!!int",
		})
	}

	b.node.Content = append(b.node.Content, keyNode, seqNode)
}

// AddUintSliceField adds a field with an unsigned integer slice value.
func (b *YamlBuilder) AddUintSliceField(key string, values []uint) {
	if len(values) == 0 {
		return
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	seqNode := &yaml.Node{
		Kind:  yaml.SequenceNode,
		Style: yaml.FlowStyle,
	}
	for _, v := range values {
		seqNode.Content = append(seqNode.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: strconv.FormatUint(uint64(v), 10),
			Tag:   "!!int",
		})
	}

	b.node.Content = append(b.node.Content, keyNode, seqNode)
}

// AddSequenceField adds a field with a sequence of string values.
// Uses literal block scalar for multiline items.
func (b *YamlBuilder) AddSequenceField(key string, values []string) {
	if len(values) == 0 {
		return
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	seqNode := &yaml.Node{
		Kind: yaml.SequenceNode,
	}
	for _, v := range values {
		seqNode.Content = append(seqNode.Content, createStringNode(v))
	}

	b.node.Content = append(b.node.Content, keyNode, seqNode)
}

// AddMappingField adds a field with a nested mapping value.
// Skips if nested is nil or empty.
func (b *YamlBuilder) AddMappingField(key string, nested *YamlBuilder) {
	if nested == nil || len(nested.node.Content) == 0 {
		return
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	b.node.Content = append(b.node.Content, keyNode, nested.node)
}

// AddMappingFieldAlways adds a field with a nested mapping value, even if empty.
func (b *YamlBuilder) AddMappingFieldAlways(key string, nested *YamlBuilder) {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	if nested == nil {
		nested = NewYamlBuilder()
	}

	b.node.Content = append(b.node.Content, keyNode, nested.node)
}

// AddSequenceOfMappings adds a field with a sequence of mapping values.
func (b *YamlBuilder) AddSequenceOfMappings(key string, items []*YamlBuilder) {
	if len(items) == 0 {
		return
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	seqNode := &yaml.Node{
		Kind: yaml.SequenceNode,
	}
	for _, item := range items {
		if item != nil && len(item.node.Content) > 0 {
			seqNode.Content = append(seqNode.Content, item.node)
		}
	}

	if len(seqNode.Content) == 0 {
		return
	}

	b.node.Content = append(b.node.Content, keyNode, seqNode)
}

// AddFlowMapping adds a field with a flow-style mapping (inline).
func (b *YamlBuilder) AddFlowMapping(key string, nested *YamlBuilder) {
	if nested == nil || len(nested.node.Content) == 0 {
		return
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	flowNode := &yaml.Node{
		Kind:    yaml.MappingNode,
		Style:   yaml.FlowStyle,
		Content: nested.node.Content,
	}

	b.node.Content = append(b.node.Content, keyNode, flowNode)
}

// AddRawNode adds a pre-built yaml.Node to the builder.
func (b *YamlBuilder) AddRawNode(key string, node *yaml.Node) {
	if node == nil {
		return
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	b.node.Content = append(b.node.Content, keyNode, node)
}

// AddFlowSequence adds a field with a sequence of flow-style mappings.
// Each item in the sequence is rendered in flow style (e.g., {key: "value", key2: "value2"}).
func (b *YamlBuilder) AddFlowSequence(key string, items []*YamlBuilder) {
	if len(items) == 0 {
		return
	}

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}

	seqNode := &yaml.Node{
		Kind: yaml.SequenceNode,
	}
	for _, item := range items {
		if item != nil && len(item.node.Content) > 0 {
			flowNode := &yaml.Node{
				Kind:    yaml.MappingNode,
				Style:   yaml.FlowStyle,
				Content: item.node.Content,
			}
			seqNode.Content = append(seqNode.Content, flowNode)
		}
	}

	if len(seqNode.Content) == 0 {
		return
	}

	b.node.Content = append(b.node.Content, keyNode, seqNode)
}

// Build marshals the YAML node to a string.
func (b *YamlBuilder) Build() (string, error) {
	if len(b.node.Content) == 0 {
		return "", nil
	}

	// Wrap in a document node for proper marshalling
	doc := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{b.node},
	}

	data, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// Node returns the underlying yaml.Node for advanced manipulation.
func (b *YamlBuilder) Node() *yaml.Node {
	return b.node
}

// HasContent returns true if any content has been added.
func (b *YamlBuilder) HasContent() bool {
	return len(b.node.Content) > 0
}

// createStringNode creates a yaml.Node for a string value.
// Uses literal block scalar (|) for multiline strings.
func createStringNode(value string) *yaml.Node {
	if strings.Contains(value, "\n") {
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: value,
			Style: yaml.LiteralStyle,
		}
	}
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
	}
}

// intToString converts an int to its string representation.
func intToString(i int) string {
	return strconv.Itoa(i)
}

// formatYamlField formats a YAML field with proper multiline handling.
// For single-line values, outputs: "key: value\n"
// For multiline values, outputs using literal block scalar (|):
//
//	key: |
//	    line1
//	    line2
//
// The indent parameter specifies the base indentation for the key.
// The contentIndent is calculated as indent + 4.
func formatYamlField(key, value string, indent int) string {
	if value == "" {
		return ""
	}

	indentStr := strings.Repeat(" ", indent)

	if strings.Contains(value, "\n") {
		// Use literal block scalar for multi-line strings
		result := indentStr + key + ": |\n"
		contentIndent := strings.Repeat(" ", indent+4)
		lines := strings.Split(value, "\n")
		for _, line := range lines {
			result += contentIndent + line + "\n"
		}
		return result
	}

	// Single line - simple format
	return indentStr + key + ": " + value + "\n"
}

// formatYamlFieldQuoted formats a YAML field with quoted value.
// Always quotes the value, even if it doesn't need it.
func formatYamlFieldQuoted(key, value string, indent int) string {
	if value == "" {
		return ""
	}
	indentStr := strings.Repeat(" ", indent)
	return indentStr + key + ": \"" + value + "\"\n"
}
