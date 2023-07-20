package span

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func hexDecodeString(s string) []byte {
	r, _ := hex.DecodeString(s)
	return r
}

func TestBase64(t *testing.T) {
	require := require.New(t)

	// no padding
	cases := [][]byte{
		[]byte("1"), []byte("MQ"), // no padding
		hexDecodeString("f8"), []byte("-A"), // URL Safe: + => -
		hexDecodeString("fc"), []byte("_A"), // URL Safe: / => _
		[]byte("123"), []byte("MTIz"), // nomarl
	}

	for i := 0; i < len(cases); i = i + 2 {
		src := cases[i]
		expect := cases[i+1]

		require.Equal(string(expect), Base64EncodeToString(src))

		decodeResult, err := Base64DecodeString(string(expect))
		require.Nil(err)
		require.Equal(hex.EncodeToString(src), hex.EncodeToString(decodeResult))
	}
}
