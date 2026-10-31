package network

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/trace"
)

// TestBreakFlagRace verifies that the break flag the StartContext watchdog sets when the context is canceled is
// read under the session lock while the operation runs. Run with -race. See issue #767.
func TestBreakFlagRace(t *testing.T) {
	connOption := &configurations.ConnectionConfig{
		SessionInfo: configurations.SessionInfo{
			SessionDataUnitSize:   0xFFFF,
			TransportDataUnitSize: 0xFFFF,
		},
	}
	for i := 0; i < 5; i++ {
		session := NewSession(connOption, trace.NilTracer())
		client, server := net.Pipe()
		go func() { _, _ = io.Copy(io.Discard, server) }()
		session.conn = client
		session.Connected = true

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		done := session.StartContext(ctx)
		deadline := time.Now().Add(5 * time.Second)
		for !session.IsBreak() && time.Now().Before(deadline) {
		}
		session.EndContext(done)
		if !session.IsBreak() {
			t.Fatal("the canceled context did not break the connection")
		}
		_ = client.Close()
		_ = server.Close()
	}
}
