package scraper

import (
	"fmt"
	"net"
	"time"
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
