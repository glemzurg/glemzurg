package object

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// BagEntry holds an element and its count in a bag.
type BagEntry struct {
	Value Object
	Count int
}

// Bag is an unordered collection that allows duplicate elements (multiset).
type Bag struct {
	elements map[string]BagEntry // keyed by hashValue(element)
}

// NewBag creates a new empty Bag.
func NewBag() *Bag {
	return &Bag{
		elements: make(map[string]BagEntry),
	}
}

func (b *Bag) Type() ObjectType { return TypeBag }

func (b *Bag) Inspect() string {
	var out bytes.Buffer
	elements := make([]string, 0, b.Size())
	for _, entry := range b.elements {
		for i := 0; i < entry.Count; i++ {
			elements = append(elements, entry.Value.Inspect())
		}
	}
	sort.Strings(elements)
	out.WriteString("(")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString(")")
	return out.String()
}

func (b *Bag) SetValue(source Object) error {
	src, ok := source.(*Bag)
	if !ok {
		return fmt.Errorf("cannot assign %T to Bag", source)
	}
	b.elements = make(map[string]BagEntry, len(src.elements))
	for k, entry := range src.elements {
		b.elements[k] = BagEntry{Value: entry.Value.Clone(), Count: entry.Count}
	}
	return nil
}

func (b *Bag) Clone() Object {
	clone := NewBag()
	for k, entry := range b.elements {
		clone.elements[k] = BagEntry{Value: entry.Value.Clone(), Count: entry.Count}
	}
	return clone
}

// Size returns the total count of all elements (including duplicates).
func (b *Bag) Size() int {
	total := 0
	for _, entry := range b.elements {
		total += entry.Count
	}
	return total
}

// UniqueCount returns the number of distinct elements.
func (b *Bag) UniqueCount() int {
	return len(b.elements)
}

// Add adds count copies of an element to the bag.
func (b *Bag) Add(elem Object, count int) {
	if count <= 0 {
		return
	}
	key := hashValue(elem)
	if entry, exists := b.elements[key]; exists {
		b.elements[key] = BagEntry{Value: entry.Value, Count: entry.Count + count}
	} else {
		b.elements[key] = BagEntry{Value: elem.Clone(), Count: count}
	}
}

// Remove removes count copies of an element from the bag.
func (b *Bag) Remove(elem Object, count int) {
	if count <= 0 {
		return
	}
	key := hashValue(elem)
	if entry, exists := b.elements[key]; exists {
		newCount := entry.Count - count
		if newCount <= 0 {
			delete(b.elements, key)
		} else {
			b.elements[key] = BagEntry{Value: entry.Value, Count: newCount}
		}
	}
}

// CopiesIn returns the number of copies of an element in the bag.
func (b *Bag) CopiesIn(elem Object) int {
	if entry, exists := b.elements[hashValue(elem)]; exists {
		return entry.Count
	}
	return 0
}

// Contains checks if an element exists in the bag.
func (b *Bag) Contains(elem Object) bool {
	return b.CopiesIn(elem) > 0
}

// Elements returns all unique elements in the bag.
func (b *Bag) Elements() []Object {
	result := make([]Object, 0, len(b.elements))
	for _, entry := range b.elements {
		result = append(result, entry.Value)
	}
	return result
}

// Union returns a new bag where each element's count is the max of the two bags.
func (b *Bag) Union(other *Bag) *Bag {
	result := NewBag()
	for k, entry := range b.elements {
		result.elements[k] = BagEntry{Value: entry.Value.Clone(), Count: entry.Count}
	}
	for k, entry := range other.elements {
		if existing, exists := result.elements[k]; exists {
			if entry.Count > existing.Count {
				result.elements[k] = BagEntry{Value: entry.Value.Clone(), Count: entry.Count}
			}
		} else {
			result.elements[k] = BagEntry{Value: entry.Value.Clone(), Count: entry.Count}
		}
	}
	return result
}

// Sum returns a new bag where each element's count is the sum of the two bags.
func (b *Bag) Sum(other *Bag) *Bag {
	result := NewBag()
	for k, entry := range b.elements {
		result.elements[k] = BagEntry{Value: entry.Value.Clone(), Count: entry.Count}
	}
	for k, entry := range other.elements {
		if existing, exists := result.elements[k]; exists {
			result.elements[k] = BagEntry{Value: existing.Value, Count: existing.Count + entry.Count}
		} else {
			result.elements[k] = BagEntry{Value: entry.Value.Clone(), Count: entry.Count}
		}
	}
	return result
}

// Difference returns a new bag with element counts reduced by the other bag.
func (b *Bag) Difference(other *Bag) *Bag {
	result := NewBag()
	for k, entry := range b.elements {
		otherCount := 0
		if otherEntry, exists := other.elements[k]; exists {
			otherCount = otherEntry.Count
		}
		newCount := entry.Count - otherCount
		if newCount > 0 {
			result.elements[k] = BagEntry{Value: entry.Value.Clone(), Count: newCount}
		}
	}
	return result
}

// IsSubBagOf checks if this bag is a sub-bag of other (⊑).
// Returns true if for every element, this bag's count <= other's count.
func (b *Bag) IsSubBagOf(other *Bag) bool {
	for k, entry := range b.elements {
		otherCount := 0
		if otherEntry, exists := other.elements[k]; exists {
			otherCount = otherEntry.Count
		}
		if entry.Count > otherCount {
			return false
		}
	}
	return true
}

// IsProperSubBagOf checks if this bag is a proper sub-bag of other (⊏).
// Returns true if this is a sub-bag of other AND they are not equal.
func (b *Bag) IsProperSubBagOf(other *Bag) bool {
	return b.IsSubBagOf(other) && !b.Equals(other)
}

// IsSuperBagOf checks if this bag is a super-bag of other (⊒).
// Returns true if other is a sub-bag of this.
func (b *Bag) IsSuperBagOf(other *Bag) bool {
	return other.IsSubBagOf(b)
}

// IsProperSuperBagOf checks if this bag is a proper super-bag of other (⊐).
// Returns true if this is a super-bag of other AND they are not equal.
func (b *Bag) IsProperSuperBagOf(other *Bag) bool {
	return b.IsSuperBagOf(other) && !b.Equals(other)
}

// Equals checks if two bags have the same elements with the same counts.
func (b *Bag) Equals(other *Bag) bool {
	if len(b.elements) != len(other.elements) {
		return false
	}
	for k, entry := range b.elements {
		otherEntry, exists := other.elements[k]
		if !exists || entry.Count != otherEntry.Count {
			return false
		}
	}
	return true
}

