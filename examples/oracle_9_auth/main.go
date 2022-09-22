package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
)

func getKeyFromUserNameAndPassword(username string, password string) ([]byte, error) {
	username = strings.ToUpper(username)
	password = strings.ToUpper(password)
	extendString := func(str string) []byte {
		ret := make([]byte, len(str)*2)
		for index, char := range []byte(str) {
			ret[index*2] = 0
			ret[index*2+1] = char
		}
		return ret
	}
	buffer := append(extendString(username), extendString(password)...)
	if len(buffer)%8 > 0 {
		buffer = append(buffer, make([]byte, 8-len(buffer)%8)...)
	}
	key := []byte{1, 35, 69, 103, 137, 171, 205, 239}

	DesEnc := func(input []byte, key []byte) ([]byte, error) {
		ret := make([]byte, 8)
		enc, err := des.NewCipher(key)
		if err != nil {
			return nil, err
		}
		for x := 0; x < len(input)/8; x++ {
			for y := 0; y < 8; y++ {
				ret[y] = uint8(int(ret[y]) ^ int(input[x*8+y]))
			}
			output := make([]byte, 8)
			enc.Encrypt(output, ret)
			copy(ret, output)
		}
		return ret, nil
	}
	key1, err := DesEnc(buffer, key)
	if err != nil {
		return nil, err
	}
	key2, err := DesEnc(buffer, key1)
	if err != nil {
		return nil, err
	}
	// function OSLogonHelper.Method1_bytearray (DecryptSessionKey)
	return append(key2, make([]byte, 8)...), nil
}
func decryptSessionKey2(encKey []byte, sessionKey string) ([]byte, error) {
	result, err := hex.DecodeString(sessionKey)
	if err != nil {
		return nil, err
	}
	blk, err := des.NewCipher(encKey)
	if err != nil {
		return nil, err
	}
	enc := cipher.NewCBCDecrypter(blk, make([]byte, 8))
	output := make([]byte, len(result))
	enc.CryptBlocks(output, result)
	return output, nil
}

func encryptPassword(password string, key []byte) (string, error) {
	padding := 0
	temp := []byte(password)
	if len(password)%8 > 0 {
		padding = 8 - (len(password) % 8)
		temp = append(temp, bytes.Repeat([]byte{0}, padding)...)
	}
	blk, err := des.NewCipher(key)
	if err != nil {
		return "", err
	}
	enc := cipher.NewCBCDecrypter(blk, make([]byte, 8))
	output := make([]byte, len(temp))
	enc.CryptBlocks(output, temp)
	encPassword := append(output, uint8(padding))
	// [36, -90, -28, -115, -91, 95, -80, -2]
	fmt.Println("enc password: ", encPassword)
	return hex.EncodeToString(encPassword), nil

}
func decryptSessionKey(padding bool, encKey []byte, sessionKey string) ([]byte, error) {
	result, err := hex.DecodeString(sessionKey)
	if err != nil {
		return nil, err
	}
	blk, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}
	//if padding {
	//	result = PKCS5Padding(result, blk.BlockSize())
	//}
	enc := cipher.NewCBCDecrypter(blk, make([]byte, 16))
	output := make([]byte, len(result))
	enc.CryptBlocks(output, result)
	cutLen := 0
	if padding {
		num := int(output[len(output)-1])
		if num < enc.BlockSize() {
			apply := true
			for x := len(output) - num; x < len(output); x++ {
				if output[x] != uint8(num) {
					apply = false
					break
				}
			}
			if apply {
				cutLen = int(output[len(output)-1])
			}
		}
	}
	return output[:len(output)-cutLen], nil
}
func main() {
	username := "LAB"
	password := "lab"
	encSessionKey := "CC22A4B3963665FA"
	//authSK, err := hex.DecodeString(authSKString)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	key, err := getKeyFromUserNameAndPassword(username, password)
	if err != nil {
		log.Fatalln(err)
	}
	sessionKey, err := decryptSessionKey2(key[:8], encSessionKey)
	if err != nil {
		log.Fatalln(err)
	}
	// [95 242 76 225 183 123 160 169]
	fmt.Println(sessionKey)

	fmt.Println(encryptPassword(password, sessionKey))

}
