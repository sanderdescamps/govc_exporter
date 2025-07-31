package objects

import "strings"

// Return string color as a float64 value
//
//	0 => Gray
//	1 => Red
//	2 => Yellow
//	3 => Green
func ColorToFloat64(color string) float64 {
	if color == "" || strings.EqualFold(color, "Gray") {
		return 0
	} else if strings.EqualFold(color, "Red") {
		return 1.0
	} else if strings.EqualFold(color, "Yellow") {
		return 2.0
	} else if strings.EqualFold(color, "Green") {
		return 3.0
	}
	return 0
}
