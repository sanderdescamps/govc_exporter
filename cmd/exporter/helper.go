package main

func mergeLists[T any](l ...[]T) []T {
	result := []T{}
	for _, i := range l {
		if len(i) != 0 {
			result = append(result, i...)
		}
	}
	return result
}

func mergePtrLists[T any](l ...*[]T) []T {
	result := []T{}
	for _, i := range l {
		if i != nil && len(*i) != 0 {
			result = append(result, *i...)
		}
	}
	return result
}

func dedup[T comparable](slice []T) []T {
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
