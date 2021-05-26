package network

import (
	"bytes"
	"encoding/binary"
)

type DataPacket struct {
	Packet
	sessionCtx *SessionContext
	dataFlag   uint16
	buffer     []byte
}

func (pck *DataPacket) bytes() []byte {
	output := bytes.Buffer{}
	temp := make([]byte, 0xA)
	if pck.sessionCtx.handshakeComplete && pck.sessionCtx.Version >= 315 {
		binary.BigEndian.PutUint32(temp, pck.length)
	} else {
		binary.BigEndian.PutUint16(temp, uint16(pck.length))
	}
	temp[4] = uint8(pck.packetType)
	temp[5] = pck.flag
	binary.BigEndian.PutUint16(temp[8:], pck.dataFlag)
	output.Write(temp)
	if len(pck.buffer) > 0 {
		output.Write(pck.buffer)
	}
	return output.Bytes()
}

func newDataPacket(initialData []byte, sessionCtx *SessionContext) *DataPacket {
	return &DataPacket{
		Packet: Packet{
			dataOffset: 0xA,
			length:     uint32(len(initialData)) + 0xA,
			packetType: DATA,
			flag:       0,
		},
		sessionCtx: sessionCtx,
		dataFlag:   0,
		buffer:     initialData,
	}
}

func newDataPacketFromData(packetData []byte, sessionCtx *SessionContext) *DataPacket {
	if len(packetData) <= 0xA || PacketType(packetData[4]) != DATA {
		return nil
	}
	pck := &DataPacket{
		Packet: Packet{
			dataOffset: 0xA,
			//length:     binary.BigEndian.Uint16(packetData),
			packetType: PacketType(packetData[4]),
			flag:       packetData[5],
		},
		sessionCtx: sessionCtx,
		dataFlag:   binary.BigEndian.Uint16(packetData[8:]),
		buffer:     packetData[10:],
	}
	if sessionCtx.handshakeComplete && sessionCtx.Version >= 315 {
		pck.length = binary.BigEndian.Uint32(packetData)
	} else {
		pck.length = uint32(binary.BigEndian.Uint16(packetData))
	}
	return pck
}

//func (pck *DataPacket) Data() []byte {
//	return pck.buffer
//}
