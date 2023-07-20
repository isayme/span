package span

import "crypto/sha256"

func Sha256(in []byte) []byte {
	h := sha256.New()
	h.Write(in)
	return h.Sum(nil)
}
