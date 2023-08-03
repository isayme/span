package span

import (
	"bytes"
	"crypto/aes"
)

var aesBlockSize = aes.BlockSize

func Pkcs5Padding(b []byte) []byte {
	padSize := aesBlockSize - len(b)%aesBlockSize
	padding := bytes.Repeat([]byte{byte(padSize)}, padSize)
	return append(b, padding...)
}

func Pkcs5UnPadding(b []byte) []byte {
	padSize := int(b[len(b)-1])
	return b[:len(b)-padSize]
}
