package network

import (
	"strconv"
	"strings"

	"github.com/sijms/go-ora/v2/trace"
)

type ClientData struct {
	ProgramPath string
	ProgramName string
	UserName    string
	HostName    string
	DriverName  string
	PID         int
}
type ConnectionOption struct {
	Port                  int
	TransportConnectTo    int
	SSLVersion            string
	WalletDict            string
	TransportDataUnitSize uint32
	SessionDataUnitSize   uint32
	Protocol              string
	Host                  string
	UserID                string
	Servers               []string
	Ports                 []int
	//IP string
	SID string
	//Addr string
	//Server string
	ServiceName  string
	InstanceName string
	DomainName   string
	DBName       string
	ClientData   ClientData
	//InAddrAny bool
	Tracer       trace.Tracer
	connData     string
	SNOConfig    map[string]string
	PrefetchRows int
	SSL          bool
	SSLVerify    bool
}

func (op *ConnectionOption) UpdateServers() {
	for i := 0; i < len(op.Servers); i++ {
		if strings.ToUpper(op.Host) == strings.ToUpper(op.Servers[i]) &&
			op.Port == op.Ports[i] {
			return
		}
	}
	op.Servers = append([]string{op.Host}, op.Servers...)
	op.Ports = append([]int{op.Port}, op.Ports...)
}

func (op *ConnectionOption) ConnectionData() string {
	//if len(op.connData) > 0 {
	//	return op.connData
	//}
	FulCid := "(CID=(PROGRAM=" + op.ClientData.ProgramPath + ")(HOST=" + op.ClientData.HostName + ")(USER=" + op.ClientData.UserName + "))"
	address := "(ADDRESS=(PROTOCOL=" + op.Protocol + ")(HOST=" + op.Host + ")(PORT=" + strconv.Itoa(op.Port) + "))"
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
