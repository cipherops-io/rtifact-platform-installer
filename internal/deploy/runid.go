package deploy

import (
	"crypto/rand"
	"encoding/hex"
)

func newRunID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "0"
	}
	return hex.EncodeToString(b[:])
}