package sqlgen

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/code/postgresql/internal/annot"
	"github.com/glemzurg/glemzurg/apps/requirements/req/pkg/core"
)

func collectTables(model core.Model, notes annot.Set) ([]tableView, map[string]string, error) {
	var tables []tableView
	byClass := map[string]string{}
	used := map[string]struct{}{}
	for dkey, domain := range model.Domains {
		for skey, sub := range domain.Subdomains {
			for ckey, class := range sub.Classes {
				walk := func(emit func(key string, nullable, derived bool)) {
					for _, attr := range class.Attributes {
						emit(attr.Key.SubKey, attr.Nullable, attr.DerivationPolicy != nil)
					}
				}
				table, skip, err := oneClassTable(dkey.SubKey, skey.SubKey, ckey.SubKey, notes, walk)
				if err != nil {
					return nil, nil, err
				}
				if skip {
					continue
				}
				if _, exists := used[table.Name]; exists {
					return nil, nil, fmt.Errorf("table %s is declared by more than one class", table.Name)
				}
				used[table.Name] = struct{}{}
				byClass[ckey.ParentKey+"/"+ckey.SubKey] = table.Name
				tables = append(tables, table)
			}
		}
	}
	// Alphabetical names keep the script stable. Foreign keys are separate statements,
	// so a referenced table does not have to be created first.
	sort.Slice(tables, func(i, j int) bool { return tables[i].Name < tables[j].Name })
	return tables, byClass, nil
}

func oneClassTable(domain, subdomain, class string, notes annot.Set, walk func(emit func(key string, nullable, derived bool))) (tableView, bool, error) {
	rel := path.Join("domains", domain, "subdomains", subdomain, "classes", class, "class.json")
	note, ok := notes.Note(rel)
	if ok && excluded(note.Include) {
		return tableView{}, true, nil
	}
	if ok {
		if err := rejectUnknown(rel, note, walk); err != nil {
			return tableView{}, false, err
		}
	}
	cols, err := columns(class, note, walk)
	if err != nil {
		return tableView{}, false, err
	}
	return newTable(class, cols), false, nil
}

func rejectUnknown(rel string, note annot.Note, walk func(emit func(key string, nullable, derived bool))) error {
	known := map[string]struct{}{}
	walk(func(key string, _ bool, _ bool) {
		known[key] = struct{}{}
	})
	var missing []string
	for key := range note.Attributes {
		if _, ok := known[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	return fmt.Errorf("%s: unknown attribute %s", rel, strings.Join(missing, ", "))
}

func columns(className string, note annot.Note, walk func(emit func(key string, nullable, derived bool))) ([]columnView, error) {
	var cols []columnView
	var err error
	// Derived attributes are computed outside the database, so they are not columns
	// and do not need a type annotation.
	walk(func(key string, nullable, derived bool) {
		if err != nil || derived {
			return
		}
		noted, ok := note.Attributes[key]
		if ok && excluded(noted.Include) {
			return
		}
		// A stored attribute has no PostgreSQL type in the requirements model.
		// The annotation must name one; an empty note is not a default type.
		if !ok || noted.Type == "" {
			err = fmt.Errorf("attribute %s on %s needs a PostgreSQL type", key, className)
			return
		}
		if key == "pk" {
			err = fmt.Errorf("attribute pk on %s collides with the synthetic primary key", className)
			return
		}
		cols = append(cols, newColumn(key, noted.Type, noted.Default, !nullable))
	})
	return cols, err
}

func excluded(include *bool) bool {
	return include != nil && !*include
}
