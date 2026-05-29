package utils

import (
	"encoding/binary"
	"span/internal/constants"
)

func GenIV(pos int64, iv []byte) {
	for i := 0; i < len(iv); i++ {
		iv[i] = 0
	}

	n := pos / int64(constants.AesBlockSize) * int64(constants.AesBlockSize)
	binary.BigEndian.PutUint64(iv[8:], uint64(n))
}
