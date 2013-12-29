package main

import (
	"math/rand"
)

// Random alphanum id generator
func GetRandId(length int) string {
	chars := []byte("1234567890abcdefghijklmnopqrstuvwxyz")
	charsLen := len(chars)
	randId := make([]byte, length)

	for i := 0; i < length; i++ {
		randId[i] = chars[rand.Intn(charsLen)]
	}

	return string(randId)
}
