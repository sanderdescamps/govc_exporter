package helper

import (
	"maps"
	"slices"
)

func Dedup[T comparable](slice []T) []T {
	allKeys := make(map[T]bool)
	for _, item := range slice {
		allKeys[item] = true
	}
	return slices.Collect(maps.Keys(allKeys))
}

func DedupFunc[T any](slice []T, cmp func(i1 T, i2 T) bool) []T {
	clean := []T{}
	for _, item := range slice {
		if slices.ContainsFunc(clean, func(other T) bool {
			return cmp(other, item)
		}) {
			clean = append(clean, item)
		}
	}
	return clean
}

func Flatten[T any](l ...[]T) []T {
	result := []T{}
	for _, i := range l {
		if len(i) != 0 {
			result = append(result, i...)
		}
	}
	return result
}

func Contains[T comparable](elems []T, v T) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

// Subtract returns the elements in a that are not in b
// It is not commutative, i.e. Subtract(a, b) != Subtract(b, a)
func Subtract[T comparable](a []T, b []T) []T {
	var diff []T
	for _, a1 := range a {
		if !slices.Contains(b, a1) {
			diff = append(diff, a1)
		}
	}
	return diff
}

// Union returns the union of all slices in a
// It is commutative, i.e. Union(a, b) == Union(b, a)
// It is not guaranteed that the order of the elements in the result is the same as in the input slices
// The result is deduplicated
func Union[T comparable](a ...[]T) []T {
	elements := make([]T, 0)
	for _, slice := range a {
		elements = append(elements, slice...)
	}
	return Dedup(elements)
}

func Intersect[T comparable](a []T, b []T) []T {
	var intersect []T
	for _, a1 := range a {
		if slices.Contains(b, a1) {
			intersect = append(intersect, a1)
		}
	}
	return intersect
}
