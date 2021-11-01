package network

import (
	"encoding/binary"
	"strings"
)

type RedirectPacket struct {
	Packet
	redirectAddr  string
	reconnectData string
}

// bytes return bytearray representation of redirect packet
func (pck *RedirectPacket) bytes() []byte {
	output := pck.Packet.bytes()
	data := append([]byte(pck.redirectAddr), 0)
	data = append(data, []byte(pck.reconnectData)...)
	binary.BigEndian.PutUint16(output[8:], uint16(len(data)))
	output = append(output, data...)
	return output
}

// newRedirectPacketFromData create new redirect packet from bytearray that has
// been read from network stream
func newRedirectPacketFromData(packetData []byte) *RedirectPacket {
	if len(packetData) < 10 {
		return nil
	}
	pck := RedirectPacket{
		Packet: Packet{
			dataOffset: 10,
			length:     binary.BigEndian.Uint16(packetData),
			packetType: PacketType(packetData[4]),
			flag:       packetData[5],
		},
	}
	//data := string(packetData[10 : 10+dataLen])
	//if pck.packet.flag&0x2 == 0 {
	//	pck.redirectAddr = data
	//	return &pck
	//}
	//length := strings.Index(data, "\x00")
	//if length > 0 {
	//	pck.redirectAddr = data[:length]
	//	pck.reconnectData = data[length:]
	//} else {
	//	pck.redirectAddr = data
	//}
	return &pck
}

// findValue search in the redirectArra which contain data in form key=value
// for appropriate value that related to entered key
func (pck *RedirectPacket) findValue(key string) string {
	redirectAddr := strings.ToUpper(pck.redirectAddr)
	start := strings.Index(redirectAddr, key)
	if start < 0 {
		return ""
	}
	end := strings.Index(redirectAddr[start:], ")")
	if end < 0 {
		return ""
	}
	end = start + end
	substr := pck.redirectAddr[start:end]
	words := strings.Split(substr, "=")
	if len(words) == 2 {
		return strings.TrimSpace(words[1])
	} else {
		return ""
	}
}

// protocol return value of protocol key
func (pck *RedirectPacket) protocol() string {
	return strings.ToLower(pck.findValue("PROTOCOL"))
}

// host return value of host key
func (pck *RedirectPacket) host() string {
	return pck.findValue("HOST")
}

// port return value of port key
func (pck *RedirectPacket) port() string {
	return pck.findValue("PORT")
}
