package common

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashRefreshToken(raw string, pepper []byte) string {
	h := sha256.New()
	h.Write(pepper)
	h.Write([]byte(raw))
	return hex.EncodeToString(h.Sum(nil))
}
