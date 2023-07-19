package span

import (
	"os"

	"github.com/isayme/go-logger"
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
