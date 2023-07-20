package span

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAesEcb(t *testing.T) {
	require := require.New(t)

	key := []byte("1234567890abcdef")
	plaintext := "abcdefghijklmnop"
	ciphertextInHex := "2ee0f95a8451707ab5b6e1166501cb1f"

	{
		// encrypt
		ciphertext, err := AesEcbEncrypt(key, []byte(plaintext))
		require.Nil(err)
		require.Equal(ciphertextInHex, hex.EncodeToString(ciphertext))
	}

	{
		// decrypt
		text, err := AesEcbDecrypt(key, hexDecodeString(ciphertextInHex))
		require.Nil(err)
		require.Equal(plaintext, string(text))
	}
}
