package utils

import (
	"bytes"
	"span/internal/constants"
)

func Pkcs5Padding(b []byte) []byte {
	padSize := constants.AesBlockSize - len(b)%constants.AesBlockSize
	padding := bytes.Repeat([]byte{byte(padSize)}, padSize)
	return append(b, padding...)
}

func Pkcs5UnPadding(b []byte) []byte {
	padSize := int(b[len(b)-1])
	return b[:len(b)-padSize]
}
