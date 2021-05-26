package advanced_nego

import (
	"errors"
	"fmt"
	"github.com/sijms/go-ora/v2/network"
)

type dataIntegrityService struct {
	defaultService
	algoID int
}

func NewDataIntegrityService(connOption *network.ConnectionOption) (*dataIntegrityService, error) {
	output := &dataIntegrityService{
		defaultService: defaultService{
			serviceType:           3,
			version:               0xB200200,
			availableServiceNames: []string{"", "MD5", "SHA1", "SHA512", "SHA256", "SHA384"},
			availableServiceIDs:   []int{0, 1, 3, 4, 5, 6},
		},
	}
	str := ""
	level := ""
	if connOption != nil {
		snConfig := connOption.SNOConfig
		if snConfig != nil {
			var exists bool
			str, exists = snConfig["sqlnet.crypto_checksum_types_client"]
			if !exists {
				str = ""
			}
			level, exists = snConfig["sqlnet.crypto_checksum_client"]
			if !exists {
				level = ""
			}
		}
	}
	output.readAdvNegoLevel(level)
	//level := conops.Encryption != null ? conops.Encryption : snoConfig[];
	err := output.buildServiceList(str, true, true)
	//output.selectedServ, err = output.validate(strings.Split(str,","), true)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (serv *dataIntegrityService) readServiceData(session *network.Session, subPacketNum int) error {
	var err error
	serv.version, err = serv.readVersion(session)
	if err != nil {
		return err
	}
	_, err = serv.readPacketHeader(session, 2)
	if err != nil {
		return err
	}
	resp, err := session.GetByte()
	if err != nil {
		return err
	}
	serv.algoID = int(resp)
	if subPacketNum != 8 {
		return nil
	}
	return errors.New("diffie hellman key exchange still under development")
	//dhGroupGLen, err := session.GetInt(2, false, true)
	//if err != nil {
	//	return err
	//}
	//dhGroupPLen, err := session.GetInt(2, false, true)
	//if err != nil {
	//	return err
	//}
	//raw1, err := serv.readBytes(session)
	//if err != nil {
	//	return err
	//}
	//raw2, err := serv.readBytes(session)
	//if err != nil {
	//	return err
	//}
	//raw3, err := serv.readBytes(session)
	//if err != nil {
	//	return err
	//}
	//raw4, err := serv.readBytes(session)
	//if err != nil {
	//	return err
	//}
	//if dhGroupGLen <= 0 || dhGroupPLen <= 0 {
	//	return errors.New("advanced negotiation error: bad parameter from server")
	//}
	//byteLen := (dhGroupPLen + 7) / 8
	//if len(raw3) != byteLen || len(raw2) != byteLen {
	//	return errors.New("advanced negotiation error: DiffieHellman negotiation out of sync")
	//}
}
func (serv *dataIntegrityService) writeServiceData(session *network.Session) error {
	serv.writeHeader(session, 2)
	err := serv.writeVersion(session)
	if err != nil {
		return err
	}
	err = serv.writePacketHeader(session, len(serv.selectedIndices), 1)
	if err != nil {
		return err
	}
	for i := 0; i < len(serv.selectedIndices); i++ {
		index := serv.selectedIndices[i]
		session.PutBytes(uint8(serv.availableServiceIDs[index]))
	}
	return nil
}

func (serv *dataIntegrityService) getServiceDataLength() int {
	return 12 + len(serv.selectedIndices)
}

func (serv *dataIntegrityService) activateAlgorithm() error {
	if serv.algoID == 0 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("advanced negotiation error: data integrity service algorithm: %d still not supported", serv.algoID))

		switch serv.availableServiceNames[serv.algoID] {
		case "MD5":
		case "SHA1":
		case "SHA512":
		case "SHA256":
		case "SHA384":
		}
		return nil
		// you can use also IDs
	}
}
