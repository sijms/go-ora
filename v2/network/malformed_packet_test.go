package network_test

import (
	"context"
	"database/sql"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/sijms/go-ora/v2"
)

// A non-Oracle peer whose first response bytes happen to carry the TNS ACCEPT
// type (0x02) but no valid accept body used to crash the driver: the packet
// constructor returns nil for malformed input, readPacket passed that nil
// through with a nil error, and Session.Connect then dereferenced the
// typed-nil packet (`*acceptPacket.sessionCtx`) with a nil pointer panic:
//
//	panic: runtime error: invalid memory address or nil pointer dereference
//	[signal SIGSEGV: segmentation violation code=0x1 addr=0x10 ...]
//	github.com/sijms/go-ora/v2/network.(*Session).Connect(...)
//
// readPacket must surface such responses as a regular error instead.
func TestMalformedAcceptPacketReturnsErrorNotPanic(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 512)
				_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
				_, _ = c.Read(buf) // discard the connect packet
				// 8-byte TNS header: length=8, type[4]=0x02 (ACCEPT), no body
				_, _ = c.Write([]byte{0x00, 0x08, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00})
				time.Sleep(200 * time.Millisecond)
			}(conn)
		}
	}()

	port := listener.Addr().(*net.TCPAddr).Port
	db, err := sql.Open("oracle", "oracle://user:pass@127.0.0.1:"+strconv.Itoa(port)+"/ORCL")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err == nil {
		t.Fatal("expected an error from the malformed server response")
	}
	if !strings.Contains(err.Error(), "invalid accept packet") {
		t.Fatalf("expected malformed-packet error, got: %v", err)
	}
}
