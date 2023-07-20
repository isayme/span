package span

import "encoding/base64"

func Base64EncodeToString(src []byte) string {
	return base64.RawURLEncoding.EncodeToString(src)
}

func Base64DecodeString(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
