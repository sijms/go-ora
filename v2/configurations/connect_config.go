package configurations

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type LobFetch int

const (
	INLINE LobFetch = 0
	STREAM LobFetch = 1
)

type AdvNegoServiceInfo struct {
	AuthService     []string
	EncServiceLevel int
	IntServiceLevel int
}
type ConnectionConfig struct {
	ClientInfo
	DatabaseInfo
	SessionInfo
	AdvNegoServiceInfo
	//Tracer       trace.Tracer
	TraceFilePath string
	TraceDir      string
	PrefetchRows  int
	Lob           LobFetch
	//Failover     int
	//RetryTime    int

}

func (config *ConnectionConfig) ConnectionData() string {
	if len(config.connStr) != 0 {
		return config.connStr
	}
	host := config.GetActiveServer(false)
	protocol := config.Protocol
	if host.Protocol != "" {
		protocol = host.Protocol
	}
	FulCid := "(CID=(PROGRAM=" + config.ProgramPath + ")(HOST=" + config.HostName + ")(USER=" + config.OSUserName + "))"
	if len(config.Cid) > 0 {
		FulCid = config.Cid
	}
	var address string
	if len(config.UnixAddress) > 0 {
		address = "(ADDRESS=(PROTOCOL=IPC)(KEY=EXTPROC1))"
	} else {
		address = "(ADDRESS=(PROTOCOL=" + protocol + ")(HOST=" + host.Addr + ")(PORT=" + strconv.Itoa(host.Port) + "))"
	}

	result := "(CONNECT_DATA="
	if config.SID != "" {
		result += "(SID=" + config.SID + ")"
	} else {
		result += "(SERVICE_NAME=" + config.ServiceName + ")"
	}
	if config.InstanceName != "" {
		result += "(INSTANCE_NAME=" + config.InstanceName + ")"
	}
	result += FulCid
	return "(DESCRIPTION=" + address + result + "))"
}

