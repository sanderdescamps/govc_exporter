package collector

func b2f(val bool) float64 {
	if val {
		return 1.0
	}
	return 0.0
}
