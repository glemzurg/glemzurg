package data_type

import (
	"testing"
)

func TestParseDataType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *DataType
		hasError bool
	}{
		{
			name:  "unconstrained string",
			input: "string",
			expected: &DataType{
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
					Details:        "string",
				},
			},
			hasError: false,
		},
		{
			name:  "unconstrained int",
			input: "int",
			expected: &DataType{
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "unconstrained",
					Details:        "int",
				},
			},
			hasError: false,
		},
		{
			name:  "enum simple",
			input: "enum {red, green, blue}",
			expected: &DataType{
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "enumeration",
					EnumValues: []AtomicEnumValue{
						{Value: "red", SortOrder: 0},
						{Value: "green", SortOrder: 0},
						{Value: "blue", SortOrder: 0},
					},
				},
			},
			hasError: false,
		},
		{
			name:     "invalid input",
			input:    "invalid type",
			expected: nil,
			hasError: true,
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
			hasError: true,
		},
		{
			name:  "enum with spaces",
			input: "enum { a , b , c }",
			expected: &DataType{
				CollectionType: "atomic",
				Atomic: &Atomic{
					ConstraintType: "enumeration",
					EnumValues: []AtomicEnumValue{
						{Value: "a", SortOrder: 0},
						{Value: "b", SortOrder: 0},
						{Value: "c", SortOrder: 0},
					},
				},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse("", []byte(tt.input))
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result == nil {
				t.Errorf("expected result but got nil")
				return
			}
			dt, ok := result.(*DataType)
			if !ok {
				t.Errorf("expected *DataType but got %T", result)
				return
			}
			if !dataTypesEqual(dt, tt.expected) {
				t.Errorf("expected %+v but got %+v", tt.expected, dt)
			}
		})
	}
}

func dataTypesEqual(a, b *DataType) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.CollectionType != b.CollectionType {
		return false
	}
	if a.Details != b.Details {
		return false
	}
	if a.Atomic == nil || b.Atomic == nil {
		return a.Atomic == b.Atomic
	}
	if a.Atomic.ConstraintType != b.Atomic.ConstraintType {
		return false
	}
	if a.Atomic.Details != b.Atomic.Details {
		return false
	}
	if len(a.Atomic.EnumValues) != len(b.Atomic.EnumValues) {
		return false
	}
	for i, ev := range a.Atomic.EnumValues {
		if ev.Value != b.Atomic.EnumValues[i].Value || ev.SortOrder != b.Atomic.EnumValues[i].SortOrder {
			return false
		}
	}
	return true
}
