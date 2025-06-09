package network

import (
	"encoding/binary"
	"regexp"
	"strconv"
	"strings"

	go_ora "github.com/sijms/go-ora/v2/lazy_init"
)

type RefusePacket struct {
	Packet
	Err          *OracleError
	SystemReason uint8
	UserReason   uint8
	message      string
}

func (pck *RefusePacket) bytes() []byte {
	output := pck.Packet.bytes()
	output[8] = pck.SystemReason
	output[9] = pck.UserReason
	data := []byte(pck.message)
	binary.BigEndian.PutUint16(output[10:], uint16(len(data)))
	output = append(output, data...)
	return output
}

func newRefusePacketFromData(packetData []byte) *RefusePacket {
	if len(packetData) < 12 {
		return nil
	}
	dataLen := binary.BigEndian.Uint16(packetData[10:])
	var message string
	if uint16(len(packetData)) >= 12+dataLen {
		message = string(packetData[12 : 12+dataLen])
	}

	return &RefusePacket{
		Packet: Packet{
			dataOffset: 12,
			length:     uint32(binary.BigEndian.Uint16(packetData)),
			packetType: PacketType(packetData[4]),
			flag:       0,
		},
		SystemReason: packetData[9],
		UserReason:   packetData[8],
		message:      message,
	}
}

var errExtractRegexp = go_ora.NewLazyInit(func() (interface{}, error) {
	return regexp.Compile(`\(\s*ERR\s*=\s*([0-9]+)\s*\)`)
})

var errorExtractRegexp = go_ora.NewLazyInit(func() (interface{}, error) {
	return regexp.Compile(`\(\s*ERROR\s*=([A-Z0-9=()]+)`)
})

var codeExtractRegexp = go_ora.NewLazyInit(func() (interface{}, error) {
	return regexp.Compile(`CODE\s*=\s*([0-9]+)`)
})

func (pck *RefusePacket) extractErrCode() {
	var err error
	pck.Err = NewOracleError(12564)
	if len(pck.message) == 0 {
		return
	}

	var errExtractRegexpAny interface{}
	errExtractRegexpAny, err = errExtractRegexp.GetValue()
	if err != nil {
		return
	}

	msg := strings.ToUpper(pck.message)
	matches := errExtractRegexpAny.(*regexp.Regexp).FindStringSubmatch(msg)
	if len(matches) != 2 {
		return
	}

	strErrCode := matches[1]
	errCode, err := strconv.ParseInt(strErrCode, 10, 32)
	if err == nil {
		pck.Err = NewOracleError(int(errCode))
		return
	}

	var errorExtractRegexpAny interface{}
	errorExtractRegexpAny, err = errorExtractRegexp.GetValue()
	if err != nil {
		return
	}

	matches = errorExtractRegexpAny.(*regexp.Regexp).FindStringSubmatch(msg)
	if len(matches) != 2 {
		return
	}
	codeStr := matches[1]

	var codeExtractRegexpAny interface{}
	codeExtractRegexpAny, err = codeExtractRegexp.GetValue()
	if err != nil {
		return
	}

	matches = codeExtractRegexpAny.(*regexp.Regexp).FindStringSubmatch(codeStr)
	if len(matches) != 2 {
		return
	}

	strErrCode = matches[1]
	errCode, err = strconv.ParseInt(strErrCode, 10, 32)
	if err == nil {
		pck.Err = NewOracleError(int(errCode))
	}
}
