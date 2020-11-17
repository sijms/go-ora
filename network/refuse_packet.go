package network

import "encoding/binary"

type RefusePacket struct {
	packet Packet
	//dataOffset uint16
	//Len uint16
	//packetType PacketType
	//Flag uint8
	SystemReason uint8
	UserReason   uint8
	message      string
}

func (pck *RefusePacket) bytes() []byte {
	output := pck.packet.bytes()
	output[8] = pck.SystemReason
	output[9] = pck.UserReason
	data := []byte(pck.message)
	binary.BigEndian.PutUint16(output[10:], uint16(len(data)))
	output = append(output, data...)
	return output
}

func (pck *RefusePacket) getPacketType() PacketType {
	return pck.packet.packetType
}
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
		packet: Packet{
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
