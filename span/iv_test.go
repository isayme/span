package span

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenIv(t *testing.T) {
	require := require.New(t)

	buf := make([]byte, 16)
	genIV(258, buf)
	require.Equal(byte(2), buf[15])
	require.Equal(byte(1), buf[14])
}
