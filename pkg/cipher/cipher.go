package cipher

// TODO Change AESKEY configurable
//
//go:generate sh ../../scripts/generate_key.sh
var AESKEY = ""

type Cipher interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}
