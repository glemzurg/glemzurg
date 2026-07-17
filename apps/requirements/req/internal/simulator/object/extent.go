package object

import "math"

// Extent record fields for class instances and association endpoint images.
// Class extents and association navigations use [id |-> N, data |-> attrs] so
// instances with identical attribute data stay distinct in TLA sets.
const (
	ExtentIDField   = "id"
	ExtentDataField = "data"
)

// NewExtentElement builds [id |-> id, data |-> attrs].
// data is cloned so evaluation cannot mutate the caller's attribute record.
func NewExtentElement(id uint64, attrs *Record) *Record {
	data := attrs
	if data != nil {
		data = data.Clone().(*Record)
	} else {
		data = NewRecord()
	}
	return NewRecordFromFields(map[string]Object{
		ExtentIDField:   NewNatural(extentIDAsInt64(id)),
		ExtentDataField: data,
	})
}

func extentIDAsInt64(id uint64) int64 {
	if id > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(id)
}

// IsExtentElement reports whether r is a [id, data] extent package.
func IsExtentElement(r *Record) bool {
	if r == nil || !r.Has(ExtentIDField) || !r.Has(ExtentDataField) {
		return false
	}
	if _, ok := r.Get(ExtentIDField).(*Number); !ok {
		return false
	}
	data, ok := r.Get(ExtentDataField).(*Record)
	return ok && data != nil
}

// ExtentData returns the data record from an extent element, or r when r is flat.
func ExtentData(r *Record) *Record {
	if r == nil {
		return nil
	}
	if data, ok := r.Get(ExtentDataField).(*Record); ok && data != nil {
		return data
	}
	return r
}

// ExtentID returns the engine id from an extent element.
func ExtentID(r *Record) (uint64, bool) {
	if r == nil {
		return 0, false
	}
	idVal := r.Get(ExtentIDField)
	if idVal == nil {
		return 0, false
	}
	n, ok := idVal.(*Number)
	if !ok || n.Sign() < 0 {
		return 0, false
	}
	v := n.Rat().Num().Int64()
	if v < 0 {
		return 0, false
	}
	return uint64(v), true
}

// RecordField returns a field from r, or from r.data when r is an extent element
// and the field is not id/data. Supports Approach A attribute access on peers
// that appear as [id, data] in association images.
func RecordField(r *Record, field string) (Object, bool) {
	if r == nil {
		return nil, false
	}
	if r.Has(field) {
		return r.Get(field), true
	}
	if field == ExtentIDField || field == ExtentDataField {
		return nil, false
	}
	if !IsExtentElement(r) {
		return nil, false
	}
	data := ExtentData(r)
	if data == nil || !data.Has(field) {
		return nil, false
	}
	return data.Get(field), true
}

// RecordHasField reports whether RecordField would succeed.
func RecordHasField(r *Record, field string) bool {
	_, ok := RecordField(r, field)
	return ok
}
