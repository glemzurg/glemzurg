package sqlgen

// columnView is one stored column a CREATE TABLE template fills in.
type columnView struct {
	Name    string
	Type    string
	NotNull bool
	Default string
}

// tableView is one CREATE TABLE template.
type tableView struct {
	Name    string
	Columns []columnView
}

// foreignKeyView is the ALTER TABLE template for one to-one association end.
type foreignKeyView struct {
	Table      string
	Column     string
	RefTable   string
	OnDelete   string
	Constraint string
	NotNull    bool
}

// classRef is one association end, reduced to the fields the SQL templates need.
type classRef struct {
	parent string
	sub    string
	lower  uint
	higher uint
}

func newColumn(name, pgType, defaultSQL string, notNull bool) columnView {
	return columnView{Name: name, Type: pgType, NotNull: notNull, Default: defaultSQL}
}

func newTable(name string, columns []columnView) tableView {
	return tableView{Name: name, Columns: columns}
}

func newForeignKey(table, column, refTable, onDelete string, notNull bool) foreignKeyView {
	return foreignKeyView{
		Table:      table,
		Column:     column,
		RefTable:   refTable,
		OnDelete:   onDelete,
		Constraint: "fk_" + table + "_" + refTable,
		NotNull:    notNull,
	}
}

func newClassRef(parent, sub string, lower, higher uint) classRef {
	return classRef{parent: parent, sub: sub, lower: lower, higher: higher}
}

func (c classRef) id() string {
	return c.parent + "/" + c.sub
}

// toOne is a single object (higher bound 1). Zero means unbounded, so it is not to-one.
func (c classRef) toOne() bool {
	return c.higher == 1
}

// many is unbounded. A bounded range above 1 is neither many nor to-one.
func (c classRef) many() bool {
	return c.higher == 0
}
