package object

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// Set is an unordered collection of unique elements.
type Set struct {
	elements map[string]Object // keyed by hashValue(element)
}

// NewSet creates a new empty Set.
func NewSet() *Set {
	return &Set{
		elements: make(map[string]Object),
	}
}

// NewSetFromElements creates a Set from a slice of elements.
func NewSetFromElements(elements []Object) *Set {
	s := NewSet()
	for _, elem := range elements {
		s.Add(elem)
	}
	return s
}

func (s *Set) Type() ObjectType { return TypeSet }

func (s *Set) Inspect() string {
	var out bytes.Buffer
	elements := make([]string, 0, len(s.elements))
	for _, elem := range s.elements {
		elements = append(elements, elem.Inspect())
	}
	sort.Strings(elements)
	out.WriteString("{")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("}")
	return out.String()
}

func (s *Set) SetValue(source Object) error {
	src, ok := source.(*Set)
	if !ok {
		return fmt.Errorf("cannot assign %T to Set", source)
	}
	s.elements = make(map[string]Object, len(src.elements))
	for k, v := range src.elements {
		s.elements[k] = v.Clone()
	}
	return nil
}

func (s *Set) Clone() Object {
	clone := NewSet()
	for k, v := range s.elements {
		clone.elements[k] = v.Clone()
	}
	return clone
}

// Size returns the number of elements in the set.
func (s *Set) Size() int {
	return len(s.elements)
}

// Add adds an element to the set.
func (s *Set) Add(elem Object) {
	s.elements[hashValue(elem)] = elem.Clone()
}

// Remove removes an element from the set.
func (s *Set) Remove(elem Object) {
	delete(s.elements, hashValue(elem))
}

// Contains checks if an element is in the set.
func (s *Set) Contains(elem Object) bool {
	_, exists := s.elements[hashValue(elem)]
	return exists
}

// Elements returns all elements as a slice.
func (s *Set) Elements() []Object {
	result := make([]Object, 0, len(s.elements))
	for _, elem := range s.elements {
		result = append(result, elem)
	}
	return result
}

// Union returns a new set containing all elements from both sets.
func (s *Set) Union(other *Set) *Set {
	result := NewSet()
	for k, v := range s.elements {
		result.elements[k] = v.Clone()
	}
	for k, v := range other.elements {
		result.elements[k] = v.Clone()
	}
	return result
}

// Intersection returns a new set containing only elements in both sets.
func (s *Set) Intersection(other *Set) *Set {
	result := NewSet()
	for k, v := range s.elements {
		if _, exists := other.elements[k]; exists {
			result.elements[k] = v.Clone()
		}
	}
	return result
}

// Difference returns a new set containing elements in s but not in other.
func (s *Set) Difference(other *Set) *Set {
	result := NewSet()
	for k, v := range s.elements {
		if _, exists := other.elements[k]; !exists {
			result.elements[k] = v.Clone()
		}
	}
	return result
}

// IsSubsetOf checks if this set is a subset of other.
func (s *Set) IsSubsetOf(other *Set) bool {
	for k := range s.elements {
		if _, exists := other.elements[k]; !exists {
			return false
		}
	}
	return true
}

// Equals checks if two sets contain the same elements.
func (s *Set) Equals(other *Set) bool {
	if len(s.elements) != len(other.elements) {
		return false
	}
	for k := range s.elements {
		if _, exists := other.elements[k]; !exists {
			return false
		}
	}
	return true
}

