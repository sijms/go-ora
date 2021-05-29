package network

import "encoding/binary"

//type ConnectPacket Packet
type ConnectPacket struct {
	packet     Packet
	sessionCtx SessionContext
	buffer     []byte
}

func (pck *ConnectPacket) bytes() []byte {
	output := pck.packet.bytes()
	//binary.BigEndian.PutUint16(output, pck.length)
	//output[4] = uint8(pck.packetType)
	//output[5] = pck.flag
	binary.BigEndian.PutUint16(output[8:], pck.sessionCtx.Version)
	binary.BigEndian.PutUint16(output[10:], pck.sessionCtx.LoVersion)
	binary.BigEndian.PutUint16(output[12:], pck.sessionCtx.Options)
	binary.BigEndian.PutUint16(output[14:], pck.sessionCtx.SessionDataUnit)
	binary.BigEndian.PutUint16(output[16:], pck.sessionCtx.TransportDataUnit)
	output[18] = 79
	output[19] = 152
	binary.BigEndian.PutUint16(output[22:], pck.sessionCtx.Histone)
	binary.BigEndian.PutUint16(output[24:], uint16(len(pck.buffer)))
	binary.BigEndian.PutUint16(output[26:], pck.packet.dataOffset)
	output[32] = pck.sessionCtx.ACFL0
	output[33] = pck.sessionCtx.ACFL1
	if len(pck.buffer) <= 230 {
		output = append(output, pck.buffer...)
	}
	return output

}
func (pck *ConnectPacket) getPacketType() PacketType {
	return pck.packet.packetType
}
func newConnectPacket(sessionCtx SessionContext) *ConnectPacket {
	connectData := sessionCtx.connOption.ConnectionData()
	length := uint16(len(connectData))
	if length > 230 {
		length = 0
	}
	length += 58

	sessionCtx.Histone = 1
	sessionCtx.ACFL0 = 4
	sessionCtx.ACFL1 = 4

	return &ConnectPacket{
		sessionCtx: sessionCtx,
		packet: Packet{
			dataOffset: 58,
			length:     length,
			packetType: CONNECT,
			flag:       0,
		},
		buffer: []byte(connectData),
	}
}

//DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=al-maaly)(PORT=1521))(CONNECT_DATA=(
//	SERVICE_NAME=newhospital)(CID=(PROGRAM=D:\projects\DotPeekExtractedProjects\TestClient\bin\Debug\TestClient.exe)
//	(HOST=LABMANAGER)(USER=lab))))
