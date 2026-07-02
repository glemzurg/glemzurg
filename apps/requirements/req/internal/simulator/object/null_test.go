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
