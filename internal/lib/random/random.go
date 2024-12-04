package random

import (
	"math/rand"
)

var table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

const codeLen = 6

func Code() string {
	result := make([]byte, codeLen)
	for i := 0; i < codeLen; i++ {
		result[i] = table[rand.Intn(len(table))]
	}

	return string(result)
}
