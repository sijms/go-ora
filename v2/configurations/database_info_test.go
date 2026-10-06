package configurations

import "testing"

func TestUpdateDatabaseInfoForRedirectKeepsTLSHostName(t *testing.T) {
	info := DatabaseInfo{}
	err := info.UpdateDatabaseInfo(`(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=scan.example.com)(PORT=2484))(CONNECT_DATA=(SERVICE_NAME=svc)))`)
	if err != nil {
		t.Fatalf("UpdateDatabaseInfo: %v", err)
	}
	if got := info.GetActiveServer(false).TLSHostName(); got != "scan.example.com" {
		t.Fatalf("TLSHostName before redirect = %q, want %q", got, "scan.example.com")
	}

	// A SCAN listener redirects to a node listener by IP address.
	err = info.UpdateDatabaseInfoForRedirect(
		`(ADDRESS=(PROTOCOL=TCPS)(HOST=10.0.0.42)(PORT=2484))`,
		`(DESCRIPTION=(CONNECT_DATA=(SERVICE_NAME=svc)(INSTANCE_NAME=inst1)))`)
	if err != nil {
		t.Fatalf("UpdateDatabaseInfoForRedirect: %v", err)
	}
	info.ResetServerIndex()
	server := info.GetActiveServer(false)
	if server.Addr != "10.0.0.42" {
		t.Errorf("Addr after redirect = %q, want %q", server.Addr, "10.0.0.42")
	}
	if got := server.TLSHostName(); got != "scan.example.com" {
		t.Errorf("TLSHostName after redirect = %q, want %q", got, "scan.example.com")
	}

	// A second redirect keeps the host from the original descriptor.
	err = info.UpdateDatabaseInfoForRedirect(
		`(ADDRESS=(PROTOCOL=TCPS)(HOST=10.0.0.48)(PORT=2484))`,
		`(DESCRIPTION=(CONNECT_DATA=(SERVICE_NAME=svc)))`)
	if err != nil {
		t.Fatalf("second UpdateDatabaseInfoForRedirect: %v", err)
	}
	info.ResetServerIndex()
	if got := info.GetActiveServer(false).TLSHostName(); got != "scan.example.com" {
		t.Errorf("TLSHostName after second redirect = %q, want %q", got, "scan.example.com")
	}
}

func TestTLSHostNameDefaultsToAddr(t *testing.T) {
	server := ServerAddr{Addr: "10.0.0.42", Port: 2484}
	if got := server.TLSHostName(); got != "10.0.0.42" {
		t.Errorf("TLSHostName() = %q, want %q", got, "10.0.0.42")
	}
}
