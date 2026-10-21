package parser_json

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
)

func TestWriteReadTestModel(t *testing.T) {
	model := test_helper.GetTestModel()

	first := writeModel(t, model)
	second := writeModel(t, model)
	assert.Equal(t, first, second)
	assert.Contains(t, string(first), `"format":"core-pipe"`)

	got := readModel(t, first)
	require.NoError(t, got.Validate())
	assert.Equal(t, authoredView(model), authoredView(got))
}

// authoredView drops caches the pipe does not carry and treats an empty
// collection as the zero the reader rebuilds from an omitted field.
func authoredView(model core.Model) core.Model {
	clearAuthoredNoise(reflect.ValueOf(&model))
	return model
}

func clearAuthoredNoise(v reflect.Value) {
	// Only composite values can nest expression trees or parsed data types.
	switch v.Kind() { //nolint:exhaustive // scalars hold no authored caches
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			clearAuthoredNoise(v.Elem())
		}
	case reflect.Struct:
		for i := range v.NumField() {
			field := v.Field(i)
			if !field.CanSet() {
				continue
			}
			switch v.Type().Field(i).Name {
			case "Expression", "ExpressionType":
				field.Set(reflect.Zero(field.Type()))
			case "DataType":
				if field.Kind() == reflect.Pointer {
					field.Set(reflect.Zero(field.Type()))
					continue
				}
				clearAuthoredNoise(field)
			default:
				clearAuthoredNoise(field)
			}
		}
	case reflect.Slice:
		if v.IsNil() {
			return
		}
		if v.Len() == 0 {
			if v.CanSet() {
				v.Set(reflect.Zero(v.Type()))
			}
			return
		}
		for i := range v.Len() {
			clearAuthoredNoise(v.Index(i))
		}
	case reflect.Map:
		if v.IsNil() {
			return
		}
		if v.Len() == 0 {
			if v.CanSet() {
				v.Set(reflect.Zero(v.Type()))
			}
			return
		}
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key)
			clone := reflect.New(val.Type()).Elem()
			clone.Set(val)
			clearAuthoredNoise(clone)
			v.SetMapIndex(key, clone)
		}
	}
}
