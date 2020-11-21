package go_ora

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/sijms/go-ora/network"
	"strconv"
	"strings"
	"time"
)

// E infront of the variable means encrypted
type AuthObject struct {
	EServerSessKey string
	EClientSessKey string
	EPassword      string
	ServerSessKey  []byte
	ClientSessKey  []byte
	KeyHash        []byte
	Salt           string
	VerifierType   int
	tcpNego        *TCPNego
}

func NewAuthObject(username string, password string, tcpNego *TCPNego, session *network.Session) (*AuthObject, error) {
	ret := new(AuthObject)
	ret.tcpNego = tcpNego
	loop := true
	for loop {
		messageCode, err := session.GetByte()
		if err != nil {
			return nil, err
		}
		switch messageCode {
		case 4:
			session.Summary, err = network.NewSummary(session)
			if err != nil {
				return nil, err
			}
			if session.HasError() {
				return nil, errors.New(session.GetError())
			}
			loop = false
		case 8:
			dictLen, err := session.GetInt(4, true, true)
			if err != nil {
				return nil, err
			}
			for x := 0; x < dictLen; x++ {
				key, val, num, err := session.GetKeyVal()
				if err != nil {
					return nil, err
				}
				if bytes.Compare(key, []byte("AUTH_SESSKEY")) == 0 {
					ret.EServerSessKey = string(val)
				} else if bytes.Compare(key, []byte("AUTH_VFR_DATA")) == 0 {
					ret.Salt = string(val)
					ret.VerifierType = num
				}
			}
		default:
			return nil, errors.New(fmt.Sprintf("message code error: received code %d and expected code is 8", messageCode))
		}
	}

	var key []byte
	padding := false
	var err error
	if ret.VerifierType == 2361 {
		key, err = getKeyFromUserNameAndPassword(username, password)
		if err != nil {
			return nil, err
		}
	} else if ret.VerifierType == 6949 {

		if ret.tcpNego.ServerCompileTimeCaps[4]&2 == 0 {
			padding = true
		}
		result, err := HexStringToBytes(ret.Salt)
		if err != nil {
			return nil, err
		}
		result = append([]byte(password), result...)
		hash := sha1.New()
		_, err = hash.Write(result)
		if err != nil {
			return nil, err
		}
		key = hash.Sum(nil)           // 20 byte key
		key = append(key, 0, 0, 0, 0) // 24 byte key

	} else {
		return nil, errors.New("unsupported verifier type")
	}
	// get the server session key
	ret.ServerSessKey, err = decryptSessionKey(padding, key, ret.EServerSessKey)
	if err != nil {
		return nil, err
	}

	// generate new key for client
	ret.ClientSessKey = make([]byte, len(ret.ServerSessKey))
	for {
		_, err = rand.Read(ret.ClientSessKey)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(ret.ClientSessKey, ret.ServerSessKey) {
			break
		}
	}

	// encrypt the client key
	ret.EClientSessKey, err = EncryptSessionKey(padding, key, ret.ClientSessKey)
	if err != nil {
		return nil, err
	}

	// get the hash key form server and client session key
	ret.KeyHash, err = CalculateKeysHash(ret.VerifierType, ret.ServerSessKey[16:], ret.ClientSessKey[16:])
	if err != nil {
		return nil, err
	}

	// encrypt the password
	ret.EPassword, err = EncryptPassword(password, ret.KeyHash)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (obj *AuthObject) Write(connOption *network.ConnectionOption, mode LogonMode, session *network.Session) error {
	session.ResetBuffer()
	keyValSize := 22
	session.PutBytes(3, 0x73, 0)
	if len(connOption.UserID) > 0 {
		session.PutInt(1, 1, false, false)
		session.PutInt(len(connOption.UserID), 4, true, true)
	} else {
		session.PutBytes(0, 0)
	}

	if len(connOption.UserID) > 0 && len(obj.EPassword) > 0 {
		mode |= UserAndPass
	}
	session.PutUint(int(mode), 4, true, true)
	session.PutUint(1, 1, false, false)
	session.PutUint(keyValSize, 4, true, true)
	session.PutBytes(1, 1)
	if len(connOption.UserID) > 0 {
		session.PutBytes([]byte(connOption.UserID)...)
	}
	index := 0
	if len(obj.EClientSessKey) > 0 {
		session.PutKeyValString("AUTH_SESSKEY", obj.EClientSessKey, 1)
		index++
	}
	if len(obj.EPassword) > 0 {
		session.PutKeyValString("AUTH_PASSWORD", obj.EPassword, 0)
		index++
	}
	// if newpassword encrypt and add {
	//	session.PutKeyValString("AUTH_NEWPASSWORD", ENewPassword, 0)
	//	index ++
	//}
	session.PutKeyValString("AUTH_TERMINAL", connOption.ClientData.HostName, 0)
	index++
	session.PutKeyValString("AUTH_PROGRAM_NM", connOption.ClientData.ProgramName, 0)
	index++
	session.PutKeyValString("AUTH_MACHINE", connOption.ClientData.HostName, 0)
	index++
	session.PutKeyValString("AUTH_PID", fmt.Sprintf("%d", connOption.ClientData.PID), 0)
	index++
	session.PutKeyValString("AUTH_SID", connOption.ClientData.UserName, 0)
	index++
	session.PutKeyValString("AUTH_CONNECT_STRING", connOption.ConnectionData(), 0)
	index++
	session.PutKeyValString("SESSION_CLIENT_CHARSET", strconv.Itoa(int(obj.tcpNego.ServerCharset)), 0)
	index++
	session.PutKeyValString("SESSION_CLIENT_LIB_TYPE", "0", 0)
	index++
	session.PutKeyValString("SESSION_CLIENT_DRIVER_NAME", connOption.ClientData.DriverName, 0)
	index++
	session.PutKeyValString("SESSION_CLIENT_VERSION", "1.0.0.0", 0)
	index++
	session.PutKeyValString("SESSION_CLIENT_LOBATTR", "1", 0)
	index++
	_, offset := time.Now().Zone()
	tz := ""
	if offset == 0 {
		tz = "00:00"
	} else {
		hours := int8(offset / 3600)

		minutes := int8((offset / 60) % 60)
		if minutes < 0 {
			minutes = minutes * -1
		}
		tz = fmt.Sprintf("%+03d:%02d", hours, minutes)
	}
	//if !strings.Contains(tz, ":") {
	//	tz += ":00"
	//}
	//session.PutKeyValString("AUTH_ALTER_SESSION",
	//	fmt.Sprintf("ALTER SESSION SET NLS_LANGUAGE='ARABIC' NLS_TERRITORY='SAUDI ARABIA'  TIME_ZONE='%s'\x00", tz), 1)
	session.PutKeyValString("AUTH_ALTER_SESSION",
		fmt.Sprintf("ALTER SESSION SET NLS_LANGUAGE='AMERICAN' NLS_TERRITORY='AMERICA'  TIME_ZONE='%s'\x00", tz), 1)
	index++
	//if (!string.IsNullOrEmpty(proxyClientName))
	//{
	//	keys[index1] = this.m_authProxyClientName;
	//	values[index1++] = this.m_marshallingEngine.m_dbCharSetConv.ConvertStringToBytes(proxyClientName, 0, proxyClientName.Length, true);
	//}
	//if (sessionId != -1)
	//{
	//	keys[index1] = this.m_authSessionId;
	//	values[index1++] = this.m_marshallingEngine.m_dbCharSetConv.ConvertStringToBytes(sessionId.ToString(), 0, sessionId.ToString().Length, true);
	//}
	//if (serialNum != -1)
	//{
	//	keys[index1] = this.m_authSerialNum;
	//	values[index1++] = this.m_marshallingEngine.m_dbCharSetConv.ConvertStringToBytes(serialNum.ToString(), 0, serialNum.ToString().Length, true);
	//}
	// fill remaining values with zeros
	for index < keyValSize {
		session.PutKeyVal(nil, nil, 0)
		index++
	}
	err := session.Write()
	if err != nil {
		return err
	}
	return nil
}

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

func PKCS5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padtext...)
}

