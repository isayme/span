package utils

import (
	"crypto/sha512"
	"os"
	"span/internal/constants"

	"github.com/isayme/go-logger"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/ssh/terminal"
)

func ReadPassword(promt string) (string, error) {
	logger.Info(promt)
	password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	return string(password), nil
}

// MustRandomMasterKey init random master key， 16bytes
func MustRandomMasterKey() []byte {
	return mustRandomBytes(constants.MaterKeySize)
}

// MustRandomSalt init random salt， 16bytes
func MustRandomSalt() []byte {
	return mustRandomBytes(constants.SaltSize)
}

func GenEncryptKeyAndAuthKeyFromPassword(password string, salt []byte) (encryptKey, authKey []byte) {
	keyLen := 256 / 8
	// Derived Key = PBKDF2-HMAC-SHA-512( Password , Salt , Iterations , Length )
	key := pbkdf2.Key([]byte(password), salt, 100000, keyLen, sha512.New)
	encryptKey = key[0 : keyLen/2]
	authKey = key[keyLen/2:]
	return
}

// MustEncryptMasterKey encryt master key. Encrypted Master Key = AES-ECB( Derived Encryption Key , Master Key )
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

func MustDecryptMasterKey(encryptKey, encryptedMasterKey []byte) []byte {
	result, err := DecryptMasterKey(encryptKey, encryptedMasterKey)
	if err != nil {
		panic(err)
	}

	return result
}

/**
 * 解密获取 masterKey
 */
func DecryptMasterKey(encryptKey, encryptedMasterKey []byte) ([]byte, error) {
	return AesEcbDecrypt(encryptKey, encryptedMasterKey)
}

/**
 * 对 authKey 进行 sha256，结果存储在本地，用于下次登录密码验证
 */
func HashAuthKey(authKey []byte) []byte {
	return Sha256(authKey)
}

func randomFileKey() ([]byte, error) {
	return randomBytes(constants.FileKeySize)
}

func MustRandomFileKey() []byte {
	result, err := randomFileKey()
	if err != nil {
		panic(err)
	}

	return result
}

func EncryptFileKey(masterKey, fileKey []byte) ([]byte, error) {
	return AesEcbEncrypt(masterKey, fileKey)
}

func DecryptFileKey(masterKey, encryptedFileKey []byte) ([]byte, error) {
	return AesEcbDecrypt(masterKey, encryptedFileKey)
}

func EncryptFileContent(masterKey, iv, content []byte) ([]byte, error) {
	result, err := aesCtrEncrypt(masterKey, iv, content)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DecryptFileContent(fileKey, iv, encryptFileContent []byte) ([]byte, error) {
	result, err := aesCtrDecrypt(fileKey, iv, encryptFileContent)
	if err != nil {
		return nil, err
	}

	return result, nil
}
