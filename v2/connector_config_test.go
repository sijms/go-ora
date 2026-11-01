package go_ora

import (
	"database/sql/driver"
	"testing"

	"github.com/sijms/go-ora/v2/configurations"
)

// NewConnectorWithConfig keeps the configuration on the connector, so two
// connectors in one process can use different credentials. RegisterConnConfig
// cannot do this, since it assigns to the package level driver.
func TestNewConnectorWithConfig(t *testing.T) {
	first := &configurations.ConnectionConfig{}
	first.UserID = "first"
	second := &configurations.ConnectionConfig{}
	second.UserID = "second"

	c1, ok := NewConnectorWithConfig(first).(*OracleConnector)
	if !ok {
		t.Fatal("NewConnectorWithConfig did not return an *OracleConnector")
	}
	c2, ok := NewConnectorWithConfig(second).(*OracleConnector)
	if !ok {
		t.Fatal("NewConnectorWithConfig did not return an *OracleConnector")
	}

	if c1.config != first || c2.config != second {
		t.Error("each connector should hold the config it was given")
	}
	if c1.config.UserID != "first" || c2.config.UserID != "second" {
		t.Errorf("configs crossed over: %q and %q", c1.config.UserID, c2.config.UserID)
	}
	// No connection string, so NewConnection uses the config rather than
	// parsing a URL.
	if c1.connectString != "" {
		t.Errorf("connectString = %q, want empty", c1.connectString)
	}
	if c1.drv == nil || c2.drv == nil {
		t.Error("connector should carry a driver, which is what registers custom types")
	}
	if c1.drv == c2.drv {
		t.Error("each connector should get its own driver, matching NewConnector")
	}
}

// A connector built from a connection string must keep working unchanged.
func TestNewConnectorStillUsesConnectionString(t *testing.T) {
	c, ok := NewConnector("oracle://user:pass@host:1521/service").(*OracleConnector)
	if !ok {
		t.Fatal("NewConnector did not return an *OracleConnector")
	}
	if c.config != nil {
		t.Error("NewConnector should not set a config")
	}
	if c.connectString == "" {
		t.Error("NewConnector should keep the connection string")
	}
}

// Both constructors must satisfy driver.Connector.
func TestConnectorsImplementDriverConnector(t *testing.T) {
	var _ driver.Connector = NewConnector("")
	var _ driver.Connector = NewConnectorWithConfig(&configurations.ConnectionConfig{})
}
