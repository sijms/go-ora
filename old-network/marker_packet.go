package network

import "encoding/binary"

type MarkerPacket struct {
	Packet
	//length     uint16
	//packetType PacketType
	//flag       uint8
	markerData uint8
	markerType uint8
}

// bytes return bytearray representation of marker packet
func (pck *MarkerPacket) bytes() []byte {
	return []byte{0, 0xB, 0, 0, 0xC, 0, 0, 0, pck.markerType, 0, pck.markerData}
}

// GetPacketType return packet type
//func (pck *MarkerPacket) getPacketType() PacketType {
//	return pck.packet.packetType
//}

// newMarkerPacket create new marker packet from marker data
func newMarkerPacket(markerData uint8) *MarkerPacket {
	return &MarkerPacket{
		Packet: Packet{
			dataOffset: 0,
			length:     0xB,
			packetType: MARKER,
			flag:       0,
		},
		markerType: 1,
		markerData: markerData,
	}
}

// newMarkerPacketFromData create marker packet from bytearray data that has been
// read from network stream
func newMarkerPacketFromData(packetData []byte) *MarkerPacket {
	if len(packetData) != 0xB {
		return nil
	}
	pck := MarkerPacket{
		Packet: Packet{
			dataOffset: 0,
			length:     binary.BigEndian.Uint16(packetData),
			packetType: PacketType(packetData[4]),
			flag:       packetData[5],
		},
		markerType: packetData[8],
		markerData: packetData[10],
	}
	if pck.packetType != MARKER {
		return nil
	}
	return &pck
}
