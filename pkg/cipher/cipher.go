package cipher

//go:generate sh ../../scripts/generate_key.sh
const AESKEY = "B5FFCEE73EF298A4"

type Cipher interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}
