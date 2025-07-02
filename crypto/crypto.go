package crypto

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"hash"
)

// Crashes program irrecoverable in case if secure random bytes are not available
func RandomBytes(n uint32) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)

	return b
}

func RandomSha1(nbytes uint32) string {
	return RandomHash(nbytes, sha1.New())
}

func RandomSha256(nbytes uint32) string {
	return RandomHash(nbytes, sha256.New())
}

func RandomHash(nbytes uint32, h hash.Hash) string {
	b := RandomBytes(nbytes)
	h.Reset()

	// Never returns an error according to official docs
	_, _ = h.Write(b)

	return hex.EncodeToString(h.Sum(nil))
}
