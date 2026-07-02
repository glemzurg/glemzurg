package database

// associationEnd mirrors the association_end PostgreSQL enum.
type associationEnd string

const (
	associationEndFrom associationEnd = "from"
	associationEndTo   associationEnd = "to"
)

func (e associationEnd) String() string {
	return string(e)
}
