package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

type OracleNetworkEncryption interface {
	Encrypt(input []byte) ([]byte, error)
	Decrypt(input []byte) ([]byte, error)
}

type OracleNetworkDataEntegrity interface {
	Hash(input []byte) ([]byte, error)
}
type OracleNetworkCBCCryptor struct {
	blk cipher.Block
	iv  []byte
	//enc cipher.BlockMode
	//dec cipher.BlockMode
}

func NewOracleNetworkCBCEncrypter(key, iv []byte) (*OracleNetworkCBCCryptor, error) {
	blk, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &OracleNetworkCBCCryptor{blk: blk, iv: iv}, nil
	//return &OracleNetworkCBCCryptor{
	//	enc: cipher.NewCBCEncrypter(plk, iv),
	//	dec: cipher.NewCBCDecrypter(plk, iv),
	//}, nil
}

func (sec *OracleNetworkCBCCryptor) Encrypt(input []byte) ([]byte, error) {
	length := len(input)
	num := 0
	if length%16 > 0 {
		num = 16 - (length % 16)
	}
	if num > 0 {
		input = append(input, make([]byte, num)...)
	}
	output := make([]byte, length+num)
	enc := cipher.NewCBCEncrypter(sec.blk, sec.iv)
	enc.CryptBlocks(output, input)
	foldingKey := uint8(0)
	return append(output, uint8(num+1), foldingKey), nil
}

func (sec *OracleNetworkCBCCryptor) Decrypt(input []byte) ([]byte, error) {
	length := len(input)
	length--
	if (length-1)%16 != 0 {
		return nil, errors.New("Invalid padding from cipher text")
	}
	num := int(input[length-1])
	if num < 0 || num > 16 {
		return nil, errors.New("Invalid padding from cipher text")
	}
	output := make([]byte, length-1)
	dec := cipher.NewCBCDecrypter(sec.blk, sec.iv)
	dec.CryptBlocks(output, input[:length-1])
	return output[:length-num], nil
}
func PKCS5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padtext...)
}
