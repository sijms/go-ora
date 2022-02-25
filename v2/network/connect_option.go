package network

import (
	"strconv"
	"strings"
	"time"

	"github.com/sijms/go-ora/v2/trace"
)

type ClientInfo struct {
	ProgramPath string
	ProgramName string
	UserName    string
	Password    string
	HostName    string
	DomainName  string
	DriverName  string
	PID         int
}
type DatabaseInfo struct {
	UserID       string
	Servers      []string
	Ports        []int
	serverIndex  int
	SID          string
	ServiceName  string
	InstanceName string
	DBName       string
	AuthType     int
}
type SessionInfo struct {
	SSLVersion string
	Timeout    time.Duration
	//WalletDict            string
	TransportDataUnitSize uint32
	SessionDataUnitSize   uint32
	Protocol              string
	SSL                   bool
	SSLVerify             bool
}
type AdvNegoSeviceInfo struct {
	AuthService []string
}
type ConnectionOption struct {
	//Port                  int
	//TransportConnectTo    int

	//Host                  string

	//IP string

	//Addr string
	//Server string

	ClientInfo
	DatabaseInfo
	SessionInfo
	AdvNegoSeviceInfo
	//InAddrAny bool
	Tracer trace.Tracer
	//connData     string
	SNOConfig    map[string]string
	PrefetchRows int
}

func (op *ConnectionOption) AddServer(host string, port int) {
	for i := 0; i < len(op.Servers); i++ {
		if strings.ToUpper(host) == strings.ToUpper(op.Servers[i]) &&
			port == op.Ports[i] {
			return
		}
	}
	op.Servers = append(op.Servers, host)
	op.Ports = append(op.Ports, port)
}

func (op *ConnectionOption) GetActiveServer(jump bool) (string, int) {
	if jump {
		op.serverIndex++
	}
	if op.serverIndex >= len(op.Servers) {
		return "", 0
	}
	return op.Servers[op.serverIndex], op.Ports[op.serverIndex]
}

func (op *ConnectionOption) ConnectionData() string {
	//if len(op.connData) > 0 {
	//	return op.connData
	//}
	host, port := op.GetActiveServer(false)
	FulCid := "(CID=(PROGRAM=" + op.ProgramPath + ")(HOST=" + op.HostName + ")(USER=" + op.UserName + "))"
	address := "(ADDRESS=(PROTOCOL=" + op.Protocol + ")(HOST=" + host + ")(PORT=" + strconv.Itoa(port) + "))"
	result := "(CONNECT_DATA="
	if op.SID != "" {
		result += "(SID=" + op.SID + ")"
	} else {
		result += "(SERVICE_NAME=" + op.ServiceName + ")"
	}
	if op.InstanceName != "" {
		result += "(INSTANCE_NAME=" + op.InstanceName + ")"
	}
	result += FulCid
	return "(DESCRIPTION=" + address + result + "))"
}
