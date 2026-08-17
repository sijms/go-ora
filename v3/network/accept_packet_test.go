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

// Malformed responses (a non-Oracle peer whose bytes happen to carry the
// ACCEPT type) make newAcceptPacketFromData return nil. readPacket must
// translate that nil into an error instead of handing callers a typed-nil
// interface, which passes their type assertions and panics on first field
// access (see Session.Connect).
func TestNewAcceptPacketFromDataMalformed(t *testing.T) {
	// truncated header: 8-byte TNS header, ACCEPT type, no body
	short := make([]byte, 8)
	short[4] = uint8(ACCEPT)
	if pck := newAcceptPacketFromData(short, &configurations.ConnectionConfig{}); pck != nil {
		t.Errorf("expected nil for 8-byte accept packet, got %+v", pck)
	}

	// full-size header whose accept-data length disagrees with the actual
	// buffer length (bytes [18:20] announce 4 bytes but the data offset
	// leaves none)
	mismatch := make([]byte, 64)
	mismatch[4] = uint8(ACCEPT)
	binary.BigEndian.PutUint16(mismatch[18:], 4)  // accept data length
	binary.BigEndian.PutUint16(mismatch[20:], 64) // accept data offset
	if pck := newAcceptPacketFromData(mismatch, &configurations.ConnectionConfig{}); pck != nil {
		t.Errorf("expected nil for length-mismatch accept packet, got %+v", pck)
	}
}
