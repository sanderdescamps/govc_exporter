package scraper

import (
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/exp/constraints"
)

func tcpConnectionCheck(endpoint string) (bool, error) {
	conn, err := net.DialTimeout("tcp", endpoint, 3*time.Second)
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()
	if err != nil {
		return false, fmt.Errorf("failed to connect to TCP %s", endpoint)
	}
	return true, nil
}

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

func cleanString(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}
