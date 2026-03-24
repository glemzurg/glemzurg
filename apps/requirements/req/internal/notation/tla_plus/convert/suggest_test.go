package convert

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SuggestTestSuite struct {
	suite.Suite
}

func TestSuggest(t *testing.T) {
	suite.Run(t, new(SuggestTestSuite))
}

func (s *SuggestTestSuite) TestLevenshteinDistance() {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"identical strings", "abc", "abc", 0},
		{"one insertion", "abc", "abcd", 1},
		{"one deletion", "abcd", "abc", 1},
		{"one substitution", "abc", "axc", 1},
		{"empty first", "", "abc", 3},
		{"empty second", "abc", "", 3},
		{"both empty", "", "", 0},
		{"completely different", "abc", "xyz", 3},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := levenshteinDistance(tt.a, tt.b)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *SuggestTestSuite) TestSuggestSimilar() {
	tests := []struct {
		name        string
		input       string
		candidates  []string
		maxDistance int
		expected    []string
	}{
		{
			name:        "exact match returns match",
			input:       "hello",
			candidates:  []string{"world", "hello", "help"},
			maxDistance: 2,
			expected:    []string{"hello", "help"},
		},
		{
			name:        "close match with 1 edit returns suggestions",
			input:       "helo",
			candidates:  []string{"hello", "help", "world"},
			maxDistance: 1,
			expected:    []string{"hello", "help"},
		},
		{
			name:        "no match when all candidates are too far",
			input:       "xyz",
			candidates:  []string{"abcdef", "ghijkl", "mnopqr"},
			maxDistance: 2,
			expected:    nil,
		},
		{
			name:        "multiple suggestions sorted by distance",
			input:       "cat",
			candidates:  []string{"car", "bat", "cap", "dog", "cats"},
			maxDistance: 2,
			expected:    []string{"bat", "cap", "car"},
		},
		{
			name:        "empty candidates returns nil",
			input:       "hello",
			candidates:  []string{},
			maxDistance: 2,
			expected:    nil,
		},
		{
			name:        "nil candidates returns nil",
			input:       "hello",
			candidates:  nil,
			maxDistance: 2,
			expected:    nil,
		},
		{
			name:        "case-sensitive matching",
			input:       "Hello",
			candidates:  []string{"hello", "HELLO", "Hello"},
			maxDistance: 0,
			expected:    []string{"Hello"},
		},
		{
			name:        "at most 3 results returned",
			input:       "ab",
			candidates:  []string{"aa", "ac", "ad", "ae", "af"},
			maxDistance: 1,
			expected:    []string{"aa", "ac", "ad"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := SuggestSimilar(tt.input, tt.candidates, tt.maxDistance)
			s.Equal(tt.expected, result)
		})
	}
}
