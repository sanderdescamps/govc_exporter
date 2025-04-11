package helper

import "slices"

func Dedup[T comparable](slice []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range slice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
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
