package network

import (
	"encoding/binary"
)

type DataPacket struct {
	Packet
	dataFlag uint16
	buffer   []byte
}

// bytes return bytearray representation of data packet
func (pck *DataPacket) bytes() []byte {
	output := pck.Packet.bytes()
	binary.BigEndian.PutUint16(output[8:], pck.dataFlag)
	if len(pck.buffer) > 0 {
		output = append(output, pck.buffer...)
	}
	return output
}

// GetPacketType return packet type
//func (pck *DataPacket) getPacketType() PacketType {
//	return pck.packet.packetType
//}

// newDataPacket create a new data packet that carry initialData bytearray
// ready for network send
func newDataPacket(initialData []byte) *DataPacket {
	return &DataPacket{
		Packet: Packet{
			dataOffset: 0xA,
			length:     uint16(len(initialData) + 0xA),
			packetType: DATA,
			flag:       0,
		},
		dataFlag: 0,
		buffer:   initialData,
	}
}

// newDataPacketFromData create data packet from bytearray which is
// read from network stream
func newDataPacketFromData(packetData []byte) *DataPacket {
	if len(packetData) <= 0xA || PacketType(packetData[4]) != DATA {
		return nil
	}
	return &DataPacket{
		Packet: Packet{
			dataOffset: 0xA,
			length:     binary.BigEndian.Uint16(packetData),
			packetType: PacketType(packetData[4]),
			flag:       packetData[5],
		},
		dataFlag: binary.BigEndian.Uint16(packetData[8:]),
		buffer:   packetData[10:],
	}
}
