package go_ora

import (
	"crypto/md5"
	"encoding/base64"
	"sync"

	"github.com/sijms/go-ora/v3/network"
)

type ConnectionCookie struct {
	ServerCharset         int
	ServerFlags           uint8
	ServerNCharset        int
	OracleVersion         int
	ProtocolServerString  string
	ServerCompileTimeCaps []byte
	ServerRuntimeCaps     []byte
	cookieKey             string
}

var (
	cookieStore = make(map[string]*ConnectionCookie)
	cookieMu    sync.Mutex
)

func getCookieKey(conn *Connection) string {
	uuid := conn.session.Context.UUID
	if len(uuid) == 0 {
		return ""
	}
	serviceName := conn.connOption.ServiceName
	if serviceName == "" {
		serviceName = conn.connOption.SID
	}
	if serviceName == "" {
		return ""
	}
	raw := base64.StdEncoding.EncodeToString(uuid) + "|" + serviceName
	h := md5.Sum([]byte(raw))
	return string(h[:])
}

func lookupCookie(conn *Connection) *ConnectionCookie {
	key := getCookieKey(conn)
	if key == "" {
		return nil
	}
	cookieMu.Lock()
	defer cookieMu.Unlock()
	return cookieStore[key]
}

func saveCookie(conn *Connection) {
	key := getCookieKey(conn)
	if key == "" || conn.tcpNego == nil {
		return
	}
	nego := conn.tcpNego
	cookie := &ConnectionCookie{
		ServerCharset:         nego.ServerCharset,
		ServerFlags:           nego.ServerFlags,
		ServerNCharset:        nego.ServerNCharset,
		OracleVersion:         nego.OracleVersion,
		ProtocolServerString:  nego.ProtocolServerString,
		ServerCompileTimeCaps: append([]byte(nil), nego.ServerCompileTimeCaps...),
		ServerRuntimeCaps:     append([]byte(nil), nego.ServerRuntimeCaps...),
		cookieKey:             key,
	}
	cookieMu.Lock()
	cookieStore[key] = cookie
	cookieMu.Unlock()
}

func (c *ConnectionCookie) writeTTICookie(session *network.Session) {
	session.PutBytes(30, 1)
	session.PutBytes(uint8(c.OracleVersion))
	session.PutUint(c.ServerCharset, 2, false, false)
	session.PutBytes(c.ServerFlags)
	session.PutUint(c.ServerNCharset, 2, false, false)
	session.PutBytes(uint8(len(c.ProtocolServerString) + 1))
	session.PutBytes([]byte(c.ProtocolServerString)...)
	session.PutBytes(0)
	session.PutBytes(uint8(len(c.ServerCompileTimeCaps)))
	session.PutBytes(c.ServerCompileTimeCaps...)
	session.PutBytes(uint8(len(c.ServerRuntimeCaps)))
	session.PutBytes(c.ServerRuntimeCaps...)
}
