package network

import (
	"encoding/binary"
)

type RedirectPacket struct {
	Packet
	redirectAddr  string
	reconnectData string
}

func (pck *RedirectPacket) bytes() []byte {
	output := pck.Packet.bytes()
	data := append([]byte(pck.redirectAddr), 0)
	data = append(data, []byte(pck.reconnectData)...)
	binary.BigEndian.PutUint16(output[8:], uint16(len(data)))
	output = append(output, data...)
	return output
}

func newRedirectPacketFromData(packetData []byte) *RedirectPacket {
	if len(packetData) < 10 {
		return nil
	}
	pck := RedirectPacket{
		Packet: Packet{
			dataOffset: 10,
			length:     uint32(binary.BigEndian.Uint16(packetData)),
			packetType: PacketType(packetData[4]),
			flag:       packetData[5],
		},
	}
	return &pck
}
