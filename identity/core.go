package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/yandzee/go-svc/crypto"
)

type Utils struct {
	SaltBytes uint32
}

func (dc *Utils) GenerateSalt() string {
	nbytes := dc.SaltBytes
	if nbytes == 0 {
		nbytes = 32
	}

	return crypto.RandomSha256(nbytes)
}

func (dc *Utils) Salt(salt string, target string) string {
	h := sha256.New()
	_, _ = fmt.Fprintf(h, "%s.%s", salt, target)

	return hex.EncodeToString(h.Sum(nil))

}
