package span

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAesGcm(t *testing.T) {
	require := require.New(t)

	key := []byte("1234567890abcdef")
	nonce := []byte("1234567890abcdef") // 12byte

	plaintext := "abcdefghijklmnop"
	ciphertextInHex := "814b89b77f7c66ff613907f869f4ae0ff3a9e9ede746751e6f67795392e5c028"

	{
		// encrypt
		ciphertext, err := AesGcmEncrypt(key, nonce, []byte(plaintext))
		require.Nil(err)
		require.Equal(ciphertextInHex, hex.EncodeToString(ciphertext))
	}

	{
		// decrypt
		text, err := AesGcmDecrypt(key, nonce, hexDecodeString(ciphertextInHex))
		require.Nil(err)
		require.Equal(plaintext, string(text))
	}
}
