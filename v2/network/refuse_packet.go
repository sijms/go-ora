package network

import (
	"encoding/binary"
	"regexp"
	"strconv"
	"strings"
)

type RefusePacket struct {
	Packet
	Err          OracleError
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

func (pck *RefusePacket) extractErrCode() {
	pck.Err.ErrCode = 12564
	pck.Err.ErrMsg = "ORA-12564: TNS connection refused"
	if len(pck.message) == 0 {
		return
	}
	r, err := regexp.Compile(`\(\s*ERR\s*=\s*([0-9]+)\s*\)`)
	if err != nil {
		return
	}
	msg := strings.ToUpper(pck.message)
	matches := r.FindStringSubmatch(msg)
	if len(matches) != 2 {
		return
	}
	strErrCode := matches[1]
	errCode, err := strconv.ParseInt(strErrCode, 10, 32)
	if err == nil {
		pck.Err.ErrCode = int(errCode)
		pck.Err.translate()
		return
	}
	r, err = regexp.Compile(`\(\s*ERROR\s*=([A-Z0-9=()]+)`)
	if err != nil {
		return
	}
	matches = r.FindStringSubmatch(msg)
	if len(matches) != 2 {
		return
	}
	codeStr := matches[1]
	r, err = regexp.Compile(`CODE\s*=\s*([0-9]+)`)
	if err != nil {
		return
	}
	matches = r.FindStringSubmatch(codeStr)
	if len(matches) != 2 {
		return
	}
	strErrCode = matches[1]
	errCode, err = strconv.ParseInt(strErrCode, 10, 32)
	if err == nil {
		pck.Err.ErrCode = int(errCode)
		pck.Err.translate()
	}
}
