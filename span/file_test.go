package span

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetIv(t *testing.T) {
	require := require.New(t)

	buf := make([]byte, 16)
	getIv(258, buf)
	require.Equal(byte(2), buf[15])
	require.Equal(byte(1), buf[14])
}
