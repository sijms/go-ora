package advanced_nego

import (
	"errors"
	"github.com/sijms/go-ora/network"
	"runtime"
)

type authService struct {
	defaultService
	status      int
	serviceName string
	active      bool
}

func NewAuthService(connOption *network.ConnectionOption) (*authService, error) {
	output := &authService{
		defaultService: defaultService{
			serviceType: 1,
			level:       -1,
			version:     0xB200200,
		},
		status: 0xFCFF,
	}
	//var avaAuth []string
	if runtime.GOOS == "windows" {
		output.availableServiceNames = []string{"", "NTS", "KERBEROS5", "TCPS"}
		output.availableServiceIDs = []int{0, 1, 1, 2}
	} else {
		output.availableServiceNames = []string{"TCPS"}
		output.availableServiceIDs = []int{2}
	}
	str := ""
	if connOption != nil {
		snConfig := connOption.SNOConfig
		if snConfig != nil {
			var exists bool
			str, exists = snConfig["sqlnet.authentication_services"]
			if !exists {
				str = ""
			}
		}
	}
	//level := conops.Encryption != null ? conops.Encryption : snoConfig[];
	err := output.buildServiceList(str, false, false)
	//output.selectedServ, err = output.validate(strings.Split(str,","), true)
	if err != nil {
		return nil, err
	}
	return output, nil
	/* user list is found in the dictionary
	sessCtx.m_conops.SNOConfig["sqlnet.authentication_services"]
	*/
	/* you need to confirm that every item in user list found in avaAuth list
	then for each item in userList you need to get index of it in the avaAuth
	return output*/
}

func (serv *authService) writeServiceData(session *network.Session) error {
	serv.writeHeader(session, 3+(len(serv.selectedIndices)*2))
	err := serv.writeVersion(session)
	if err != nil {
		return err
	}
	err = serv.writePacketHeader(session, 2, 3)
	if err != nil {
		return err
	}
	session.PutInt(0xE0E1, 2, true, false)
	err = serv.writePacketHeader(session, 2, 6)
	if err != nil {
		return err
	}
	session.PutInt(serv.status, 2, true, false)
	for i := 0; i < len(serv.selectedIndices); i++ {
		index := serv.selectedIndices[i]
		session.PutBytes(uint8(serv.availableServiceIDs[index]))
		session.PutBytes([]byte(serv.availableServiceNames[index])...)
	}
	return nil
}

func (serv *authService) readServiceData(session *network.Session, subPacketNum int) error {
	// read version
	var err error
	serv.version, err = serv.readVersion(session)
	if err != nil {
		return err
	}
	// read status
	_, err = serv.readPacketHeader(session, 6)
	if err != nil {
		return err
	}
	status, err := session.GetInt(2, false, true)
	if err != nil {
		return err
	}
	if status == 0xFAFF && subPacketNum > 2 {
		// get 1 byte with header
		_, err = serv.readPacketHeader(session, 2)
		if err != nil {
			return err
		}
		_, err = session.GetByte()
		if err != nil {
			return err
		}
		stringLen, err := serv.readPacketHeader(session, 0)
		if err != nil {
			return err
		}
		serviceNameBytes, err := session.GetBytes(stringLen)
		if err != nil {
			return err
		}
		serv.serviceName = string(serviceNameBytes)
		if subPacketNum > 4 {
			_, err = serv.readVersion(session)
			if err != nil {
				return err
			}
			_, err = serv.readPacketHeader(session, 4)
			if err != nil {
				return err
			}
			_, err = session.GetInt(4, false, true)
			if err != nil {
				return err
			}
			_, err = serv.readPacketHeader(session, 4)
			if err != nil {
				return err
			}
			_, err = session.GetInt(4, false, true)
			if err != nil {
				return err
			}
		}
		serv.active = true
	} else {
		if status != 0xFBFF {
			return errors.New("advanced negotiation error: reading authentication service")
		}
		serv.active = false
	}
	return nil
}

func (serv *authService) getServiceDataLength() int {
	size := 20
	for i := 0; i < len(serv.selectedIndices); i++ {
		index := serv.selectedIndices[i]
		size = size + 5 + (4 + len(serv.availableServiceNames[index]))
	}
	return size
}