func ParseConfig(dsn string) (*ConnectionConfig, error) {
	walletPath := ""
	walletPass := ""
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	config := &ConnectionConfig{
		PrefetchRows: 25,
		SessionInfo: SessionInfo{
			Timeout: time.Second * time.Duration(120),
			//TransportDataUnitSize: 0xFFFF,
			//SessionDataUnitSize:   0xFFFF,
			TransportDataUnitSize: 0x200000,
			SessionDataUnitSize:   0x200000,
			Protocol:              "tcp",
			SSL:                   false,
			SSLVerify:             true,
		},
		DatabaseInfo: DatabaseInfo{
			Servers: make([]ServerAddr, 0, 3),
		},
		ClientInfo: ClientInfo{Territory: "AMERICA", Language: "AMERICAN"},
	}
	//ret := &ConnectionString{
	//Port:         defaultPort,
	//DBAPrivilege: NONE,
	//EnList:                TRUE,
	//IncrPoolSize:          5,
	//DecrPoolSize:          5,
	//MaxPoolSize:           100,
	//MinPoolSize:           1,
	//PromotableTransaction: Promotable,
	//StmtCacheSize:         20,
	//MetadataBooling:       true,
	//SelfTuning:            true,
	//PoolRegulator:         100,
	//ConnectionPoolTimeout: 15,
	//}
	config.UserID = u.User.Username()
	config.DatabaseInfo.Password, _ = u.User.Password()
	if strings.ToUpper(config.UserID) == "SYS" {
		config.DBAPrivilege = SYSDBA
	}

	host, p, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, err
	}
	if len(host) > 0 {
		tempAddr := ServerAddr{Addr: host, Port: defaultPort}
		tempAddr.Port, err = strconv.Atoi(p)
		if err != nil {
			tempAddr.Port = defaultPort
		}
		config.Servers = append(config.Servers, tempAddr)
	}
	config.ServiceName = strings.Trim(u.Path, "/")
	for key, val := range q {
		switch strings.ToUpper(key) {
		case "CID":
			config.Cid = val[0]
		case "CONNSTR":
			err = config.UpdateDatabaseInfo(q.Get("connStr"))
			if err != nil {
				return nil, err
			}
		case "SERVER":
			for _, srv := range val {
				srv = strings.TrimSpace(srv)
				if srv != "" {
					host, p, err := net.SplitHostPort(srv)
					if err != nil {
						return nil, err
					}
					tempAddr := ServerAddr{Addr: host, Port: defaultPort}
					if p != "" {
						tempAddr.Port, err = strconv.Atoi(p)
						if err != nil {
							tempAddr.Port = defaultPort
						}
					}
					config.Servers = append(config.Servers, tempAddr)
				}
			}
		case "SERVICE NAME":
			config.ServiceName = val[0]
		case "SID":
			config.SID = val[0]
		case "INSTANCE NAME":
			config.InstanceName = val[0]
		case "WALLET":
			walletPath = val[0]
		case "WALLET PASSWORD":
			walletPass = val[0]
		case "AUTH TYPE":
			if strings.ToUpper(val[0]) == "OS" {
				config.AuthType = OS
			} else if strings.ToUpper(val[0]) == "KERBEROS" {
				config.AuthType = Kerberos
			} else if strings.ToUpper(val[0]) == "TCPS" {
				config.AuthType = TCPS
			} else {
				config.AuthType = Normal
			}
		case "OS USER":
			config.OSUserName = val[0]
		case "OS PASS":
			fallthrough
		case "OS PASSWORD":
			config.OSPassword = val[0]
		case "OS HASH":
			fallthrough
		case "OS PASSHASH":
			fallthrough
		case "OS PASSWORD HASH":
			config.ClientInfo.OSPassword = val[0]

		case "DOMAIN":
			config.DomainName = val[0]
		case "AUTH SERV":
			for _, tempVal := range val {
				config.AuthService, _ = uniqueAppendString(config.AuthService, strings.ToUpper(strings.TrimSpace(tempVal)), false)
			}
		case "ENCRYPTION":
			switch strings.ToUpper(val[0]) {
			case "ACCEPTED":
				config.EncServiceLevel = 0
			case "REJECTED":
				config.EncServiceLevel = 1
			case "REQUESTED":
				config.EncServiceLevel = 2
			case "REQUIRED":
				config.EncServiceLevel = 3
			default:
				return nil, fmt.Errorf("unknown encryption service level: %s use one of the following [ACCEPTED, REJECTED, REQUESTED, REQUIRED]", val[0])
			}
		case "DATA INTEGRITY":
			switch strings.ToUpper(val[0]) {
			case "ACCEPTED":
				config.IntServiceLevel = 0
			case "REJECTED":
				config.IntServiceLevel = 1
			case "REQUESTED":
				config.IntServiceLevel = 2
			case "REQUIRED":
				config.IntServiceLevel = 3
			default:
				return nil, fmt.Errorf("unknown data integrity service level: %s use one of the following [ACCEPTED, REJECTED, REQUESTED, REQUIRED]", val[0])
			}
		case "SSL":
			config.SSL = strings.ToUpper(val[0]) == "TRUE" ||
				strings.ToUpper(val[0]) == "ENABLE" ||
				strings.ToUpper(val[0]) == "ENABLED"
		case "SSL VERIFY":
			config.SSLVerify = strings.ToUpper(val[0]) == "TRUE" ||
				strings.ToUpper(val[0]) == "ENABLE" ||
				strings.ToUpper(val[0]) == "ENABLED"
		case "DBA PRIVILEGE":
			config.DBAPrivilege = DBAPrivilegeFromString(val[0])
		case "TIMEOUT":
			fallthrough
		case "CONNECT TIMEOUT":
			fallthrough
		case "CONNECTION TIMEOUT":
			to, err := strconv.Atoi(val[0])
			if err != nil {
				return nil, errors.New("CONNECTION TIMEOUT value must be an integer")
			}
			config.SessionInfo.Timeout = time.Second * time.Duration(to)
		case "TRACE FILE":
			config.TraceFilePath = val[0]
			//if len(val[0]) > 0 {
			//	tf, err := os.Create(val[0])
			//	if err != nil {
			//		//noinspection GoErrorStringFormat
			//		return nil, fmt.Errorf("Can't open trace file: %w", err)
			//	}
			//	config.Tracer = trace.NewTraceWriter(tf)
			//} else {
			//	config.Tracer = trace.NilTracer()
			//}
		case "TRACE DIR":
			fallthrough
		case "TRACE FOLDER":
			fallthrough
		case "TRACE DIRECTORY":
			config.TraceDir = val[0]
		case "USE_OOB":
			fallthrough
		case "ENABLE_OOB":
			fallthrough
		case "ENABLE URGENT DATA TRANSPORT":
			config.EnableOOB = true
		case "PREFETCH_ROWS":
			config.PrefetchRows, err = strconv.Atoi(val[0])
			if err != nil {
				config.PrefetchRows = 25
			}
		case "UNIX SOCKET":
			config.SessionInfo.UnixAddress = val[0]
		case "PROXY CLIENT NAME":
			config.DatabaseInfo.ProxyClientName = val[0]
		case "FAILOVER":
			return nil, errors.New("starting from v2.7.0 this feature (FAILOVER) is not supported and the driver use database/sql package fail over")
			//config.Failover, err = strconv.Atoi(val[0])
			//if err != nil {
			//	config.Failover = 0
			//}
		case "RETRYTIME":
			fallthrough
		case "RE-TRY TIME":
			fallthrough
		case "RETRY TIME":
			return nil, errors.New("starting from v2.7.0 this feature (RETRY TIME) is not supported and the driver use database/sql package fail over")
			//config.RetryTime, err = strconv.Atoi(val[0])
			//if err != nil {
			//	config.RetryTime = 0
			//}
		case "LOB FETCH":
			tempVal := strings.ToUpper(val[0])
			if tempVal == "PRE" || tempVal == "INLINE" {
				config.Lob = INLINE
			} else if tempVal == "POST" || tempVal == "STREAM" {
				config.Lob = STREAM
			} else {
				return nil, errors.New("LOB FETCH value should be either INLINE/PRE (default) or STREAM/POST")
			}
		case "LANGUAGE":
			config.Language = val[0]
		case "TERRITORY":
			config.Territory = val[0]
		case "CHARSET":
			fallthrough
		case "CLIENT CHARSET":
			config.CharsetID, err = getCharsetID(val[0])
			if err != nil {
				return nil, err
			}
		case "PROGRAM":
			config.ClientInfo.ProgramName = val[0]
		default:
			return nil, fmt.Errorf("unknown URL option: %s", key)
			//else if tempVal == "IMPLICIT" || tempVal == "AUTO" {
			//	config.Lob = 1
			//} else if tempVal == "EXPLICIT" || tempVal == "MANUAL" {
			//	config.Lob = 2
			//} else {
			//	return nil, errors.New("LOB value should be: Prefetch, Implicit(AUTO) or Explicit(manual)")
			//}
			//case "ENLIST":
			//	ret.EnList = EnListFromString(val[0])
			//case "INC POOL SIZE":
			//	ret.IncrPoolSize, err = strconv.Atoi(val[0])
			//	if err != nil {
			//		return nil, errors.New("INC POOL SIZE value must be an integer")
			//	}
			//case "DECR POOL SIZE":
			//	ret.DecrPoolSize, err = strconv.Atoi(val[0])
			//	if err != nil {
			//		return nil, errors.New("DECR POOL SIZE value must be an integer")
			//	}
			//case "MAX POOL SIZE":
			//	ret.MaxPoolSize, err = strconv.Atoi(val[0])
			//	if err != nil {
			//		return nil, errors.New("MAX POOL SIZE value must be an integer")
			//	}
			//case "MIN POOL SIZE":
			//	ret.MinPoolSize, err = strconv.Atoi(val[0])
			//	if err != nil {
			//		return nil, errors.New("MIN POOL SIZE value must be an integer")
			//	}
			//case "POOL REGULATOR":
			//	ret.PoolRegulator, err = strconv.Atoi(val[0])
			//	if err != nil {
			//		return nil, errors.New("POOL REGULATOR value must be an integer")
			//	}
			//case "STATEMENT CACHE SIZE":
			//	ret.StmtCacheSize, err = strconv.Atoi(val[0])
			//	if err != nil {
			//		return nil, errors.New("STATEMENT CACHE SIZE value must be an integer")
			//	}
			//case "CONNECTION POOL TIMEOUT":
			//	ret.ConnectionPoolTimeout, err = strconv.Atoi(val[0])
			//	if err != nil {
			//		return nil, errors.New("CONNECTION POOL TIMEOUT value must be an integer")
			//	}
			//case "CONNECTION LIFETIME":
			//	ret.ConnectionLifeTime, err = strconv.Atoi(val[0])
			//	if err != nil {
			//		return nil, errors.New("CONNECTION LIFETIME value must be an integer")
			//	}
			//case "PERSIST SECURITY INFO":
			//	ret.PasswordSecurityInfo = val[0] == "TRUE"
			//case "POOLING":
			//	ret.Pooling = val[0] == "TRUE"
			//case "VALIDATE CONNECTION":
			//	ret.ValidateConnection = val[0] == "TRUE"
			//case "STATEMENT CACHE PURGE":
			//	ret.StmtCachePurge = val[0] == "TRUE"
			//case "HA EVENTS":
			//	ret.HaEvent = val[0] == "TRUE"
			//case "LOAD BALANCING":
			//	ret.LoadBalance = val[0] == "TRUE"
			//case "METADATA POOLING":
			//	ret.MetadataBooling = val[0] == "TRUE"
			//case "SELF TUNING":
			//	ret.SelfTuning = val[0] == "TRUE"
			//case "CONTEXT CONNECTION":
			//	ret.ContextConnection = val[0] == "TRUE"
			//case "PROMOTABLE TRANSACTION":
			//	if val[0] == "PROMOTABLE" {
			//		ret.PromotableTransaction = Promotable
			//	} else {
			//		ret.PromotableTransaction = Local
			//	}
			//case "APPLICATION EDITION":
			//	ret.ApplicationEdition = val[0]
			//case "PROXY USER ID":
			//	ret.ProxyUserID = val[0]
			//case "PROXY PASSWORD":
			//	ret.ProxyPassword = val[0]
		}
	}
	if len(config.Servers) == 0 {
		return nil, errors.New("empty connection servers")
	}
	if len(walletPath) > 0 {
		if len(config.ServiceName) == 0 {
			return nil, errors.New("you should specify server/service if you will use wallet")
		}
		if _, err = os.Stat(path.Join(walletPath, "ewallet.p12")); err == nil && len(walletPass) > 0 {
			fileData, err := os.ReadFile(path.Join(walletPath, "ewallet.p12"))
			if err != nil {
				return nil, err
			}
			config.Wallet = &Wallet{password: []byte(walletPass)}
			err = config.Wallet.readPKCS12(fileData)
			if err != nil {
				return nil, err
			}
		} else {
			config.Wallet, err = NewWallet(path.Join(walletPath, "cwallet.sso"))
			if err != nil {
				return nil, err
			}
		}

		if len(config.UserID) > 0 {
			if len(config.Password) == 0 {
				serv := config.Servers[0]
				cred, err := config.Wallet.getCredential(serv.Addr, serv.Port, config.ServiceName, config.UserID)
				if err != nil {
					return nil, err
				}
				if cred == nil {
					return nil, errors.New(
						fmt.Sprintf("cannot find credentials for server: %s:%d, service: %s,  username: %s",
							serv.Addr, serv.Port, config.ServiceName, config.UserID))
				}
				config.UserID = cred.username
				config.Password = cred.password
			}
		}
	}
	return config, config.validate()
}

