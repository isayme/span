package span

import (
	"crypto/aes"
	"crypto/cipher"
)

func AesCtrEncrypt(key, iv []byte, plaintext []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	b := cipher.NewCTR(c, iv)
	dst := make([]byte, len(plaintext))
	b.XORKeyStream(dst, plaintext)

	return dst, nil
}

func AesCtrDecrypt(key, iv []byte, ciphertext []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	b := cipher.NewCTR(c, iv)
	dst := make([]byte, len(ciphertext))
	b.XORKeyStream(dst, ciphertext)

	return dst, nil
}
