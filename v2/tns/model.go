package tns

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// TNS represents the top-level structure of a TNS (Transparent Network Substrate) configuration file.
//
// Docs: https://docs.oracle.com/en/database/oracle/oracle-database/21/netrf/local-naming-parameters-in-tns-ora-file.html
type TNS struct {
	Entries []TNSMappingEntry `yaml:"entries"`
}

// TNSMappingEntry is a mapping of network service names to connect descriptor.
type TNSMappingEntry struct {
	// NetServiceName is an alias mapped to network database address stored inside ConnectDescriptor.
	NetServiceName string `yaml:"net_service_name"`
	// ConnectDescriptor
	ConnectDescriptor ConnectDescriptor `yaml:"connect_descriptor"`
}

// ConnectDescriptor represents a network connectivity information piece.
//
// In Oracle TNS configuration, you use `DESCRIPTION` and `DESCRIPTION_LIST` to define
// network connectivity information for Oracle databases. The decision to use either depends
// on the complexity and requirements of the network configuration.
//
//   - A single `DESCRIPTION` is used when you have a straightforward network configuration,
//     where you need to define a single connection endpoint for the Oracle database. It includes
//     details such as the protocol, host, and port.
//
//     Example of a single `DESCRIPTION`:
//
//     ORCLDB =
//     (DESCRIPTION =
//     (ADDRESS = (PROTOCOL = TCP)(HOST = myhost)(PORT = 1521))
//     (CONNECT_DATA =
//     (SERVER = DEDICATED)
//     (SERVICE_NAME = orcl)
//     )
//     )
//
//   - `DESCRIPTION_LIST` is used for more complex network configurations, such as load balancing,
//     failover, or when you need to define multiple connection endpoints for redundancy or distribution.
//     Each `DESCRIPTION` in the `DESCRIPTION_LIST` can specify separate connection details.
//
//     Example of `DESCRIPTION_LIST`:
//
//     ORCLDB =
//     (DESCRIPTION_LIST =
//     (LOAD_BALANCE = ON)
//     (FAILOVER = ON)
//     (DESCRIPTION =
//     (ADDRESS = (PROTOCOL = TCP)(HOST = myhost1)(PORT = 1521))
//     (CONNECT_DATA =
//     (SERVER = DEDICATED)
//     (SERVICE_NAME = orcl1)
//     )
//     )
//     (DESCRIPTION =
//     (ADDRESS = (PROTOCOL = TCP)(HOST = myhost2)(PORT = 1522))
//     (CONNECT_DATA =
//     (SERVER = DEDICATED)
//     (SERVICE_NAME = orcl2)
//     )
//     )
//     )
//
//     Note: `LOAD_BALANCE` and `FAILOVER` options can be used to distribute connections and provide redundancy.
//
//   - Multiple `ADDRESS` entries within a single `DESCRIPTION` are used to define multiple potential endpoints
//     for a single logical connection configuration. This supports connection failover within the same configuration context.
//
//     Example of multiple `ADDRESS` entries:
//
//     ORCLDB =
//     (DESCRIPTION =
//     (ADDRESS_LIST =
//     (LOAD_BALANCE = OFF)
//     (FAILOVER = ON)
//     (ADDRESS = (PROTOCOL = TCP)(HOST = myhost1)(PORT = 1521))
//     (ADDRESS = (PROTOCOL = TCP)(HOST = myhost2)(PORT = 1521))
//     )
//     (CONNECT_DATA =
//     (SERVER = DEDICATED)
//     (SERVICE_NAME = orcl)
//     )
//     )
//
//   - This configuration primarily supports failover, though it can also be used for manual load balancing.
//
//   - Both addresses point to the same logical SERVICE_NAME, indicating they service the same database.
//
// Differences:
//
// - **Multiple `DESCRIPTION` Entries (in `DESCRIPTION_LIST`)**
//   - Support both load balancing (`LOAD_BALANCE = ON`) and failover (`FAILOVER = ON`).
//   - Each `DESCRIPTION` can be a completely different configuration, potentially pointing to different database instances.
//   - Ideal for complex configurations requiring high availability and distribution among different services.
//
// - **Multiple `ADDRESS` Entries (in One `DESCRIPTION`)**
//   - Primarily support failover, though load balancing can be manually configured.
//   - All addresses share the same logical `CONNECT_DATA`, meaning they refer to the same database instance/service.
//   - Ideal for redundancy and high availability connecting to the same database service through multiple network paths.
//
// TNS provides a way to represent both configurations within a Go struct:
// - `Description` is used for the single `DESCRIPTION` configuration.
// - `DescriptionList` is used for the `DESCRIPTION_LIST` configuration to handle multiple `DESCRIPTION` entries.
type ConnectDescriptor struct {
	// The DESCRIPTION_LIST parameter of the tnsnames.ora file defines a list of connect descriptors for a particular net service name.
	//
	// https://docs.oracle.com/en/database/oracle/oracle-database/21/netrf/local-naming-parameters-in-tns-ora-file.html#GUID-C5C15BAA-A493-49DB-8D62-ECE520EADD8F
	DescriptionList DescriptionList `yaml:"description_list"`

	// The DESCRIPTION parameter of the tnsnames.ora file defines a connect descriptor for a particular net service name.
	//
	// https://docs.oracle.com/en/database/oracle/oracle-database/21/netrf/local-naming-parameters-in-tns-ora-file.html#GUID-6540A1EC-93B2-43BD-B462-1277B82EAE25
	Description Description `yaml:"description"` // Represents a single DESCRIPTION configuration.
}

