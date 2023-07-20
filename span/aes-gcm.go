package span

import (
	"crypto/aes"
	"crypto/cipher"
)

var aseGcmNonceSize = 16

func AesGcmEncrypt(key, nonce []byte, plaintext []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	b, err := cipher.NewGCMWithNonceSize(c, aseGcmNonceSize)
	if err != nil {
		return nil, err
	}

	dst := b.Seal(nil, nonce, plaintext, nil)

	return dst, nil
}

func AesGcmDecrypt(key, nonce []byte, ciphertext []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	b, err := cipher.NewGCMWithNonceSize(c, aseGcmNonceSize)
	if err != nil {
		return nil, err
	}

	dst, err := b.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return dst, nil
}
