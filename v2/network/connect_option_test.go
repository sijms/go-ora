package network

import (
	"testing"

	"github.com/sijms/go-ora/v2/configurations"
)

func TestExtractServers(t *testing.T) {
	text := `(DESCRIPTION=
(ADDRESS_LIST=(LOAD_BALANCE=OFF)(FAILOVER=ON)
(ADDRESS=(PROTOCOL=tcp)(HOST=host_dguard)(PORT=1521))
(ADDRESS=(PROTOCOL=tcp)(HOST=host_active)(PORT=1521))
)
(CONNECT_DATA=(SERVICE_NAME=service)(SERVER=DEDICATED))
)`
	t.Log(configurations.ExtractServers(text))
	text = `(DESCRIPTION_LIST=(LOAD_BALANCE=off)(FAILOVER=on)
(DESCRIPTION=(CONNECT_TIMEOUT=5)
(ADDRESS=(PROTOCOL=TCP)(HOST=host_dguard)(PORT=1521))
(CONNECT_DATA=(SERVICE_NAME=service)(SERVER=DEDICATED))
)
(DESCRIPTION=(CONNECT_TIMEOUT=5)
(ADDRESS=(PROTOCOL=TCP)(HOST=host_active)(PORT=1521))
(CONNECT_DATA=(SERVICE_NAME=service)(SERVER=DEDICATED))
)
)`
	t.Log(configurations.ExtractServers(text))
}

func TestUpdateDatabaseInfo(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedConnStr string
	}{
		{
			name:            "Standard connection string",
			input:           `(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=host.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=service)))`,
			expectedConnStr: `(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=host.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=service)))`,
		},
		{
			name:            "Quoted ssl_server_cert_dn",
			input:           `(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=host.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=service))(SECURITY=(SSL_SERVER_CERT_DN="CN=cname,O=org,L=location")))`,
			expectedConnStr: `(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=host.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=service))(SECURITY=(SSL_SERVER_CERT_DN=CN=cname,O=org,L=location)))`,
		},
		{
			name:            "Unquoted ssl_server_cert_dn",
			input:           `(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=host.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=service))(SECURITY=(SSL_SERVER_CERT_DN=CN=cname,O=org,L=location)))`,
			expectedConnStr: `(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=host.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=service))(SECURITY=(SSL_SERVER_CERT_DN=CN=cname,O=org,L=location)))`,
		},
		{
			name:            "Quoted ssl_server_cert_dn with escaped parentheses",
			input:           `(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=host.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=service))(SECURITY=(SSL_SERVER_CERT_DN="CN=My \) Name")))`,
			expectedConnStr: `(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=host.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=service))(SECURITY=(SSL_SERVER_CERT_DN=CN=My \) Name)))`,
		},
		{
			name:            "Quoted service_name",
			input:           `(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=host.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME="service_name"))(SECURITY=(SSL_SERVER_CERT_DN=CN=cname)))`,
			expectedConnStr: `(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=host.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=service_name))(SECURITY=(SSL_SERVER_CERT_DN=CN=cname)))`,
		},
		{
			name: "Failover description list",
			input: `(DESCRIPTION_LIST=(LOAD_BALANCE=off)(FAILOVER=on)
 (DESCRIPTION=(CONNECT_TIMEOUT=5)(ADDRESS=(PROTOCOL=TCP)
  (HOST=dataguard_host)(PORT=1521))
  (CONNECT_DATA=(SERVICE_NAME=SERVICE_RO)(SERVER=DEDICATED)))
 (DESCRIPTION=(CONNECT_TIMEOUT=5)(ADDRESS=(PROTOCOL=TCP)
  (HOST=active_instance)(PORT=1521))
  (CONNECT_DATA=(SERVICE_NAME=SERVICE)(SERVER=DEDICATED))))`,
			expectedConnStr: `(DESCRIPTION_LIST=(LOAD_BALANCE=off)(FAILOVER=on) (DESCRIPTION=(CONNECT_TIMEOUT=5)(ADDRESS=(PROTOCOL=TCP)  (HOST=dataguard_host)(PORT=1521))  (CONNECT_DATA=(SERVICE_NAME=SERVICE_RO)(SERVER=DEDICATED))) (DESCRIPTION=(CONNECT_TIMEOUT=5)(ADDRESS=(PROTOCOL=TCP)  (HOST=active_instance)(PORT=1521))  (CONNECT_DATA=(SERVICE_NAME=SERVICE)(SERVER=DEDICATED))))`,
		},
		{
			name:            "Quoted wallet info",
			input:           `(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=host.com)(PORT=1521))(CONNECT_DATA=(SERVER=DEDICATED)(SERVICE_NAME=service))(SECURITY=(SSL_SERVER_CERT_DN="CN=cname,O=org,L=location")))`,
			expectedConnStr: `(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=host.com)(PORT=1521))(CONNECT_DATA=(SERVER=DEDICATED)(SERVICE_NAME=service))(SECURITY=(SSL_SERVER_CERT_DN=CN=cname,O=org,L=location)))`,
		},
		{
			name:            "Multiple descriptions list",
			input:           `(DESCRIPTION_LIST=(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=host1.domain.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=ServiceName)))(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=host2.domain.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=ServiceName))))`,
			expectedConnStr: `(DESCRIPTION_LIST=(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=host1.domain.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=ServiceName)))(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=host2.domain.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=ServiceName))))`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			op := &configurations.ConnectionConfig{}
			err := op.UpdateDatabaseInfo(tc.input)
			if err != nil {
				t.Fatalf("UpdateDatabaseInfo failed: %v", err)
			}
			gotConnStr := op.ConnectionData()
			if gotConnStr != tc.expectedConnStr {
				t.Errorf("Expected connStr:\n%s\nGot:\n%s", tc.expectedConnStr, gotConnStr)
			}
		})
	}
}
