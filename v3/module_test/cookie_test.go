package v2_test

import (
	"context"
	"database/sql"
	"net"
	"os"
	"sync/atomic"
	"testing"

	go_ora "github.com/sijms/go-ora/v3"
)

type countingDialer struct {
	dials atomic.Int32
}

func (d *countingDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d.dials.Add(1)
	var dialer net.Dialer
	return dialer.DialContext(ctx, network, address)
}

// TestCookieRefused opens connections one after another: a server that
// refuses the connection cookie (#732) is dialed once a connection, not
// twice, after its first refusal.
func TestCookieRefused(t *testing.T) {
	dsn := os.Getenv("ORACLE_DSN")
	if dsn == "" {
		t.Skip("ORACLE_DSN is not set")
	}
	connector := go_ora.NewConnector(dsn).(*go_ora.OracleConnector)
	dialer := &countingDialer{}
	connector.Dialer(dialer)
	db := sql.OpenDB(connector)
	defer func() {
		_ = db.Close()
	}()
	db.SetMaxIdleConns(0)
	const opens = 4
	for range opens {
		conn, err := db.Conn(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if err := conn.PingContext(t.Context()); err != nil {
			t.Fatal(err)
		}
		_ = conn.Close()
	}
	// The second open may be refused once, with a cookie the first saved.
	if dials := dialer.dials.Load(); dials > opens+1 {
		t.Errorf("%d dials for %d connections", dials, opens)
	}
}
