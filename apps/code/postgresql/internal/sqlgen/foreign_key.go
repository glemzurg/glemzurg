package sqlgen

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/code/postgresql/internal/annot"
	"github.com/glemzurg/glemzurg/apps/requirements/req/pkg/core"
)

// knownTables is the set of tables already emitted, used while placing foreign keys.
type knownTables struct {
	byClass  map[string]string
	occupied map[string]map[string]struct{}
}

func collectForeignKeys(model core.Model, notes annot.Set, tables []tableView, byClass map[string]string) ([]foreignKeyView, error) {
	known := knownTables{byClass: byClass, occupied: occupy(tables)}
	var keys []foreignKeyView
	err := eachAssociation(model, func(rel, name string, from, to classRef) error {
		key, skip, err := place(rel, name, from, to, notes, known)
		if err != nil || skip {
			return err
		}
		keys = append(keys, key)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Table != keys[j].Table {
			return keys[i].Table < keys[j].Table
		}
		return keys[i].Column < keys[j].Column
	})
	return keys, nil
}

func occupy(tables []tableView) map[string]map[string]struct{} {
	occupied := make(map[string]map[string]struct{}, len(tables))
	for _, table := range tables {
		cols := map[string]struct{}{"pk": {}}
		for _, col := range table.Columns {
			cols[col.Name] = struct{}{}
		}
		occupied[table.Name] = cols
	}
	return occupied
}

func eachAssociation(model core.Model, emit func(rel, name string, from, to classRef) error) error {
	if err := modelAssociations(model, emit); err != nil {
		return err
	}
	if err := domainAssociations(model, emit); err != nil {
		return err
	}
	return subdomainAssociations(model, emit)
}

func modelAssociations(model core.Model, emit func(rel, name string, from, to classRef) error) error {
	dir := "class_associations"
	for _, assoc := range model.ClassAssociations {
		from := newClassRef(assoc.FromClassKey.ParentKey, assoc.FromClassKey.SubKey, assoc.FromMultiplicity.LowerBound, assoc.FromMultiplicity.HigherBound)
		to := newClassRef(assoc.ToClassKey.ParentKey, assoc.ToClassKey.SubKey, assoc.ToMultiplicity.LowerBound, assoc.ToMultiplicity.HigherBound)
		stem, err := scopedStem("model", from, to, assoc.Key.SubKey3)
		if err != nil {
			return err
		}
		if err := consider(path.Join(dir, stem+".assoc.json"), assoc.Key.SubKey3, from, to, assoc.AssociationClassKey != nil, emit); err != nil {
			return err
		}
	}
	return nil
}

func domainAssociations(model core.Model, emit func(rel, name string, from, to classRef) error) error {
	for dkey, domain := range model.Domains {
		dir := path.Join("domains", dkey.SubKey, "class_associations")
		for _, assoc := range domain.ClassAssociations {
			from := newClassRef(assoc.FromClassKey.ParentKey, assoc.FromClassKey.SubKey, assoc.FromMultiplicity.LowerBound, assoc.FromMultiplicity.HigherBound)
			to := newClassRef(assoc.ToClassKey.ParentKey, assoc.ToClassKey.SubKey, assoc.ToMultiplicity.LowerBound, assoc.ToMultiplicity.HigherBound)
			stem, err := scopedStem("domain", from, to, assoc.Key.SubKey3)
			if err != nil {
				return err
			}
			if err := consider(path.Join(dir, stem+".assoc.json"), assoc.Key.SubKey3, from, to, assoc.AssociationClassKey != nil, emit); err != nil {
				return err
			}
		}
	}
	return nil
}

func subdomainAssociations(model core.Model, emit func(rel, name string, from, to classRef) error) error {
	for dkey, domain := range model.Domains {
		for skey, sub := range domain.Subdomains {
			dir := path.Join("domains", dkey.SubKey, "subdomains", skey.SubKey, "class_associations")
			for _, assoc := range sub.ClassAssociations {
				from := newClassRef(assoc.FromClassKey.ParentKey, assoc.FromClassKey.SubKey, assoc.FromMultiplicity.LowerBound, assoc.FromMultiplicity.HigherBound)
				to := newClassRef(assoc.ToClassKey.ParentKey, assoc.ToClassKey.SubKey, assoc.ToMultiplicity.LowerBound, assoc.ToMultiplicity.HigherBound)
				stem := assoc.FromClassKey.SubKey + "--" + assoc.ToClassKey.SubKey + "--" + assoc.Key.SubKey3
				if err := consider(path.Join(dir, stem+".assoc.json"), assoc.Key.SubKey3, from, to, assoc.AssociationClassKey != nil, emit); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func consider(rel, name string, from, to classRef, associationClass bool, emit func(rel, name string, from, to classRef) error) error {
	if associationClass {
		return fmt.Errorf("association %s is an association class; its table is not generated yet", name)
	}
	return emit(rel, name, from, to)
}

func place(rel, name string, from, to classRef, notes annot.Set, known knownTables) (foreignKeyView, bool, error) {
	note, ok := notes.Note(rel)
	if ok && excluded(note.Include) {
		return foreignKeyView{}, true, nil
	}
	if note.OnDelete == "" {
		return foreignKeyView{}, false, fmt.Errorf("association %s needs an on_delete annotation", name)
	}
	key, err := foreignKey(name, from, to, note.OnDelete, known)
	if err != nil {
		return foreignKeyView{}, false, err
	}
	cols := known.occupied[key.Table]
	if _, exists := cols[key.Column]; exists {
		return foreignKeyView{}, false, fmt.Errorf("foreign key column %s already exists on %s", key.Column, key.Table)
	}
	cols[key.Column] = struct{}{}
	return key, false, nil
}

func foreignKey(name string, from, to classRef, onDelete string, known knownTables) (foreignKeyView, error) {
	switch {
	case from.toOne() && to.many():
		return link(name, to, from, from.lower >= 1, onDelete, known)
	case to.toOne() && from.many():
		return link(name, from, to, to.lower >= 1, onDelete, known)
	case from.toOne() && to.toOne():
		// Both ends are a single object, so either table could hold the foreign key.
		return foreignKeyView{}, fmt.Errorf("association %s is one-to-one; an annotation must say which end holds the foreign key", name)
	case from.many() && to.many():
		// Neither end is a single object, so the link is a join table rather than a column.
		return foreignKeyView{}, fmt.Errorf("association %s is many-to-many; a join table is not generated yet", name)
	default:
		return foreignKeyView{}, fmt.Errorf("association %s has a multiplicity that is not a single foreign key", name)
	}
}

func link(name string, holder, referenced classRef, notNull bool, onDelete string, known knownTables) (foreignKeyView, error) {
	table, ok := known.byClass[holder.id()]
	if !ok {
		return foreignKeyView{}, fmt.Errorf("association %s needs table %s", name, holder.sub)
	}
	ref, ok := known.byClass[referenced.id()]
	if !ok {
		return foreignKeyView{}, fmt.Errorf("association %s references %s, which has no table", name, referenced.sub)
	}
	return newForeignKey(table, ref+"_pk", ref, onDelete, notNull), nil
}

func scopedStem(level string, from, to classRef, name string) (string, error) {
	fromName, err := scopedClass(level, from)
	if err != nil {
		return "", err
	}
	toName, err := scopedClass(level, to)
	if err != nil {
		return "", err
	}
	return fromName + "--" + toName + "--" + name, nil
}

func scopedClass(level string, ref classRef) (string, error) {
	domain, subdomain, class, err := splitClassParent(ref.parent, ref.sub)
	if err != nil {
		return "", err
	}
	switch level {
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
