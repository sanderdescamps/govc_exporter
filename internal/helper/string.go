package helper

import (
	"crypto/sha256"
	"fmt"
)

func CommonPrefixLen(x, y string) int {
	n := 0
	for n < len(x) && n < len(y) && x[n] == y[n] {
		n++
	}
	return n
}

func CommonPrefix(x, y string) string {
	return x[0:CommonPrefixLen(x, y)]
}

func HashStrings(strs ...string) string {
	h := sha256.New()
	for _, s := range strs {
		h.Write([]byte(s))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
