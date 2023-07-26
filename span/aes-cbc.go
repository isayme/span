package span

import (
	"crypto/aes"
	"crypto/cipher"
)

func MustAesCbcEncrypt(key, iv []byte, plaintext []byte) []byte {
	result, err := AesCbcEncrypt(key, iv, plaintext)
	if err != nil {
		panic(err)
	}

	return result
}

func AesCbcEncrypt(key, iv []byte, plaintext []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	b := cipher.NewCBCEncrypter(c, iv)
	dst := make([]byte, len(plaintext))
	b.CryptBlocks(dst, plaintext)

	return dst, nil
}

func MustAesCbcDecrypt(key, iv []byte, ciphertext []byte) []byte {
	result, err := AesCbcDecrypt(key, iv, ciphertext)
	if err != nil {
		panic(err)
	}

	return result
}

func AesCbcDecrypt(key, iv []byte, ciphertext []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	b := cipher.NewCBCDecrypter(c, iv)
	dst := make([]byte, len(ciphertext))
	b.CryptBlocks(dst, ciphertext)

	return dst, nil
}
