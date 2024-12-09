package configurations

import (
	"os"
	"os/user"
)

type ClientInfo struct {
	ProgramPath string
	ProgramName string
	OSUserName  string
	OSPassword  string
	HostName    string
	DomainName  string
	DriverName  string
	PID         int
	UseKerberos bool
	Language    string
	Territory   string
	CharsetID   int
	Cid         string
}

func getCurrentUser() *user.User {
	if userName := os.Getenv("USER"); len(userName) > 0 {
		return &user.User{
			Uid:      "",
			Gid:      "",
			Username: userName,
			Name:     userName,
			HomeDir:  "",
		}
	} else {
		temp, _ := user.Current()
		return temp
	}
}
