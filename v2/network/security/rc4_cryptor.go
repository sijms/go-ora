package security

import (
	"crypto/rc4"
)

type OracleNetworkRC4Cryptor struct {
	cipher *rc4.Cipher
}

func NewOracleNetworkRC4Cryptor(key []byte) (*OracleNetworkRC4Cryptor, error) {
	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &OracleNetworkRC4Cryptor{cipher: cipher}, nil
}

func (sec *OracleNetworkRC4Cryptor) Encrypt(input []byte) ([]byte, error) {
	output := make([]byte, len(input))
	sec.cipher.XORKeyStream(output, input)
	return output, nil
}

func (sec *OracleNetworkRC4Cryptor) Decrypt(input []byte) ([]byte, error) {
	return sec.Encrypt(input)
}
