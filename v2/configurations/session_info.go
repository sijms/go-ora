package configurations

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

type DialerContext interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type SessionInfo struct {
	SSLVersion            string
	Timeout               time.Duration
	EnableOOB             bool
	UnixAddress           string
	TransportDataUnitSize uint32
	SessionDataUnitSize   uint32
	Protocol              string
	SSL                   bool
	SSLVerify             bool
	Dialer                DialerContext
}

func (si *SessionInfo) RegisterDial(dialer func(ctx context.Context, network, address string) (net.Conn, error)) {
	var temp = &customDial{DialCtx: dialer}
	si.Dialer = temp
}

type customDial struct {
	DialCtx func(ctx context.Context, network, address string) (net.Conn, error)
}

func (c *customDial) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return c.DialCtx(ctx, network, address)
}

func (si *SessionInfo) UpdateSSL(server *ServerAddr) error {
	if server != nil {
		if strings.ToLower(server.Protocol) == "tcps" {
			si.SSL = true
			return nil
		} else if strings.ToLower(server.Protocol) == "tcp" {
			si.SSL = false
			return nil
		}
	}
	if strings.ToLower(si.Protocol) == "tcp" {
		si.SSL = false
	} else if strings.ToLower(si.Protocol) == "tcps" {
		si.SSL = true
	} else {
		return fmt.Errorf("unknown or missing protocol: %s", si.Protocol)
	}
	return nil
}
