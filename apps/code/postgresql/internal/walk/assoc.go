package walk

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// classEnd is one class referenced by an association filename.
type classEnd struct {
	parent string
	sub    string
}

func (c *context) associations(level, dir, labelPrefix string, walkFn func(emit func(fromParent, fromSub, toParent, toSub, name string))) ([]AssocVisit, error) {
	var visits []AssocVisit
	var emitErr error
	walkFn(func(fromParent, fromSub, toParent, toSub, name string) {
		if emitErr != nil {
			return
		}
		from := classEnd{parent: fromParent, sub: fromSub}
		to := classEnd{parent: toParent, sub: toSub}
		visit, err := c.oneAssociation(level, dir, labelPrefix, from, to, name)
		if err != nil {
			emitErr = err
			return
		}
		visits = append(visits, visit)
	})
	if emitErr != nil {
		return nil, emitErr
	}
	sort.Slice(visits, func(i, j int) bool { return visits[i].Label < visits[j].Label })
	return visits, nil
}

func (c *context) oneAssociation(level, dir, labelPrefix string, from, to classEnd, name string) (AssocVisit, error) {
	fromName, err := dottedClass(level, from.parent, from.sub)
	if err != nil {
		return AssocVisit{}, err
	}
	toName, err := dottedClass(level, to.parent, to.sub)
	if err != nil {
		return AssocVisit{}, err
	}
	stem := fromName + "--" + toName + "--" + name
	rel := path.Join(dir, stem+".assoc.json")
	visit := AssocVisit{Label: labelPrefix + stem}
	note, ok := c.bind(rel)
	if ok {
		visit.OnDelete = note.OnDelete
		visit.Include = copyBool(note.Include)
	}
	return visit, nil
}

func dottedClass(level, parentKey, subKey string) (string, error) {
	domain, subdomain, class, err := splitClassParent(parentKey, subKey)
	if err != nil {
		return "", err
	}
	switch level {
	case "subdomain":
		return class, nil
	case "domain":
		return subdomain + "." + class, nil
	case "model":
		return domain + "." + subdomain + "." + class, nil
	default:
		return "", fmt.Errorf("association level %q is unknown", level)
	}
}

func splitClassParent(parentKey, subKey string) (string, string, string, error) {
	parts := strings.Split(parentKey, "/")
	if len(parts) != 4 || parts[0] != "domain" || parts[2] != "subdomain" || parts[1] == "" || parts[3] == "" || subKey == "" {
		return "", "", "", fmt.Errorf("class key %s/%s is not a domain/subdomain/class key", parentKey, subKey)
	}
	return parts[1], parts[3], subKey, nil
}
