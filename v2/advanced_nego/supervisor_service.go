package advanced_nego

import (
	"errors"
	"github.com/sijms/go-ora/v2/network"
)

type supervisorService struct {
	defaultService
	cid       []byte
	servArray []int
}

func NewSupervisorService() (*supervisorService, error) {
	output := &supervisorService{
		defaultService: defaultService{
			serviceType: 4,
			version:     0xB200200,
		},
		cid:       []byte{0, 0, 16, 28, 102, 236, 40, 234},
		servArray: []int{4, 1, 2, 3},
	}
	return output, nil
}

func (serv *supervisorService) readServiceData(session *network.Session, subPacketNum int) error {
	var err error
	_, err = serv.readVersion(session)
	if err != nil {
		return err
	}
	_, err = serv.readPacketHeader(session, 6)
	if err != nil {
		return err
	}
	status, err := session.GetInt(2, false, true)
	if err != nil {
		return err
	}
	if status != 31 {
		return errors.New("advanced negotiation error: reading supervisor service")
	}

	_, err = serv.readPacketHeader(session, 1)
	if err != nil {
		return err
	}
	num1, err := session.GetInt64(4, false, true)
	if err != nil {
		return err
	}
	num2, err := session.GetInt(2, false, true)
	if err != nil {
		return err
	}
	size, err := session.GetInt(4, false, true)
	if err != nil {
		return err
	}
	if num1 != 0xDEADBEEF || num2 != 3 {
		return errors.New("advanced negotiation error: reading supervisor service")
	}
	serv.servArray = make([]int, size)
	for i := 0; i < size; i++ {
		serv.servArray[i], err = session.GetInt(2, false, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (serv *supervisorService) writeServiceData(session *network.Session) error {
	serv.writeHeader(session, 3)
	err := serv.writeVersion(session)
	if err != nil {
		return err
	}
	// send cid
	err = serv.writePacketHeader(session, len(serv.cid), 1)
	if err != nil {
		return err
	}
	session.PutBytes(serv.cid...)

	// send the serv-array
	err = serv.writePacketHeader(session, 10+len(serv.servArray)*2, 1)
	if err != nil {
		return err
	}
	session.PutInt(uint64(0xDEADBEEF), 4, true, false)
	session.PutInt(3, 2, true, false)
	session.PutInt(len(serv.servArray), 4, true, false)
	for i := 0; i < len(serv.servArray); i++ {
		session.PutInt(serv.servArray[i], 2, true, false)
	}
	return nil
}

func (serv *supervisorService) getServiceDataLength() int {
	return 12 + len(serv.cid) + 4 + 10 + (len(serv.servArray) * 2)
}
