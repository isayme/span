package span

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAesCtr(t *testing.T) {
	require := require.New(t)

	key := []byte("1234567890abcdef")
	iv := []byte("1234567890abcdef")

	{
		plaintext := "abcdefghijklmnop"
		ciphertextInHex := "f4d271d4d9efe9345e848609e50d3079"

		{
			// encrypt
			ciphertext, err := AesCtrEncrypt(key, iv, []byte(plaintext))
			require.Nil(err)
			require.Equal(ciphertextInHex, hex.EncodeToString(ciphertext))
		}

		{
			// decrypt
			text, err := AesCtrDecrypt(key, iv, hexDecodeString(ciphertextInHex))
			require.Nil(err)
			require.Equal(plaintext, string(text))
		}
	}

	{
		plaintext := "abc"
		ciphertextInHex := "f4d271"

		{
			// encrypt
			ciphertext, err := AesCtrEncrypt(key, iv, []byte(plaintext))
			require.Nil(err)
			require.Equal(ciphertextInHex, hex.EncodeToString(ciphertext))
		}

		{
			// decrypt
			text, err := AesCtrDecrypt(key, iv, hexDecodeString(ciphertextInHex))
			require.Nil(err)
			require.Equal(plaintext, string(text))
		}
	}

}
