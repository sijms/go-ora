package go_ora

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sijms/go-ora/network"
)

type TCPNego struct {
	MessageCode           uint8
	ProtocolServerVersion uint8
	ProtocolServerString  string
	OracleVersion         int
	ServerCharset         int
	ServerFlags           uint8
	CharsetElem           int
	ServernCharset        int
	ServerCompileTimeCaps []byte
	ServerRuntimeCaps     []byte
}

func NewTCPNego(session *network.Session) (*TCPNego, error) {
	session.ResetBuffer()
	session.PutBytes(1, 6, 0)
	session.PutBytes([]byte("OracleClientGo\x00")...)
	err := session.Write()
	if err != nil {
		return nil, err
	}
	result := TCPNego{}
	result.MessageCode, err = session.GetByte()
	if err != nil {
		return nil, err
	}
	if result.MessageCode != 1 {
		return nil, errors.New(fmt.Sprintf("message code error: received code %d and expected code is 1", result.MessageCode))
	}
	result.ProtocolServerVersion, err = session.GetByte()
	if err != nil {
		return nil, err
	}
	switch result.ProtocolServerVersion {
	case 4:
		result.OracleVersion = 7230
	case 5:
		result.OracleVersion = 8030
	case 6:
		result.OracleVersion = 8100
	default:
		return nil, errors.New("unsupported server version")
	}
	_, _ = session.GetByte()
	result.ProtocolServerString, err = session.GetNullTermString(50)
	if err != nil {
		return nil, err
	}

	result.ServerCharset, err = session.GetInt(2, false, false)
	if err != nil {
		return nil, err
	}
	result.ServerFlags, err = session.GetByte()
	if err != nil {
		return nil, err
	}
	result.CharsetElem, err = session.GetInt(2, false, false)
	if err != nil {
		return nil, err
	}
	if result.CharsetElem > 0 {
		_, _ = session.GetBytes(result.CharsetElem * 5)
	}

	len1, err := session.GetInt(2, false, true)
	if err != nil {
		return nil, err
	}
	numArray, err := session.GetBytes(len1)
	if err != nil {
		return nil, err
	}
	num3 := int(6 + (numArray[5]) + (numArray[6]))
	result.ServernCharset = int(binary.BigEndian.Uint16(numArray[(num3 + 3):(num3 + 5)]))
	len2, err := session.GetByte()
	if err != nil {
		return nil, err
	}
	result.ServerCompileTimeCaps, err = session.GetBytes(int(len2))
	if err != nil {
		return nil, err
	}
	len3, err := session.GetByte()
	if err != nil {
		return nil, err
	}
	result.ServerRuntimeCaps, err = session.GetBytes(int(len3))
	if err != nil {
		return nil, err
	}
	if result.ServerCompileTimeCaps[15]&1 != 0 {
		session.HasEOSCapability = true
	}
	if result.ServerCompileTimeCaps[16]&1 != 0 {
		session.HasFSAPCapability = true
	}
	return &result, nil
}
