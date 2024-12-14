package random

import (
	"math/rand/v2"
)

var table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

const codeLen = 6

func Code() string {
	result := make([]byte, codeLen)
	for i := 0; i < codeLen; i++ {
		result[i] = table[rand.IntN(len(table))]
	}

	return string(result)
}
