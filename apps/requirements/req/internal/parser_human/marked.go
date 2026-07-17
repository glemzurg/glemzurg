package parser_human

import (
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// parseMarkedClassSubKeys reads a this.marked file: a YAML list of class subkeys.
func parseMarkedClassSubKeys(filename, contents string) ([]string, error) {
	trimmed := strings.TrimSpace(contents)
	if trimmed == "" {
		return nil, nil
	}

	var keys []string
	if err := yaml.Unmarshal([]byte(contents), &keys); err != nil {
		return nil, errors.Wrapf(err, "failed to parse marked class list in %s", filename)
	}
	return keys, nil
}

// applyMarkedClassSubKeys sets Marked=true on subdomain classes listed by subkey.
// Unknown subkeys are an error so typos surface at parse time.
func applyMarkedClassSubKeys(subdomainKey identity.Key, classes map[identity.Key]model_class.Class, subKeys []string, filename string) (map[identity.Key]model_class.Class, error) {
	if len(subKeys) == 0 {
		return classes, nil
	}
	if classes == nil {
		classes = make(map[identity.Key]model_class.Class)
	}
	for _, subKey := range subKeys {
		subKey = strings.TrimSpace(subKey)
		if subKey == "" {
			return nil, errors.Errorf("%s: marked class list contains an empty entry", filename)
		}
		classKey, err := identity.NewClassKey(subdomainKey, subKey)
		if err != nil {
			return nil, errors.Wrapf(err, "%s: invalid marked class subkey %q", filename, subKey)
		}
		class, ok := classes[classKey]
		if !ok {
			return nil, errors.Errorf("%s: marked class %q not found in subdomain", filename, subKey)
		}
		class.SetMarked(true)
		classes[classKey] = class
	}
	return classes, nil
}

// generateMarkedContent builds this.marked YAML from marked classes in the subdomain.
// Returns empty string when no class is marked so callers can omit the file.
func generateMarkedContent(classes map[identity.Key]model_class.Class) string {
	var keys []string
	for _, class := range classes {
		if class.Marked {
			keys = append(keys, class.Key.SubKey)
		}
	}
	if len(keys) == 0 {
		return ""
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, key := range keys {
		b.WriteString("- ")
		b.WriteString(key)
		b.WriteString("\n")
	}
	return b.String()
}
