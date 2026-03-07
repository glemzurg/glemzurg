package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogicEquality_String(t *testing.T) {
	tests := []struct {
		name     string
		expr     *LogicEquality
		expected string
	}{
		{
			name: "equal",
			expr: &LogicEquality{
				Operator: EqualityOperatorEqual,
				Left:     NewIntLiteral(1),
				Right:    NewIntLiteral(2),
			},
			expected: "1 = 2",
		},
		{
			name: "not equal",
			expr: &LogicEquality{
				Operator: EqualityOperatorNotEqual,
				Left:     NewIntLiteral(1),
				Right:    NewIntLiteral(2),
			},
			expected: "1 ≠ 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.expr.String())
		})
	}
}

func TestLogicEquality_ASCII(t *testing.T) {
	tests := []struct {
		name     string
		expr     *LogicEquality
		expected string
	}{
		{
			name: "equal",
			expr: &LogicEquality{
				Operator: EqualityOperatorEqual,
				Left:     NewIntLiteral(1),
				Right:    NewIntLiteral(2),
			},
			expected: "1 = 2",
		},
		{
			name: "not equal",
			expr: &LogicEquality{
				Operator: EqualityOperatorNotEqual,
				Left:     NewIntLiteral(1),
				Right:    NewIntLiteral(2),
			},
			expected: "1 /= 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.expr.ASCII())
		})
	}
}

func TestLogicEquality_Validate(t *testing.T) {
	tests := []struct {
		name    string
		expr    *LogicEquality
		wantErr bool
	}{
		{
			name: "valid equal",
			expr: &LogicEquality{
				Operator: EqualityOperatorEqual,
				Left:     NewIntLiteral(1),
				Right:    NewIntLiteral(2),
			},
			wantErr: false,
		},
		{
			name: "valid not equal",
			expr: &LogicEquality{
				Operator: EqualityOperatorNotEqual,
				Left:     NewIntLiteral(1),
				Right:    NewIntLiteral(2),
			},
			wantErr: false,
		},
		{
			name: "invalid operator",
			expr: &LogicEquality{
				Operator: "<",
				Left:     NewIntLiteral(1),
				Right:    NewIntLiteral(2),
			},
			wantErr: true,
		},
		{
			name: "missing left",
			expr: &LogicEquality{
				Operator: EqualityOperatorEqual,
				Right:    NewIntLiteral(2),
			},
			wantErr: true,
		},
		{
			name: "missing right",
			expr: &LogicEquality{
				Operator: EqualityOperatorEqual,
				Left:     NewIntLiteral(1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.expr.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
