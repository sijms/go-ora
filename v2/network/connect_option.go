package network

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sijms/go-ora/v2/trace"
)

type ServerAddr struct {
	Protocol string
	Addr     string
	Port     int
}
type ClientInfo struct {
	ProgramPath string
	ProgramName string
	UserName    string
	Password    string
	HostName    string
	DomainName  string
	DriverName  string
	PID         int
	UseKerberos bool
	EnableOOB   bool
	Language    string
	Territory   string
	CharsetID   int
	Cid         string
}
type DatabaseInfo struct {
	UserID          string
	Servers         []ServerAddr
	serverIndex     int
	SID             string
	ProxyClientName string
	ServiceName     string
	InstanceName    string
	DBName          string
	AuthType        int
	connStr         string
}

type DialerContext interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type SessionInfo struct {
	SSLVersion            string
	Timeout               time.Duration
	UnixAddress           string
	TransportDataUnitSize uint32
	SessionDataUnitSize   uint32
	Protocol              string
	SSL                   bool
	SSLVerify             bool
	Dialer                DialerContext
}
type AdvNegoServiceInfo struct {
	AuthService     []string
	EncServiceLevel int
	IntServiceLevel int
}
type ConnectionOption struct {
	ClientInfo
	DatabaseInfo
	SessionInfo
	AdvNegoServiceInfo
	Tracer       trace.Tracer
	TraceDir     string
	PrefetchRows int
	Lob          int
	//Failover     int
	//RetryTime    int

}

func extractServers(connStr string) ([]ServerAddr, error) {
	r, err := regexp.Compile(`(?i)\(\s*ADDRESS\s*=\s*(\(\s*(HOST)\s*=\s*([\w.-]+)\s*\)|\(\s*(PORT)\s*=\s*([0-9]+)\s*\)|\(\s*(COMMUNITY)\s*=\s*([\w.-]+)\s*\)|\(\s*(PROTOCOL)\s*=\s*(\w+)\s*\)\s*)+\)`)
	if err != nil {
		return nil, err
	}
	ret := make([]ServerAddr, 0, 5)
	matches := r.FindAllStringSubmatch(connStr, -1)
	for _, match := range matches {
		server := ServerAddr{
			Port: 1521,
		}
		for x := 2; x < len(match); x++ {
			if strings.ToUpper(match[x]) == "PROTOCOL" {
				x++
				server.Protocol = match[x]
				continue
			}
			if strings.ToUpper(match[x]) == "PORT" {
				x++
				server.Port, err = strconv.Atoi(match[x])
				if err != nil {
					return nil, err
				}
				continue
			}
			if strings.ToUpper(match[x]) == "HOST" {
				x++
				server.Addr = match[x]
			}
		}
		if len(server.Addr) > 0 {
			ret = append(ret, server)
		}
	}
	return ret, nil
}

func (si *SessionInfo) updateSSL(server *ServerAddr) error {
	if server != nil {
		if strings.ToLower(server.Protocol) == "tcps" {
			si.SSL = true
			return nil
		} else if strings.ToLower(server.Protocol) == "tcp" {
			si.SSL = false
			return nil
		}
	}
	if strings.ToLower(si.Protocol) == "tcp" {
		si.SSL = false
	} else if strings.ToLower(si.Protocol) == "tcps" {
		si.SSL = true
	} else {
		return fmt.Errorf("unknown or missing protocol: %s", si.Protocol)
	}
	return nil
}

func (info *DatabaseInfo) UpdateDatabaseInfo(connStr string) error {
	connStr = strings.ReplaceAll(connStr, "\r", "")
	connStr = strings.ReplaceAll(connStr, "\n", "")
	info.connStr = connStr
	var err error
	info.Servers, err = extractServers(connStr)
	if err != nil {
		return err
	}
	if len(info.Servers) == 0 {
		return errors.New("no address passed in connection string")
	}
	r, err := regexp.Compile(`(?i)\(\s*SERVICE_NAME\s*=\s*([\w.\-]+)\s*\)`)
	if err != nil {
		return err
	}
	match := r.FindStringSubmatch(connStr)
	if len(match) > 1 {
		info.ServiceName = match[1]
	}
	r, err = regexp.Compile(`(?i)\(\s*SID\s*=\s*([\w.\-]+)\s*\)`)
	if err != nil {
		return err
	}
	match = r.FindStringSubmatch(connStr)
	if len(match) > 1 {
		info.SID = match[1]
	}
	r, err = regexp.Compile(`(?i)\(\s*INSTANCE_NAME\s*=\s*([\w.\-]+)\s*\)`)
	if err != nil {
		return err
	}
	match = r.FindStringSubmatch(connStr)
	if len(match) > 1 {
		info.InstanceName = match[1]
	}
	return nil
}
func (info *DatabaseInfo) AddServer(server ServerAddr) {
	for i := 0; i < len(info.Servers); i++ {
		if server.IsEqual(&info.Servers[i]) {
			return
		}
	}
	info.Servers = append(info.Servers, server)
}
func (serv *ServerAddr) IsEqual(input *ServerAddr) bool {
	return strings.ToUpper(serv.Addr) == strings.ToUpper(input.Addr) &&
		serv.Port == input.Port
}
func (serv *ServerAddr) networkAddr() string {
	return net.JoinHostPort(serv.Addr, strconv.Itoa(serv.Port))
}
func (info *DatabaseInfo) ResetServerIndex() {
	info.serverIndex = 0
}
func (info *DatabaseInfo) GetActiveServer(jump bool) *ServerAddr {
	if jump {
		info.serverIndex++
	}
	if info.serverIndex >= len(info.Servers) {
		return nil
	}
	return &info.Servers[info.serverIndex]
}

func (op *ConnectionOption) ConnectionData() string {
	if len(op.connStr) != 0 {
		return op.connStr
	}
	host := op.GetActiveServer(false)
	protocol := op.Protocol
	if host.Protocol != "" {
		protocol = host.Protocol
	}
	FulCid := "(CID=(PROGRAM=" + op.ProgramPath + ")(HOST=" + op.HostName + ")(USER=" + op.UserName + "))"
	if len(op.Cid) > 0 {
		FulCid = op.Cid
	}
	var address string
	if len(op.UnixAddress) > 0 {
		address = "(ADDRESS=(PROTOCOL=IPC)(KEY=EXTPROC1))"
	} else {
		address = "(ADDRESS=(PROTOCOL=" + protocol + ")(HOST=" + host.Addr + ")(PORT=" + strconv.Itoa(host.Port) + "))"
	}

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