type DescriptionList struct {
	Descriptions []Description `yaml:"description"`
}

type Description struct {
	// It's really messy how Oracle handle addresses in TNS files.
	//
	// According to the docs, ADDRESS_LIST is not required, and was introduced in Oracle 8.0. Before that, ADDRESS was used
	// directly under DESCRIPTION.
	//
	// It looks like, that should be merged... But the tricky detail is that AddressList has additional properties,
	// that should be applied to that addresses, but not to the addresses under DESCRIPTION.
	AddressList  `yaml:",inline"` // FIXME: An embedded AddressList structure, without ADDRESS_LIST prefix
	AddressLists []AddressList    `yaml:"address_list"` // A list of lists of addresses

	RetryCount              int         `yaml:"retry_count,omitempty"`               // The number of times to retry a connection, e.g., 3
	TransportConnectTimeout int         `yaml:"transport_connect_timeout,omitempty"` // The number of seconds to wait for a transport connection to be established, e.g., 30
	ConnectTimeout          int         `yaml:"connect_timeout,omitempty"`           // The number of seconds to wait for a connection to be established, e.g., 30
	ConnectData             ConnectData `yaml:"connect_data,omitempty"`              // A ConnectData structure containing the service name, e.g., {"service_name": "sales.us.example.com"}
	Security                Security    `yaml:"security,omitempty"`                  // A Security structure containing the SSL server certificate DN, e.g., {"ssl_server_cert_dn": "CN=server.example.com,OU=Example,O=Example,L=Example,ST=Example,C=US"}
	EncryptionClient        string      `yaml:"encryption_client,omitempty"`         // Added encryption client
	AuthenticationServices  string      `yaml:"authentication_services,omitempty"`   // Added authentication services
	Compression             string      `yaml:"compression,omitempty"`               // Enable or disable data compression, e.g., "on"
	CompressionLevels       []string    `yaml:"compression_levels,omitempty"`        // Specify the compression levels, e.g., ["low", "high"]
}

