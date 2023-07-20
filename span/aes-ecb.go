package span

import (
	"crypto/aes"
)

func AesEcbEncrypt(key []byte, plaintext []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	dst := make([]byte, len(plaintext))
	c.Encrypt(dst, plaintext)

	return dst, nil
}

func AesEcbDecrypt(key []byte, ciphertext []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	dst := make([]byte, len(ciphertext))
	c.Decrypt(dst, ciphertext)

	return dst, nil
}
