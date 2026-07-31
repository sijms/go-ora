package network

import (
	"encoding/binary"
	"testing"

	"github.com/sijms/go-ora/v3/configurations"
)

// Servers speaking TNS protocol versions below 315 (e.g. Oracle 11g) reply
// with a 32-byte accept packet that ends after the reconnect-address fields.
// The fast-auth negotiation flags at bytes [41:45] only exist in accept
// packets from newer servers, so parsing must not read past the end of the
// short form.
func TestNewAcceptPacketFromDataShortPacket(t *testing.T) {
	packetData := make([]byte, 32)
	binary.BigEndian.PutUint16(packetData[0:], 32)      // packet length
	packetData[4] = uint8(ACCEPT)                       // packet type
	binary.BigEndian.PutUint16(packetData[8:], 314)     // protocol version
	binary.BigEndian.PutUint16(packetData[10:], 0x0801) // negotiated options
	binary.BigEndian.PutUint16(packetData[12:], 8192)   // session data unit
	binary.BigEndian.PutUint16(packetData[14:], 32767)  // transport data unit
	binary.BigEndian.PutUint16(packetData[16:], 0x7f08) // NT characteristics
	binary.BigEndian.PutUint16(packetData[18:], 0)      // accept data length
	binary.BigEndian.PutUint16(packetData[20:], 32)     // accept data offset

	pck := newAcceptPacketFromData(packetData, &configurations.ConnectionConfig{})
	if pck == nil {
		t.Fatal("expected accept packet, got nil")
	}
	if pck.sessionCtx.Version != 314 {
		t.Errorf("Version = %d, want 314", pck.sessionCtx.Version)
	}
	if pck.sessionCtx.SessionDataUnit != 8192 {
		t.Errorf("SessionDataUnit = %d, want 8192", pck.sessionCtx.SessionDataUnit)
	}
	if pck.sessionCtx.NegotiatedOptions2 != 0 {
		t.Errorf("NegotiatedOptions2 = %d, want 0 for short accept packet", pck.sessionCtx.NegotiatedOptions2)
	}
	if pck.sessionCtx.FastAuthEnabled {
		t.Error("FastAuthEnabled = true, want false for short accept packet")
	}
	if pck.sessionCtx.EODDAFlagEnabled {
		t.Error("EODDAFlagEnabled = true, want false for short accept packet")
	}
}
