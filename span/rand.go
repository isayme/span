package span

import "crypto/rand"

func randomBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func mustRandomBytes(n int) []byte {
	b, err := randomBytes(n)
	if err != nil {
		panic("random bytes fail: " + err.Error())
	}
	return b
}
