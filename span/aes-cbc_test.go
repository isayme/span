package span

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAesCbc(t *testing.T) {
	require := require.New(t)

	key := []byte("1234567890abcdef")
	iv := []byte("1234567890abcdef")

	plaintext := "abcdefghijklmnop"
	ciphertextInHex := "f248daeace1898570bab65ee01db4f5c"

	{
		// encrypt
		ciphertext, err := AesCbcEncrypt(key, iv, []byte(plaintext))
		require.Nil(err)
		require.Equal(ciphertextInHex, hex.EncodeToString(ciphertext))
	}

	{
		// decrypt
		text, err := AesCbcDecrypt(key, iv, hexDecodeString(ciphertextInHex))
		require.Nil(err)
		require.Equal(plaintext, string(text))
	}
}