func HexStringToBytes(input string) ([]byte, error) {
	result := make([]byte, len(input)/2)
	for x := 0; x < len(input); x += 2 {
		num, err := strconv.ParseUint(input[x:x+2], 16, 8)
		if err != nil {
			return nil, err
		}
		result[x/2] = uint8(num)
	}
	return result, nil
}
func decryptSessionKey(padding bool, encKey []byte, sessionKey string) ([]byte, error) {
	result, err := HexStringToBytes(sessionKey)
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

func EncryptSessionKey(padding bool, encKey []byte, sessionKey []byte) (string, error) {
	blk, err := aes.NewCipher(encKey)
	if err != nil {
		return "", err
	}
	enc := cipher.NewCBCEncrypter(blk, make([]byte, 16))
	if padding {
		sessionKey = PKCS5Padding(sessionKey, blk.BlockSize())
	}
	output := make([]byte, len(sessionKey))
	enc.CryptBlocks(output, sessionKey)
	return fmt.Sprintf("%X", output), nil
}

func EncryptPassword(password string, key []byte) (string, error) {
	buff1 := make([]byte, 0x10)
	_, err := rand.Read(buff1)
	//buff_1 = []byte{109, 250, 127, 252, 157, 165, 29, 6, 165, 174, 50, 93, 165, 202, 192, 100}
	if err != nil {
		return "", nil
	}
	buffer := append(buff1, []byte(password)...)
	return EncryptSessionKey(true, key, buffer)
}

func CalculateKeysHash(verifierType int, key1 []byte, key2 []byte) ([]byte, error) {
	hash := md5.New()
	switch verifierType {
	case 2361:
		buffer := make([]byte, 16)
		for x := 0; x < 16; x++ {
			buffer[x] = key1[x] ^ key2[x]
		}

		_, err := hash.Write(buffer)
		if err != nil {
			return nil, err
		}
		return hash.Sum(nil), nil
	case 6949:
		buffer := make([]byte, 24)
		for x := 0; x < 24; x++ {
			buffer[x] = key1[x] ^ key2[x]
		}
		_, err := hash.Write(buffer[:16])
		if err != nil {
			return nil, err
		}
		ret := hash.Sum(nil)
		hash.Reset()
		_, err = hash.Write(buffer[16:])
		if err != nil {
			return nil, err
		}
		ret = append(ret, hash.Sum(nil)...)
		return ret[:24], nil
	}
	return nil, nil
}

func (obj *AuthObject) VerifyResponse(response string) bool {
	key, err := decryptSessionKey(true, obj.KeyHash, response)
	if err != nil {
		fmt.Println(err)
		return false
	}
	//fmt.Printf("%#v\n", key)
	return bytes.Compare(key[16:], []byte{83, 69, 82, 86, 69, 82, 95, 84, 79, 95, 67, 76, 73, 69, 78, 84}) == 0
	//KZSR_SVR_RESPONSE = new byte[16]{ (byte) 83, (byte) 69, (byte) 82, (byte) 86, (byte) 69, (byte) 82, (byte) 95, (byte) 84, (byte) 79,
	//(byte) 95, (byte) 67, (byte) 76, (byte) 73, (byte) 69, (byte) 78, (byte) 84 };

}
