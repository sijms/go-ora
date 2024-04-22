package configurations

import (
	"errors"
	"net"
	"regexp"
	"strconv"
	"strings"
)

const defaultPort int = 1521

type DBAPrivilege int

const (
	NONE    DBAPrivilege = 0
	SYSDBA  DBAPrivilege = 0x20
	SYSOPER DBAPrivilege = 0x40
)

type AuthType int

const (
	Normal   AuthType = 0
	OS       AuthType = 1
	Kerberos AuthType = 2
	TCPS     AuthType = 3
)

type ServerAddr struct {
	Protocol string
	Addr     string
	Port     int
}

type DatabaseInfo struct {
	UserID          string
	Password        string
	Servers         []ServerAddr
	serverIndex     int
	SID             string
	ProxyClientName string
	ServiceName     string
	InstanceName    string
	DBName          string
	DBAPrivilege    DBAPrivilege
	AuthType        AuthType
	Wallet          *Wallet
	connStr         string
}

func ExtractServers(connStr string) ([]ServerAddr, error) {
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

func (info *DatabaseInfo) UpdateDatabaseInfo(connStr string) error {
	connStr = strings.ReplaceAll(connStr, "\r", "")
	connStr = strings.ReplaceAll(connStr, "\n", "")
	info.connStr = connStr
	var err error
	info.Servers, err = ExtractServers(connStr)
	if err != nil {
		return err
	}
	if len(info.Servers) == 0 {
		return errors.New("no address passed in connection string")
	}
	r, err := regexp.Compile(`(?i)\(\s*SERVICE_NAME\s*=\s*([\w.-]+)\s*\)`)
	if err != nil {
		return err
	}
	match := r.FindStringSubmatch(connStr)
	if len(match) > 1 {
		info.ServiceName = match[1]
	}
	r, err = regexp.Compile(`(?i)\(\s*SID\s*=\s*([\w.-]+)\s*\)`)
	if err != nil {
		return err
	}
	match = r.FindStringSubmatch(connStr)
	if len(match) > 1 {
		info.SID = match[1]
	}
	r, err = regexp.Compile(`(?i)\(\s*INSTANCE_NAME\s*=\s*([\w.-]+)\s*\)`)
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
func (serv *ServerAddr) NetworkAddr() string {
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

func DBAPrivilegeFromString(s string) DBAPrivilege {
	S := strings.ToUpper(s)
	if S == "SYSDBA" {
		return SYSDBA
	} else if S == "SYSOPER" {
		return SYSOPER
	} else {
		return NONE
	}
}
