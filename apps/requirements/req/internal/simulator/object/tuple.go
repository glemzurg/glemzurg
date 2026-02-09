package object

import (
	"bytes"
	"fmt"
	"strings"
)

// Tuple is an ordered collection of elements (sequence).
type Tuple struct {
	elements []Object
}

// NewTuple creates a new empty Tuple.
func NewTuple() *Tuple {
	return &Tuple{
		elements: make([]Object, 0),
	}
}

// NewTupleFromElements creates a Tuple from a slice of elements.
func NewTupleFromElements(elements []Object) *Tuple {
	t := NewTuple()
	for _, elem := range elements {
		t.elements = append(t.elements, elem.Clone())
	}
	return t
}

func (t *Tuple) Type() ObjectType { return TypeTuple }

func (t *Tuple) Inspect() string {
	var out bytes.Buffer
	elements := make([]string, len(t.elements))
	for i, elem := range t.elements {
		elements[i] = elem.Inspect()
	}
	out.WriteString("<<")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString(">>")
	return out.String()
}

func (t *Tuple) SetValue(source Object) error {
	src, ok := source.(*Tuple)
	if !ok {
		return fmt.Errorf("cannot assign %T to Tuple", source)
	}
	t.elements = make([]Object, len(src.elements))
	for i, elem := range src.elements {
		t.elements[i] = elem.Clone()
	}
	return nil
}

func (t *Tuple) Clone() Object {
	clone := NewTuple()
	clone.elements = make([]Object, len(t.elements))
	for i, elem := range t.elements {
		clone.elements[i] = elem.Clone()
	}
	return clone
}

// Len returns the number of elements in the tuple.
func (t *Tuple) Len() int {
	return len(t.elements)
}

// At returns the element at the given index (1-indexed, TLA+ style).
func (t *Tuple) At(index int) Object {
	if index < 1 || index > len(t.elements) {
		return nil
	}
	return t.elements[index-1]
}

// Set sets the element at the given index (1-indexed).
func (t *Tuple) Set(index int, value Object) error {
	if index < 1 || index > len(t.elements) {
		return fmt.Errorf("index %d out of bounds [1, %d]", index, len(t.elements))
	}
	t.elements[index-1] = value.Clone()
	return nil
}

// Elements returns all elements as a slice.
func (t *Tuple) Elements() []Object {
	result := make([]Object, len(t.elements))
	copy(result, t.elements)
	return result
}

// Head returns the first element, or nil if empty.
func (t *Tuple) Head() Object {
	if len(t.elements) == 0 {
		return nil
	}
	return t.elements[0]
}

// Tail returns a new tuple with all elements except the first.
func (t *Tuple) Tail() *Tuple {
	if len(t.elements) <= 1 {
		return NewTuple()
	}
	result := NewTuple()
	result.elements = make([]Object, len(t.elements)-1)
	for i := 1; i < len(t.elements); i++ {
		result.elements[i-1] = t.elements[i].Clone()
	}
	return result
}

// Append returns a new tuple with the element added at the end.
func (t *Tuple) Append(elem Object) *Tuple {
	result := t.Clone().(*Tuple)
	result.elements = append(result.elements, elem.Clone())
	return result
}

// Prepend returns a new tuple with the element added at the beginning.
func (t *Tuple) Prepend(elem Object) *Tuple {
	result := NewTuple()
	result.elements = make([]Object, len(t.elements)+1)
	result.elements[0] = elem.Clone()
	for i, e := range t.elements {
		result.elements[i+1] = e.Clone()
	}
	return result
}

// Concat returns a new tuple with all elements from both tuples.
func (t *Tuple) Concat(other *Tuple) *Tuple {
	result := NewTuple()
	result.elements = make([]Object, len(t.elements)+len(other.elements))
	for i, e := range t.elements {
		result.elements[i] = e.Clone()
	}
	for i, e := range other.elements {
		result.elements[len(t.elements)+i] = e.Clone()
	}
	return result
}

// SubSeq returns a subsequence from start to end (1-indexed, inclusive).
func (t *Tuple) SubSeq(start, end int) *Tuple {
	if start < 1 {
		start = 1
	}
	if end > len(t.elements) {
		end = len(t.elements)
	}
	if start > end {
		return NewTuple()
	}
	result := NewTuple()
	result.elements = make([]Object, end-start+1)
	for i := start; i <= end; i++ {
		result.elements[i-start] = t.elements[i-1].Clone()
	}
	return result
}

// Reverse returns a new tuple with elements in reverse order.
func (t *Tuple) Reverse() *Tuple {
	result := NewTuple()
	result.elements = make([]Object, len(t.elements))
	for i, elem := range t.elements {
		result.elements[len(t.elements)-1-i] = elem.Clone()
	}
	return result
}

// Contains checks if an element is in the tuple.
func (t *Tuple) Contains(elem Object) bool {
	key := hashValue(elem)
	for _, e := range t.elements {
		if hashValue(e) == key {
			return true
		}
	}
	return false
}

// Push adds an element to the end (stack-style: push to top).
func (t *Tuple) Push(elem Object) {
	t.elements = append(t.elements, elem.Clone())
}

// Pop removes and returns the last element (stack-style: pop from top).
func (t *Tuple) Pop() Object {
	if len(t.elements) == 0 {
		return nil
	}
	last := t.elements[len(t.elements)-1]
	t.elements = t.elements[:len(t.elements)-1]
	return last
}

// Enqueue adds an element to the end (queue-style: add to back).
func (t *Tuple) Enqueue(elem Object) {
	t.elements = append(t.elements, elem.Clone())
}

// Dequeue removes and returns the first element (queue-style: remove from front).
func (t *Tuple) Dequeue() Object {
	if len(t.elements) == 0 {
		return nil
	}
	first := t.elements[0]
	t.elements = t.elements[1:]
	return first
}

// Equals checks if two tuples have the same elements in the same order.
func (t *Tuple) Equals(other *Tuple) bool {
	if len(t.elements) != len(other.elements) {
		return false
	}
	for i := range t.elements {
		if hashValue(t.elements[i]) != hashValue(other.elements[i]) {
			return false
		}
	}
	return true
}
