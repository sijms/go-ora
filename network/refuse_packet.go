package network

import "encoding/binary"

type RefusePacket struct {
	Packet
	SystemReason uint8
	UserReason   uint8
	message      string
}

// bytes return bytearray representation of refuse packet
func (pck *RefusePacket) bytes() []byte {
	output := pck.Packet.bytes()
	output[8] = pck.SystemReason
	output[9] = pck.UserReason
	data := []byte(pck.message)
	binary.BigEndian.PutUint16(output[10:], uint16(len(data)))
	output = append(output, data...)
	return output
}

//func (pck *RefusePacket) getPacketType() PacketType {
//	return pck.packet.packetType
//}

// newRefusePacketFromData create new refuse packet from bytearray that
// has been read from network stream
func newRefusePacketFromData(packetData []byte) *RefusePacket {
	if len(packetData) < 12 {
		return nil
	}
	dataLen := binary.BigEndian.Uint16(packetData[10:])
	var message string
	if uint16(len(packetData)-1) >= 12+dataLen {
		message = string(packetData[12 : 12+dataLen])
	}
	return &RefusePacket{
		Packet: Packet{
			dataOffset: 12,
			length:     binary.BigEndian.Uint16(packetData),
			packetType: PacketType(packetData[4]),
			flag:       0,
		},
		SystemReason: packetData[9],
		UserReason:   packetData[8],
		message:      message,
	}
}
