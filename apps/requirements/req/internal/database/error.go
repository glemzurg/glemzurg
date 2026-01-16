package database

import "fmt"

const (
	_POSTGRES_NOT_FOUND = `sql: no rows in result set`
)

var ErrNotFound = fmt.Errorf("not found")
