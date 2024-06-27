package cipher

import (
	"errors"

	"github.com/lucheng0127/virtuallan/pkg/utils"
)

//go:generate sh ../../scripts/generate_key.sh
var AESKEY = ""

type Cipher interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

func SetAESKey(key string) error {
	if !utils.ValidateKey(key) {
		return errors.New("invalid key size, 16, 24 or 32")
	}

	AESKEY = key
	return nil
}
