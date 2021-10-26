package network

import (
	"strconv"

	"github.com/sijms/go-ora/trace"
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
	TransportDataUnitSize uint16
	SessionDataUnitSize   uint16
	Protocol              string
	Host                  string
	UserID                string
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
	PrefetchRows int
}

// ConnectionData build connection data string from connection parameters stored in
// connection option object
func (op *ConnectionOption) ConnectionData() string {
	if len(op.connData) > 0 {
		return op.connData
	}
	FulCid := "(CID=(PROGRAM=" + op.ClientData.ProgramPath + ")(HOST=" + op.ClientData.HostName + ")(USER=" + op.ClientData.UserName + "))"
	address := "(ADDRESS=(PROTOCOL=" + op.Protocol + ")(HOST=" + op.Host + ")(PORT=" + strconv.Itoa(op.Port) + "))"
	result := "(CONNECT_DATA="
	if op.SID != "" {
		result += "(SID=" + op.SID + ")"
	} else {
		result += "(SERVICE_NAME=" + op.ServiceName + ")"
	}
	//if op.ServiceName != "" {
	//
	//} else {
	//	if op.SID != "" {
	//
	//	}
	//}
	if op.InstanceName != "" {
		result += "(INSTANCE_NAME=" + op.InstanceName + ")"
	}
	result += FulCid
	return "(DESCRIPTION=" + address + result + "))"
}
