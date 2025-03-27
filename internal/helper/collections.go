package helper

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
