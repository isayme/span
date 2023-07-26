package span

import (
	"bytes"
	"crypto/aes"
)

var aseBlockSize = aes.BlockSize

func Pkcs5Padding(b []byte) []byte {
	padSize := aseBlockSize - len(b)%aseBlockSize
	padding := bytes.Repeat([]byte{byte(padSize)}, padSize)
	return append(b, padding...)
}

func Pkcs5UnPadding(b []byte) []byte {
	padSize := int(b[len(b)-1])
	return b[:len(b)-padSize]
}
