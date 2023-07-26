package span

import "github.com/nbutton23/zxcvbn-go"

/**
 * 检测密码强度
 */
func IsPasswordTooWeak(password string) bool {
	result := zxcvbn.PasswordStrength(password, nil)
	if result.Score < 4 {
		return true
	}

	return false
}
