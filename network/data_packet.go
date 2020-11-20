package network

import (
	"encoding/binary"
)

type DataPacket struct {
	packet   Packet
	dataFlag uint16
	buffer   []byte
}

func (pck *DataPacket) bytes() []byte {
	output := pck.packet.bytes()
	binary.BigEndian.PutUint16(output[8:], pck.dataFlag)
	if len(pck.buffer) > 0 {
		output = append(output, pck.buffer...)
	}
	return output
}

func (pck *DataPacket) getPacketType() PacketType {
	return pck.packet.packetType
}
func newDataPacket(initialData []byte) *DataPacket {
	return &DataPacket{
		packet: Packet{
			dataOffset: 0xA,
			length:     uint16(len(initialData) + 0xA),
			packetType: DATA,
			flag:       0,
		},
		dataFlag: 0,
		buffer:   initialData,
	}
}

func newDataPacketFromData(packetData []byte) *DataPacket {
	if len(packetData) <= 0xA || PacketType(packetData[4]) != DATA {
		return nil
	}
	return &DataPacket{
		packet: Packet{
			dataOffset: 0xA,
			length:     binary.BigEndian.Uint16(packetData),
			packetType: PacketType(packetData[4]),
			flag:       packetData[5],
		},
		dataFlag: binary.BigEndian.Uint16(packetData[8:]),
		buffer:   packetData[10:],
	}
}

//func (pck *DataPacket) Data() []byte {
//	return pck.buffer
//}
