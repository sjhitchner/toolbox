package sort

import (
	"sort"

	"golang.org/x/exp/constraints"
)

func MapSort[T constraints.Ordered, V any](m map[T]V, order SortOrder) <-chan T {
	var keys []T
	for key := range m {
		keys = append(keys, key)
	}

	SortSlice(keys, order)

	out := make(chan T)
	go func() {
		defer close(out)

		for _, key := range keys {
			out <- key
		}
	}()

	return out
}

type SortOrder bool

const (
	Descending SortOrder = true
	Ascending  SortOrder = false
)

func SortSlice[T constraints.Ordered](s []T, order SortOrder) {
	sort.Slice(s, func(i, j int) bool {
		if order {
			return s[i] > s[j]
		} else {
			return s[i] < s[j]
		}
	})
}

type Number interface {
	int | int64 | float32 | float64
}

type Pair[T constraints.Ordered, V Number] struct {
	key   T
	value V
}

func SortMap[T constraints.Ordered, V Number](m map[T]V) []T {
	// Convert the map to a slice of key-value pairs

	var pairs = make([]Pair[T, V], 0, len(m))
	for k, v := range m {
		pairs = append(pairs, Pair[T, V]{k, v})
	}

	// Sort the slice based on values
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value > pairs[j].value
	})

	// Rebuild the map from the sorted slice
	sorted := make([]T, 0, len(pairs))
	for _, p := range pairs {
		sorted = append(sorted, p.key)
	}

	return sorted
}
