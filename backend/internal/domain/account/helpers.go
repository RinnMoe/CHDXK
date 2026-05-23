package account

import (
	"crypto/rand"
)

func numericCode(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	code := make([]byte, length)
	for i, b := range buf {
		code[i] = byte('0' + int(b)%10)
	}
	return string(code), nil
}
