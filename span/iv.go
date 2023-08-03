package span

import "encoding/binary"

func genIV(pos int64, iv []byte) {
	for i := 0; i < len(iv); i++ {
		iv[i] = 0
	}

	n := pos / int64(aesBlockSize) * int64(aesBlockSize)
	binary.BigEndian.PutUint64(iv[8:], uint64(n))
}
