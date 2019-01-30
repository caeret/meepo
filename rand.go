package meepo

import (
	"math/rand"
	"time"
)

const seeds = `abcdefghigklmnopqrstuvwxyzabcdefghigklmnopqrstuvwxyz0123456789`

func init() {
	rand.Seed(time.Now().Unix())
}

func randStr(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = seeds[rand.Intn(62)]
	}
	return string(b)
}
