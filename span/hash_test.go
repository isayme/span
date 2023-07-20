package span

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSha256(t *testing.T) {
	result := Sha256([]byte("123"))

	assert.Equal(t, "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3", hex.EncodeToString(result))
}
