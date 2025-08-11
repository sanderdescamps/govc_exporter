package perfmetrics

import (
	"golang.org/x/exp/constraints"
)

func Avg[T constraints.Integer | constraints.Float](slice []T) float64 {
	return float64(Sum(slice)) / float64(len(slice))
}

func Sum[T constraints.Integer | constraints.Float](slice []T) T {
	var sum T
	for _, v := range slice {
		sum += v
	}
	return sum
}

func AllTrue(slice []bool) bool {
	for _, b := range slice {
		if !b {
			return false
		}
	}
	return true
}
