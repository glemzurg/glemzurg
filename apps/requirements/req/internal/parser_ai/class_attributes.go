package parser_ai

import (
	"fmt"
	"path/filepath"
	"strings"
)

// attributeKeysInOrder returns attribute keys in declaration order.
func (c *inputClass) attributeKeysInOrder() []string {
	keys := make([]string, 0, len(c.Attributes))
	for _, attr := range c.Attributes {
		keys = append(keys, attr.Key)
	}
	return keys
}

func (c *inputClass) hasAttributeKey(key string) bool {
	_, ok := c.attributeByKey(key)
	return ok
}

func (c *inputClass) attributeByKey(key string) (inputAttribute, bool) {
	for _, attr := range c.Attributes {
		if attr.Key == key {
			return attr, true
		}
	}
	return inputAttribute{}, false
}

// safeAttributeDirKey rejects attribute keys that would escape the attributes/ subtree.
func safeAttributeDirKey(key string) (string, error) {
	if key == "" || key != filepath.Base(key) || strings.Contains(key, "..") {
		return "", fmt.Errorf("invalid attribute key %q for filesystem path", key)
	}
	return key, nil
}