// AddressList represents a list of addresses and source routing information.
type AddressList struct {
	Address       []Address `yaml:"address"`         // A list of Address structures, e.g., [{"host": "host1.example.com", "port": 1521, "protocol": "tcp"}]
	Enable        string    `yaml:"enable"`          // The keepalive feature, e.g., "broken"
	Failover      string    `yaml:"failover"`        // Indicates if failover is enabled (reoriented relevant property)
	LoadBalance   string    `yaml:"load_balance"`    // Indicates if load balancing is enabled (reoriented relevant property)
	RecvBufSize   int       `yaml:"recv_buf_size"`   // The buffer space for receive operations, e.g., 11784
	Sdu           int       `yaml:"sdu"`             // The session data unit size, e.g., 8192
	SendBufSize   int       `yaml:"send_buf_size"`   // The buffer space for send operations, e.g., 11784
	SourceRoute   string    `yaml:"source_route"`    // Indicates whether source routing is enabled (yes or no), e.g., "yes"
	TypeOfService string    `yaml:"type_of_service"` // The type of service for an Oracle Rdb database, e.g., "rdb_database"
}

type Address struct {
	Host     string `yaml:"host"`     // The hostname or IP address of the server, e.g., "host1.example.com"
	Port     int    `yaml:"port"`     // The port number on which the server is listening, e.g., 1521
	Protocol string `yaml:"protocol"` // The protocol used for communication, e.g., "tcp"

	//  The HTTPS proxy server, e.g., "proxy.example.com:80"
	//
	// https://docs.oracle.com/en/database/oracle/oracle-database/21/netrf/local-naming-parameters-in-tns-ora-file.html#GUID-C672E92D-CE32-4759-9931-92D7960850F7
	HttpsProxyHost string `yaml:"https_proxy"`

	// The HTTPS proxy port, e.g., 80
	//
	// https://docs.oracle.com/en/database/oracle/oracle-database/21/netrf/local-naming-parameters-in-tns-ora-file.html#GUID-C672E92D-CE32-4759-9931-92D7960850F7
	HttpsProxyPort int `yaml:"https_proxy_port"`

	SendBufSize int `yaml:"send_buf_size"` // The buffer space for send operations of sessions, e.g., 11784
	RecvBufSize int `yaml:"recv_buf_size"` // The buffer space for receive operations of sessions, e.g., 11784
}

type ConnectData struct {
	// ColocationTag is used to specify a colocation tag for the connection.
	//
	// https://docs.oracle.com/en/database/oracle/oracle-database/21/netrf/local-naming-parameters-in-tns-ora-file.html#GUID-9D06528E-BBC4-4629-A119-C216EBF70201
	ColocationTag string `yaml:"colocation_tag"`

	// ConnectionIdPrefix is used to specify a prefix for the connection ID.
	//
	// https://docs.oracle.com/en/database/oracle/oracle-database/21/netrf/local-naming-parameters-in-tns-ora-file.html#GUID-9C46B9AF-4582-46DC-8D21-5F5F686983BA
	ConnectionIdPrefix string `yaml:"connection_id_prefix"`

	// PoolConnectionClass is used to specify the connection class for the connection pool.
	//
	// https://docs.oracle.com/en/database/oracle/oracle-database/21/netrf/local-naming-parameters-in-tns-ora-file.html#GUID-FC3A42EA-2809-4B7E-A969-E962B8C25906
	PoolConnectionClass string `yaml:"pool_connection_class"`

	// PoolPurity is used to specify the purity of the connection pool.
	//
	// https://docs.oracle.com/en/database/oracle/oracle-database/21/netrf/local-naming-parameters-in-tns-ora-file.html#GUID-640A7C66-CF4C-47C4-A725-F6E1097C460D
	PoolPurity string `yaml:"pool_purity"`

	// ShardingKey is used to specify the sharding key for the connection.
	//
	// https://docs.oracle.com/en/database/oracle/oracle-database/21/netrf/local-naming-parameters-in-tns-ora-file.html#GUID-9254F571-77DA-4345-9B20-9460DBF43B3D
	ShardingKey string `yaml:"sharding_key"`

	// SuperShardingKey is used to specify the super sharding key for the connection.
	//
	// https://docs.oracle.com/en/database/oracle/oracle-database/21/netrf/local-naming-parameters-in-tns-ora-file.html#GUID-8AA96248-466D-423D-A516-235B41A3C24A
	SuperShardingKey string `yaml:"super_sharding_key"`

	// TunnelServiceName is used to specify the tunnel service name for the connection.
	//
	// https://docs.oracle.com/en/database/oracle/oracle-database/21/netrf/local-naming-parameters-in-tns-ora-file.html#GUID-7DF419C2-3237-4B79-9AA3-EA1C2D786712
	TunnelServiceName string `yaml:"tunnel_service_name"`

	FailoverMode string     `yaml:"failover_mode"` // The failover mode, e.g., "BASIC"
	GlobalName   string     `yaml:"global_name"`   // The global name of the Oracle Rdb database, e.g., "alpha5"
	Hs           string     `yaml:"hs"`            // Indicates connection to a non-Oracle system through Heterogeneous Services, e.g., "ok"
	InstanceName string     `yaml:"instance_name"` // The database instance to access, e.g., "sales1"
	RdbDatabase  string     `yaml:"rdb_database"`  // The file name of an Oracle Rdb database, e.g., "[.mf]mf_personal.rdb"
	Server       ServerType `yaml:"server"`        // The type of server connection, e.g., "dedicated", "shared", or "pooled"
	ServiceName  string     `yaml:"service_name"`  // The name of the service to connect to, e.g., "sales.us.example.com"
	Sid          string     `yaml:"sid"`           // The Oracle System Identifier (SID), e.g., "sales6"
}

