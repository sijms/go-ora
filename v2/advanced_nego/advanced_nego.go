package advanced_nego

import (
	"errors"
	"github.com/sijms/go-ora/v2/network"
)

type AdvNego struct {
	serviceList []AdvNegoService
}

func NewAdvNego(connOption *network.ConnectionOption) (*AdvNego, error) {
	output := &AdvNego{
		serviceList: make([]AdvNegoService, 5),
	}
	var err error
	output.serviceList[1], err = NewAuthService(connOption)
	if err != nil {
		return nil, err
	}
	output.serviceList[2], err = NewEncryptService(connOption)
	if err != nil {
		return nil, err
	}
	output.serviceList[3], err = NewDataIntegrityService(connOption)
	if err != nil {
		return nil, err
	}
	output.serviceList[4], err = NewSupervisorService()
	if err != nil {
		return nil, err
	}
	return output, nil
}
func (nego *AdvNego) readHeader(session *network.Session) ([]int, error) {
	num, err := session.GetInt64(4, false, true)
	if err != nil {
		return nil, err
	}
	if num != 0xDEADBEEF {
		return nil, errors.New("advanced negotiation error: during receive header")
	}
	output := make([]int, 4)
	output[0], err = session.GetInt(2, false, true)
	if err != nil {
		return nil, err
	}
	output[1], err = session.GetInt(4, false, true)
	if err != nil {
		return nil, err
	}
	output[2], err = session.GetInt(2, false, true)
	if err != nil {
		return nil, err
	}
	output[3], err = session.GetInt(1, false, true)
	return output, err
}
func (nego *AdvNego) readServiceHeader(session *network.Session) ([]int, error) {
	output := make([]int, 3)
	var err error
	output[0], err = session.GetInt(2, false, true)
	if err != nil {
		return nil, err
	}
	output[1], err = session.GetInt(2, false, true)
	if err != nil {
		return nil, err
	}
	output[2], err = session.GetInt(4, false, true)
	return output, err
}
func (nego *AdvNego) Read(session *network.Session) error {
	header, err := nego.readHeader(session)
	if err != nil {
		return err
	}
	for i := 0; i < header[2]; i++ {
		serviceHeader, err := nego.readServiceHeader(session)
		if err != nil {
			return err
		}
		if serviceHeader[2] != 0 {
			return errors.New("advanced negotiation error: during receive service header")
		}
		err = nego.serviceList[serviceHeader[0]].readServiceData(session, serviceHeader[1])
		if err != nil {
			return err
		}
		err = nego.serviceList[serviceHeader[0]].validateResponse()
		if err != nil {
			return err
		}
	}
	for i := 1; i < 5; i++ {
		err = nego.serviceList[i].activateAlgorithm()
		if err != nil {
			return err
		}
	}
	if authServ, ok := nego.serviceList[1].(*authService); ok {
		if authServ.active {
			errors.New("advanced negotiation: advanced authentication still not supported")
			if authServ.serviceName == "KERBEROS5" {

			} else if authServ.serviceName == "NTS" {

			}
		}
	}
	return nil
}
func (nego *AdvNego) Write(session *network.Session) error {
	session.ResetBuffer()
	size := 0
	for i := 1; i < 5; i++ {
		size = size + 8 + nego.serviceList[i].getServiceDataLength()
	}
	size += 13
	session.PutInt(uint64(0xDEADBEEF), 4, true, false)
	session.PutInt(size, 2, true, false)
	session.PutInt(nego.serviceList[1].getVersion(), 4, true, false)
	session.PutInt(4, 2, true, false)
	session.PutBytes(0)
	err := nego.serviceList[4].writeServiceData(session)
	if err != nil {
		return err
	}
	err = nego.serviceList[1].writeServiceData(session)
	if err != nil {
		return err
	}
	err = nego.serviceList[2].writeServiceData(session)
	if err != nil {
		return err
	}
	err = nego.serviceList[3].writeServiceData(session)
	if err != nil {
		return err
	}
	return session.Write()
}
