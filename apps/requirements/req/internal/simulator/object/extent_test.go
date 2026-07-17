package object

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtentElementsDistinctInSetWhenDataEqual(t *testing.T) {
	t.Parallel()
	a := NewExtentElement(1, NewRecordFromFields(map[string]Object{"_state": NewString("Exists")}))
	b := NewExtentElement(2, NewRecordFromFields(map[string]Object{"_state": NewString("Exists")}))
	set := NewSet()
	set.Add(a)
	set.Add(b)
	require.Equal(t, 2, set.Size())
}

func TestRecordFieldProjectsThroughData(t *testing.T) {
	t.Parallel()
	ext := NewExtentElement(3, NewRecordFromFields(map[string]Object{
		"_state": NewString("Exists"),
		"amount": NewInteger(42),
	}))
	val, ok := RecordField(ext, "amount")
	require.True(t, ok)
	require.Equal(t, "42", val.Inspect())
	id, ok := RecordField(ext, ExtentIDField)
	require.True(t, ok)
	require.Equal(t, "3", id.Inspect())
}
