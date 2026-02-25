package object

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// Record is a string-keyed collection of fields.
type Record struct {
	fields map[string]Object
}

// NewRecord creates a new empty Record.
func NewRecord() *Record {
	return &Record{
		fields: make(map[string]Object),
	}
}

// NewRecordFromFields creates a Record with initial field values.
func NewRecordFromFields(fields map[string]Object) *Record {
	r := NewRecord()
	for name, val := range fields {
		r.fields[name] = val.Clone()
	}
	return r
}

func (r *Record) Type() ObjectType { return TypeRecord }

func (r *Record) Inspect() string {
	var out bytes.Buffer
	fieldNames := r.sortedFieldNames()

	fields := make([]string, len(fieldNames))
	for i, name := range fieldNames {
		fields[i] = fmt.Sprintf("%s |-> %s", name, r.fields[name].Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(fields, ", "))
	out.WriteString("]")
	return out.String()
}

func (r *Record) SetValue(source Object) error {
	src, ok := source.(*Record)
	if !ok {
		return fmt.Errorf("cannot assign %T to Record", source)
	}
	r.fields = make(map[string]Object)
	for name, val := range src.fields {
		r.fields[name] = val.Clone()
	}
	return nil
}

func (r *Record) Clone() Object {
	clone := NewRecord()
	for name, val := range r.fields {
		clone.fields[name] = val.Clone()
	}
	return clone
}

// Get returns the value of a field, or nil if it doesn't exist.
func (r *Record) Get(name string) Object {
	return r.fields[name]
}

// Set sets the value of a field.
func (r *Record) Set(name string, value Object) {
	r.fields[name] = value.Clone()
}

// Has returns true if the field exists.
func (r *Record) Has(name string) bool {
	_, ok := r.fields[name]
	return ok
}

// FieldNames returns the sorted list of field names.
func (r *Record) FieldNames() []string {
	return r.sortedFieldNames()
}

// Fields returns a copy of the fields map.
func (r *Record) Fields() map[string]Object {
	result := make(map[string]Object)
	for name, val := range r.fields {
		result[name] = val
	}
	return result
}

// Len returns the number of fields.
func (r *Record) Len() int {
	return len(r.fields)
}

// Equals checks if two records have the same fields with equal values.
func (r *Record) Equals(other *Record) bool {
	if len(r.fields) != len(other.fields) {
		return false
	}
	for name, val := range r.fields {
		otherVal, ok := other.fields[name]
		if !ok {
			return false
		}
		if hashValue(val) != hashValue(otherVal) {
			return false
		}
	}
	return true
}

// sortedFieldNames returns field names sorted alphabetically.
func (r *Record) sortedFieldNames() []string {
	names := make([]string, 0, len(r.fields))
	for name := range r.fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// WithField returns a new Record with the specified field updated.
func (r *Record) WithField(name string, value Object) *Record {
	clone := r.Clone().(*Record)
	clone.fields[name] = value.Clone()
	return clone
}

// Without returns a new Record without the specified field.
func (r *Record) Without(name string) *Record {
	clone := r.Clone().(*Record)
	delete(clone.fields, name)
	return clone
}
