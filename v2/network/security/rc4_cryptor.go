package security

import (
	"crypto/rc4"
)

type OracleNetworkRC4Cryptor struct {
	encryptCipher *rc4.Cipher
	decryptCipher *rc4.Cipher
}

func NewOracleNetworkRC4Cryptor(key []byte) (*OracleNetworkRC4Cryptor, error) {
	encryptCipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decryptCipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &OracleNetworkRC4Cryptor{
		encryptCipher: encryptCipher,
		decryptCipher: decryptCipher,
	}, nil
}

func (sec *OracleNetworkRC4Cryptor) Encrypt(input []byte) ([]byte, error) {
	//fmt.Println("encrypt string:", string(input))
	sec.encryptCipher.XORKeyStream(input, input)
	return input, nil
}

func (sec *OracleNetworkRC4Cryptor) Decrypt(input []byte) ([]byte, error) {
	//fmt.Println("decrypt: ")
	sec.decryptCipher.XORKeyStream(input, input)
	return input, nil
}
