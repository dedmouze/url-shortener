package random

import (
	"math/rand"
	"time"
)

const alphabet = "abcdefghjklmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
const alphabetLen = len(alphabet)

func NewRandomString(length int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = alphabet[rnd.Intn(alphabetLen)]
	}
	return string(b)
}
