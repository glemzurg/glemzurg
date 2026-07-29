package object

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeSimulatorValue(t *testing.T) {
	t.Run("empty string becomes null", func(t *testing.T) {
		normalized := NormalizeSimulatorValue(NewString(""))
		require.IsType(t, &Set{}, normalized)
		require.True(t, IsNull(normalized))
	})

	t.Run("non-empty string unchanged", func(t *testing.T) {
		normalized := NormalizeSimulatorValue(NewString("hello"))
		require.Equal(t, "hello", normalized.(*String).Value())
	})

	t.Run("nil unchanged", func(t *testing.T) {
		require.Nil(t, NormalizeSimulatorValue(nil))
	})
}

func TestRecordSetNormalizesEmptyString(t *testing.T) {
	record := NewRecord()
	record.Set("name", NewString(""))
	require.True(t, IsNull(record.Get("name")))
}

func TestCoerceToNumber(t *testing.T) {
	t.Run("number passthrough", func(t *testing.T) {
		n, ok := CoerceToNumber(NewInteger(7))
		require.True(t, ok)
		require.Equal(t, 0, n.Cmp(NewInteger(7)))
	})
	t.Run("integer string", func(t *testing.T) {
		n, ok := CoerceToNumber(NewString("-42"))
		require.True(t, ok)
		require.Equal(t, 0, n.Cmp(NewInteger(-42)))
	})
	t.Run("rational string", func(t *testing.T) {
		n, ok := CoerceToNumber(NewString("3/2"))
		require.True(t, ok)
		require.Equal(t, 0, n.Cmp(NewRational(3, 2)))
	})
	t.Run("non-numeric string", func(t *testing.T) {
		_, ok := CoerceToNumber(NewString("abc"))
		require.False(t, ok)
	})
}
