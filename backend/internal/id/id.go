package id

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

func New() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	raw := hex.EncodeToString(b[:])
	return strings.Join([]string{
		raw[0:8],
		raw[8:12],
		raw[12:16],
		raw[16:20],
		raw[20:32],
	}, "-")
}
