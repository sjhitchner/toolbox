package set

import (
	"bytes"
	"encoding/json"
	"fmt"
	// "reflect"
	// "strings"
)

type Set[T comparable] map[T]struct{}

// New create new set
func New[T comparable](s ...T) Set[T] {
	set := make(Set[T])
	set.Add(s...)
	return set
}

// Add to set
func (set Set[T]) Add(s ...T) {
	for _, v := range s {
		set[v] = struct{}{}
	}
}

// CheckAndAdd check if exists and add
func (set Set[T]) CheckAndAdd(i T) bool {
	_, found := set[i]
	set[i] = struct{}{}
	return !found //False if it existed already
}

// Merge Sets
func (set Set[T]) Merge(others ...Set[T]) {
	for _, other := range others {
		for s := range other {
			set[s] = struct{}{}
		}
	}
}

// Contains does value exist in set
func (set Set[T]) Contains(i ...T) bool {
	for _, val := range i {
		if _, ok := set[val]; !ok {
			return false
		}
	}
	return true
}

func (set Set[T]) IsSubset(other Set[T]) bool {
	for elem := range set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (set Set[T]) IsSuperset(other Set[T]) bool {
	return other.IsSubset(set)
}

func (set Set[T]) Union(other ...Set[T]) Set[T] {
	unionedSet := New[T]()
	for elem := range set {
		unionedSet.Add(elem)
	}

	for _, oset := range other {
		for elem := range oset {
			unionedSet.Add(elem)
		}
	}
	return unionedSet
}

func (set Set[T]) Intersect(other Set[T]) Set[T] {
	intersection := New[T]()

	// loop over smaller set
	if set.Cardinality() < other.Cardinality() {
		for elem := range set {
			if other.Contains(elem) {
				intersection.Add(elem)
			}
		}
		return intersection

	}

	for elem := range other {
		if set.Contains(elem) {
			intersection.Add(elem)
		}
	}

	return intersection
}

func (set Set[T]) Difference(other Set[T]) Set[T] {
	difference := New[T]()
	for elem := range set {
		if !other.Contains(elem) {
			difference.Add(elem)
		}
	}
	return difference
}

func (set Set[T]) SymmetricDifference(other Set[T]) Set[T] {
	aDiff := set.Difference(other)
	bDiff := other.Difference(set)
	return aDiff.Union(bDiff)
}

func (set Set[T]) HasOverlap(other Set[T]) bool {

	if set.Cardinality() < other.Cardinality() {
		for elem := range set {
			if other.Contains(elem) {
				return true
			}
		}
		return false
	}

	for elem := range other {
		if set.Contains(elem) {
			return true
		}
	}
	return false
}

func (set *Set[T]) Clear() {
	*set = New[T]()
}

func (set Set[T]) Remove(i T) {
	delete(set, i)
}

func (set Set[T]) Cardinality() int {
	return len(set)
}

func (set Set[T]) IsEmpty() bool {
	return len(set) == 0
}

func (set Set[T]) Iter() <-chan T {
	ch := make(chan T)
	go func() {
		for elem := range set {
			ch <- elem
		}
		close(ch)
	}()

	return ch
}

func (set Set[T]) Iterator() *Iterator[T] {
	iterator, ch, stopCh := newIterator[T]()

	go func() {
	L:
		for elem := range set {
			select {
			case <-stopCh:
				break L
			case ch <- elem:
			}
		}
		close(ch)
	}()

	return iterator
}

func (set Set[T]) Equals(other Set[T]) bool {
	if set.Cardinality() != other.Cardinality() {
		return false
	}

	for elem := range set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (set Set[T]) Clone() Set[T] {
	clonedSet := New[T]()
	for elem := range set {
		clonedSet.Add(elem)
	}
	return clonedSet
}

func (set Set[T]) String() string {

	buf := &bytes.Buffer{}
	fmt.Fprint(buf, "Set{")

	var index int
	for elem := range set {
		if index > 0 {
			buf.WriteString(", ")
		}

		fmt.Fprintf(buf, "%v", elem)
		index++
	}

	buf.WriteRune('}')
	return buf.String()
}

/*
func (set StringSet) PowerSet() StringSet {
	powSet := NewStringSet()
	nullset := NewStringSet()
	powSet.AddSet(nullset)

	for es := range set {
		u := NewStringSet()
		j := powSet.Iter()
		for er := range j {
			p := NewStringSet()

			for ek := range er {
				p.Add(ek)
			}
			p.Add(es)
			u.AddSet(p)
		}

		powSet = powSet.Union(u)
	}

	return powSet
}
*/

/*
func (set StringSet) CartesianProduct(other StringSet) StringSet {
	cartProduct := NewStringSet()

	for i := range set {
		for j := range other {
			elem := orderedPair{first: i, second: j}
			cartProduct.Add(elem)
		}
	}

	return cartProduct
}
*/

func (set Set[T]) ToSlice() []T {
	keys := make([]T, 0, set.Cardinality())
	for elem := range set {
		keys = append(keys, elem)
	}
	return keys
}

// MarshalJSON creates a JSON array from the set, it marshals all elements
func (set Set[T]) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteRune('[')

	var index int
	for elem := range set {
		if index > 0 {
			buf.WriteRune(',')
		}

		fmt.Fprintf(buf, `"%v"`, elem)
		index++
	}

	buf.WriteRune(']')
	return buf.Bytes(), nil
}

// UnmarshalJSON recreates a set from a JSON array, it only decodes
// primitive types. Numbers are decoded as json.Number.
func (set *Set[T]) UnmarshalJSON(b []byte) error {
	var list []T
	if err := json.Unmarshal(b, &list); err != nil {
		return err
	}

	s := New[T](list...)
	*set = s

	return nil
}

// Iterator defines an iterator over a Set, its C channel can be used to range over the Set's
// elements.
type Iterator[T comparable] struct {
	C    <-chan T
	stop chan struct{}
}

// Stop stops the Iterator, no further elements will be received on C, C will be closed.
func (i *Iterator[T]) Stop() {
	// Allows for Stop() to be called multiple times
	// (close() panics when called on already closed channel)
	defer func() {
		recover()
	}()

	close(i.stop)

	// Exhaust any remaining elements.
	for _ = range i.C {
	}
}

// newIterator returns a new Iterator instance together with its item and stop channels.
func newIterator[T comparable]() (*Iterator[T], chan<- T, <-chan struct{}) {
	itemChan := make(chan T)
	stopChan := make(chan struct{})
	return &Iterator[T]{
		C:    itemChan,
		stop: stopChan,
	}, itemChan, stopChan
}

type orderedPair struct {
	first  string
	second string
}

func (pair *orderedPair) Equal(other orderedPair) bool {
	if pair.first == other.first &&
		pair.second == other.second {
		return true
	}

	return false
}

func (pair orderedPair) String() string {
	return fmt.Sprintf("(%v, %v)", pair.first, pair.second)
}
