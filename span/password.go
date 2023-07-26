package span

import (
	"crypto/aes"
	"crypto/sha512"
	"os"

	"github.com/isayme/go-logger"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/ssh/terminal"
)

const materKeySize = aes.BlockSize
const saltSize = aes.BlockSize
const fileKeySize = aes.BlockSize

func ReadPassword(promt string) (string, error) {
	logger.Info(promt)
	password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	return string(password), nil
}

func MustRandomMasterKey() []byte {
	return mustRandomBytes(materKeySize)
}

func MustRandomSalt() []byte {
	return mustRandomBytes(saltSize)
}

func GenEncryptKeyAndAuthKeyFromPassword(password string, salt []byte) (encryptKey, authKey []byte) {
	keyLen := 512 / 8
	key := pbkdf2.Key([]byte(password), salt, 100000, keyLen, sha512.New)
	encryptKey = key[0 : keyLen/2]
	authKey = key[keyLen/2:]
	return
}

func MustEncryptMasterKey(encryptKey, masterKey []byte) []byte {
	result, err := EncryptMasterKey(encryptKey, masterKey)
	if err != nil {
		panic(err)
	}

	return result
}

/**
 * masterKey 加密后存储在本地
 */
func EncryptMasterKey(encryptKey, masterKey []byte) ([]byte, error) {
	return AesEcbEncrypt(encryptKey, masterKey)
}

func MustDecryptMasterKey(encryptKey, encryptMasterKey []byte) []byte {
	result, err := DecryptMasterKey(encryptKey, encryptMasterKey)
	if err != nil {
		panic(err)
	}

	return result
}

/**
 * 解密获取 masterKey
 */
func DecryptMasterKey(encryptKey, encryptMasterKey []byte) ([]byte, error) {
	return AesEcbDecrypt(encryptKey, encryptMasterKey)
}

/**
 * 对 authKey 进行 sha256，结果存储在本地，用于下次登录密码验证
 */
func HashAuthKey(authKey []byte) []byte {
	return Sha256(authKey)
}

func RandomFileKey() ([]byte, error) {
	return randomBytes(fileKeySize)
}

func EncryptFileKey(masterKey, fileKey []byte) ([]byte, error) {
	return AesEcbEncrypt(masterKey, fileKey)
}

func DecryptFileKey(masterKey, encryptFileKey []byte) ([]byte, error) {
	return AesEcbDecrypt(masterKey, encryptFileKey)
}

func MustEncryptFileName(masterKey, fileName []byte) []byte {
	result, err := EncryptFileName(masterKey, fileName)
	if err != nil {
		panic(err)
	}

	return result
}

func EncryptFileName(masterKey, fileName []byte) ([]byte, error) {
	iv := Sha256(fileName)[0:16]
	result, err := AesCbcEncrypt(masterKey, iv, Pkcs5Padding(fileName))
	if err != nil {
		return nil, err
	}

	return append(iv, result...), nil
}

func MustDecryptFileName(masterKey, encryptFileName []byte) []byte {
	result, err := DecryptFileName(masterKey, encryptFileName)
	if err != nil {
		panic(err)
	}

	return result
}

func DecryptFileName(masterKey, encryptFileName []byte) ([]byte, error) {
	iv := encryptFileName[0:16]
	encryptFileName = encryptFileName[16:]

	result, err := AesCbcDecrypt(masterKey, iv, encryptFileName)
	if err != nil {
		return nil, err
	}

	return Pkcs5UnPadding(result), nil
}

func EncryptFileContent(masterKey, iv, content []byte) ([]byte, error) {
	result, err := AesCtrEncrypt(masterKey, iv, content)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DecryptFileContent(masterKey, iv, encryptFileContent []byte) ([]byte, error) {
	result, err := AesCtrDecrypt(masterKey, iv, encryptFileContent)
	if err != nil {
		return nil, err
	}

	return result, nil
}
