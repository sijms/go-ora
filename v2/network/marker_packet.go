package network

import "encoding/binary"

type MarkerPacket struct {
	Packet
	markerData uint8
	markerType uint8
}

func (pck *MarkerPacket) bytes() []byte {
	if pck.sessionCtx.handshakeComplete && pck.sessionCtx.Version >= 315 {
		return []byte{0, 0x0, 0, 0xB, 0xC, 0, 0, 0, pck.markerType, 0, pck.markerData}
	} else {
		return []byte{0, 0xB, 0, 0, 0xC, 0, 0, 0, pck.markerType, 0, pck.markerData}
	}
}

func newMarkerPacket(markerData uint8, sessionCtx *SessionContext) *MarkerPacket {
	return &MarkerPacket{
		Packet: Packet{
			sessionCtx: sessionCtx,
			dataOffset: 0,
			length:     0xB,
			packetType: MARKER,
			flag:       0x20,
		},
		markerType: 1,
		markerData: markerData,
	}
}
func newMarkerPacketFromData(packetData []byte, sessionCtx *SessionContext) *MarkerPacket {
	if len(packetData) != 0xB {
		return nil
	}
	pck := MarkerPacket{
		Packet: Packet{
			sessionCtx: sessionCtx,
			dataOffset: 0,
			packetType: PacketType(packetData[4]),
			flag:       packetData[5],
		},
		markerType: packetData[8],
		markerData: packetData[10],
	}
	if sessionCtx.handshakeComplete && sessionCtx.Version >= 315 {
		pck.length = binary.BigEndian.Uint32(packetData)
	} else {
		pck.length = uint32(binary.BigEndian.Uint16(packetData))
	}
	if pck.packetType != MARKER {
		return nil
	}
	return &pck
}
