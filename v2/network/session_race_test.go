package network

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sijms/go-ora/v2/configurations"
	"github.com/sijms/go-ora/v2/trace"
)

// TestConnectDisconnectRace verifies that Session.Connect and
// Session.Disconnect do not race on session.conn when the context is
// canceled during connection establishment. Run with -race.
// See issue #736.
func TestConnectDisconnectRace(t *testing.T) {
	connOption := &configurations.ConnectionConfig{
		DatabaseInfo: configurations.DatabaseInfo{
			Servers: []configurations.ServerAddr{
				{Addr: "192.0.2.1", Port: 1521}, // TEST-NET-1: TCP handshake never completes
			},
			ServiceName: "svc",
		},
		SessionInfo: configurations.SessionInfo{
			SessionDataUnitSize:   0xFFFF,
			TransportDataUnitSize: 0xFFFF,
		},
	}

	for i := 0; i < 5; i++ {
		session := NewSession(connOption, trace.NilTracer())
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)

		var wg sync.WaitGroup
		wg.Add(2)

		// goroutine 1: Connect (writes session.conn)
		go func() {
			defer wg.Done()
			_ = session.Connect(ctx)
		}()

		// goroutine 2: wait for ctx deadline then Disconnect (reads session.conn)
		go func() {
			defer wg.Done()
			<-ctx.Done()
			session.Disconnect()
		}()

		wg.Wait()
		cancel()
	}
}
