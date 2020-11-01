package network

import (
	"encoding/binary"
	"strings"
)

type RedirectPacket struct {
	packet        Packet
	redirectAddr  string
	reconnectData string
}

func (pck *RedirectPacket) bytes() []byte {
	output := pck.packet.bytes()
	data := append([]byte(pck.redirectAddr), 0)
	data = append(data, []byte(pck.reconnectData)...)
	binary.BigEndian.PutUint16(output[8:], uint16(len(data)))
	output = append(output, data...)
	return output
}

func (pck *RedirectPacket) getPacketType() PacketType {
	return pck.packet.packetType
}

func newRedirectPacketFromData(packetData []byte) *RedirectPacket {
	if len(packetData) < 10 {
		return nil
	}
	dataLen := binary.BigEndian.Uint16(packetData[8:])

	pck := RedirectPacket{
		packet: Packet{
			dataOffset: 10,
			length:     binary.BigEndian.Uint16(packetData),
			packetType: PacketType(packetData[4]),
			flag:       packetData[5],
		},
	}
	data := string(packetData[10 : 10+dataLen])
	if pck.packet.flag&0x2 == 0 {
		pck.redirectAddr = data
		return &pck
	}
	length := strings.Index(data, "\x00")
	if length > 0 {
		pck.redirectAddr = data[:length]
		pck.reconnectData = data[length:]
	} else {
		pck.redirectAddr = data
	}
	return &pck
}
