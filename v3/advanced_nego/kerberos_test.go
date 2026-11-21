package advanced_nego

import (
	"bytes"
	"errors"
	"testing"
)

type plainAuth struct{}

func (plainAuth) Authenticate(server, service string) ([]byte, error) {
	return []byte(service + "/" + server), nil
}

type mutualAuth struct {
	plainAuth
	replies *[][]byte
}

func (m mutualAuth) AuthenticateMutual(server, service string) ([]byte, func([]byte) error, error) {
	accept := func(reply []byte) error {
		*m.replies = append(*m.replies, reply)
		if len(reply) == 0 {
			return errors.New("no AP-REP")
		}
		return nil
	}
	return []byte("mutual " + service + "/" + server), accept, nil
}

func TestKerberosTicket(t *testing.T) {
	var replies [][]byte
	tests := map[string]struct {
		auth       KerberosAuthInterface
		wantTicket string
		wantAccept bool
	}{
		"Authenticate":       {auth: plainAuth{}, wantTicket: "oracle/db.example.com"},
		"AuthenticateMutual": {auth: mutualAuth{replies: &replies}, wantTicket: "mutual oracle/db.example.com", wantAccept: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ticket, accept, err := kerberosTicket(tc.auth, "db.example.com", "oracle")
			if err != nil {
				t.Fatal(err)
			}
			if string(ticket) != tc.wantTicket {
				t.Errorf("ticket %q, want %q", ticket, tc.wantTicket)
			}
			if (accept != nil) != tc.wantAccept {
				t.Fatalf("accept set %t, want %t", accept != nil, tc.wantAccept)
			}
			if accept == nil {
				return
			}
			if err := accept([]byte{0x6f}); err != nil {
				t.Errorf("accept of a reply: %v", err)
			}
			if err := accept(nil); err == nil {
				t.Error("accept of no reply: no error")
			}
			if len(replies) != 2 || !bytes.Equal(replies[0], []byte{0x6f}) {
				t.Errorf("replies %v", replies)
			}
		})
	}
}
