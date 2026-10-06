package network

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sijms/go-ora/v2/configurations"
	"github.com/sijms/go-ora/v2/trace"
)

// newTestCertificate returns a CA pool and a server certificate whose only SAN is dnsName.
func newTestCertificate(t *testing.T, dnsName string) (*x509.CertPool, tls.Certificate) {
	t.Helper()
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: dnsName},
		DNSNames:     []string{dnsName},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(caCert)
	return pool, tls.Certificate{Certificate: [][]byte{serverDER}, PrivateKey: serverKey}
}

// handshakeAfterRedirect runs negotiate after a redirect from scan.example.com to 10.0.0.42.
func handshakeAfterRedirect(t *testing.T, certName string, useTLSConfig bool) error {
	t.Helper()
	pool, serverCert := newTestCertificate(t, certName)

	connOption := &configurations.ConnectionConfig{
		DatabaseInfo: configurations.DatabaseInfo{
			Servers: []configurations.ServerAddr{
				{Protocol: "TCPS", Addr: "scan.example.com", Port: 2484},
			},
			ServiceName: "svc",
		},
		SessionInfo: configurations.SessionInfo{
			SessionDataUnitSize:   0xFFFF,
			TransportDataUnitSize: 0xFFFF,
			SSLVerify:             true,
		},
	}
	if useTLSConfig {
		connOption.TLSConfig = &tls.Config{RootCAs: pool}
	}
	err := connOption.UpdateDatabaseInfoForRedirect(
		`(ADDRESS=(PROTOCOL=TCPS)(HOST=10.0.0.42)(PORT=2484))`,
		`(DESCRIPTION=(CONNECT_DATA=(SERVICE_NAME=svc)))`)
	if err != nil {
		t.Fatalf("UpdateDatabaseInfoForRedirect: %v", err)
	}
	connOption.ResetServerIndex()

	// not net.Pipe: it is unbuffered and deadlocks when the client sends an alert
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		serverConn, err := listener.Accept()
		if err != nil {
			return
		}
		defer serverConn.Close()
		_ = tls.Server(serverConn, &tls.Config{Certificates: []tls.Certificate{serverCert}}).Handshake()
	}()
	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer clientConn.Close()
	if err := clientConn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		t.Fatal(err)
	}

	session := NewSession(connOption, trace.NilTracer())
	session.conn = clientConn
	session.SSL.roots = pool
	session.negotiate()
	return session.sslConn.Handshake()
}

func TestNegotiateVerifiesOriginalHostAfterRedirect(t *testing.T) {
	for _, tc := range []struct {
		name         string
		useTLSConfig bool
	}{
		{name: "wallet roots", useTLSConfig: false},
		{name: "caller tls.Config", useTLSConfig: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := handshakeAfterRedirect(t, "scan.example.com", tc.useTLSConfig); err != nil {
				t.Fatalf("TLS handshake after redirect failed: %v", err)
			}
		})
	}
}

func TestNegotiateStillVerifiesHostName(t *testing.T) {
	err := handshakeAfterRedirect(t, "other.example.com", false)
	if err == nil {
		t.Fatal("TLS handshake succeeded with a certificate for the wrong host")
	}
	if !strings.Contains(err.Error(), "scan.example.com") {
		t.Errorf("error = %v, want a host name mismatch for scan.example.com", err)
	}
}

func TestNegotiateDoesNotModifyCallerTLSConfig(t *testing.T) {
	shared := &tls.Config{}
	var wg sync.WaitGroup
	for _, addr := range []string{"a.example.com", "b.example.com", "c.example.com", "d.example.com"} {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			connOption := &configurations.ConnectionConfig{
				DatabaseInfo: configurations.DatabaseInfo{
					Servers: []configurations.ServerAddr{
						{Protocol: "TCPS", Addr: addr, Port: 2484},
					},
					ServiceName: "svc",
				},
				SessionInfo: configurations.SessionInfo{
					SessionDataUnitSize:   0xFFFF,
					TransportDataUnitSize: 0xFFFF,
					SSLVerify:             true,
				},
			}
			connOption.TLSConfig = shared
			connOption.ResetServerIndex()
			clientConn, serverConn := net.Pipe()
			defer clientConn.Close()
			defer serverConn.Close()
			session := NewSession(connOption, trace.NilTracer())
			session.conn = clientConn
			session.negotiate()
		}(addr)
	}
	wg.Wait()
	if shared.ServerName != "" {
		t.Errorf("caller tls.Config ServerName = %q, want it left empty", shared.ServerName)
	}
}
