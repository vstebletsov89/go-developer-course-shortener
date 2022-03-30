package rand

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateRandom is used in two places: middleware and handlers
func GenerateRandom(size int) []byte {
	var Letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")
	b := make([]byte, size)
	for i := range b {
		b[i] = Letters[rand.Intn(len(Letters))]
	}
	return b
}