type Security struct {
	SslServerCertDN       string `yaml:"ssl_server_cert_dn"`     // The distinguished name of the server's SSL certificate, e.g., "CN=server.example.com,OU=Example,O=Example,L=Example,ST=Example,C=US"
	AuthenticationService string `yaml:"authentication_service"` // The authentication service to use, e.g., "KERBEROS5"
	IgnoreAnoEncryption   bool   `yaml:"ignore_ano_encryption"`  // Whether to ignore ANO encryption for TCPS, e.g., true
	Kerberos5CCName       string `yaml:"kerberos5_cc_name"`      // The complete path name to the Kerberos credentials cache (CC) file, e.g., "/tmp/kcache"
	SslServerDNMatch      bool   `yaml:"ssl_server_dn_match"`    // Whether to enforce server-side certificate validation through DN matching, e.g., true
	SslVersion            string `yaml:"ssl_version"`            // The version of TLS to use, e.g., "1.2"
	WalletLocation        string `yaml:"wallet_location"`        // The location where Oracle wallets are stored, e.g., "/home/oracle/wallets/databases"
}

// ServerType represents the type of server connection.
type ServerType string

const (
	// Dedicated server type.
	Dedicated ServerType = "dedicated"
	// Shared server type.
	Shared ServerType = "shared"
	// Pooled server type.
	Pooled ServerType = "pooled"
)

var (
	_ yaml.Marshaler   = (*ServerType)(nil)
	_ yaml.Unmarshaler = (*ServerType)(nil)
)

func (j *ServerType) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(string(*j))
}

func (j *ServerType) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	t := ServerType(strings.ToLower(s))
	switch t {
	case Dedicated, Shared, Pooled:
		*j = t
		return nil
	default:
		return fmt.Errorf("invalid server type: %s", t)
	}
}

// GetDescriptorForService returns a connect descriptor for the specified service.
//
// Returns nil if there's no connect descriptor for the specified service.
func (t *TNS) GetDescriptorForService(name string) *ConnectDescriptor {
	for _, cd := range t.Entries {
		if cd.NetServiceName == name {
			desc := cd.ConnectDescriptor
			return &desc
		}
	}
	return nil
}
