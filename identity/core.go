package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/yandzee/go-svc/crypto"
)

type DefaultCore struct{}

func (dc *DefaultCore) IsValid(pc *PlainCredentials) (string, bool) {
	return pc.IsValid()
}

func (dc *DefaultCore) GenerateSalt() string {
	return crypto.RandomSha256(32)
}

func (dc *DefaultCore) Salt(salt string, target string) string {
	h := sha256.New()
	_, _ = fmt.Fprintf(h, "%s.%s", salt, target)

	return hex.EncodeToString(h.Sum(nil))

}