// validate check is data in connection configuration is correct and fulfilled
func (config *ConnectionConfig) validate() error {
	if len(config.SID) == 0 && len(config.ServiceName) == 0 {
		return errors.New("empty SID and service name")
	}
	if config.AuthType == Kerberos {
		config.AuthService = append(config.AuthService, "KERBEROS5")
	}
	if config.AuthType == TCPS {
		config.AuthService = append(config.AuthService, "TCPS")
	}
	if len(config.UserID) == 0 || len(config.Password) == 0 && config.AuthType == Normal {
		config.AuthType = OS
	}
	if config.AuthType == OS {
		if runtime.GOOS == "windows" {
			config.AuthService = append(config.AuthService, "NTS")
		}
	}

	if config.SSL {
		config.Protocol = "tcps"
	}

	// get client info
	var idx int
	var temp = getCurrentUser()

	if temp != nil {
		idx = strings.Index(temp.Username, "\\")
		if idx >= 0 {
			if len(config.DomainName) == 0 {
				config.DomainName = temp.Username[:idx]
			}
			if len(config.ClientInfo.OSUserName) == 0 {
				config.ClientInfo.OSUserName = temp.Username[idx+1:]
			}
		} else {
			if len(config.ClientInfo.OSUserName) == 0 {
				config.ClientInfo.OSUserName = temp.Username
			}
		}
	}
	config.HostName, _ = os.Hostname()

	config.ProgramPath = os.Args[0]
	if config.ProgramName == "" {
		config.ProgramName = filepath.Base(os.Args[0])
	}
	config.DriverName = "OracleClientGo"
	config.PID = os.Getpid()
	return nil
}
